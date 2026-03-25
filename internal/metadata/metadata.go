package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// MetadataKey is the key used to store the metadata JSON in the Secret.
const MetadataKey = "metadata.json"

// Metadata holds the installer metadata for a cluster, used for generating
// configuration that maps vCenter resources to OpenShift infrastructure.
type Metadata struct {
	// ClusterName is the human-readable cluster name.
	ClusterName string `json:"clusterName"`
	// ClusterID is the unique cluster identifier.
	ClusterID string `json:"clusterID"`
	// InfraID is the infrastructure identifier used to name cloud resources.
	InfraID string `json:"infraID"`
	// VCenter is the primary vCenter server address.
	VCenter string `json:"vcenter"`
	// Username is the primary vCenter username.
	Username string `json:"username"`
	// Password is the primary vCenter password.
	Password string `json:"password"`
	// TerraformPlatform identifies the Terraform provider platform.
	TerraformPlatform string `json:"terraformPlatform"`
	// VCenters is the list of all vCenter configurations.
	VCenters []VCenters `json:"vcenters"`
}

// VCenters holds the connection and topology details for a single vCenter.
type VCenters struct {
	// Server is the vCenter FQDN or IP address.
	Server string `json:"server"`
	// Port is the vCenter connection port.
	Port int32 `json:"port"`
	// Username is the vCenter login username.
	Username string `json:"username"`
	// Password is the vCenter login password.
	Password string `json:"password"`
	// Datacenters is the list of datacenter names.
	Datacenters []string `json:"datacenters"`
	// DefaultDC is the default datacenter.
	DefaultDC string `json:"defaultDC"`
	// Cluster is the compute cluster path.
	Cluster string `json:"cluster"`
	// Datastore is the datastore path.
	Datastore string `json:"datastore"`
	// Network is the network port group path.
	Network string `json:"network"`
	// ResourcePool is the resource pool path.
	ResourcePool string `json:"resourcePool"`
	// Folder is the VM folder path.
	Folder string `json:"folder"`
}

// MetadataManager generates and persists installer metadata.
type MetadataManager struct {
	kubeClient kubernetes.Interface
}

// NewMetadataManager creates a new MetadataManager with the given Kubernetes client.
func NewMetadataManager(kubeClient kubernetes.Interface) *MetadataManager {
	return &MetadataManager{kubeClient: kubeClient}
}

// GenerateMetadata builds a Metadata object from the given failure domains,
// Infrastructure resource, and credentials. The credentials map is keyed by
// vCenter server and the values are formatted as "username:password".
// Each failure domain contributes a VCenters entry.
func (m *MetadataManager) GenerateMetadata(ctx context.Context, failureDomains []configv1.VSpherePlatformFailureDomainSpec, infra *configv1.Infrastructure, credentials map[string]string) (*Metadata, error) {
	log := klog.FromContext(ctx)
	log.V(2).Info("generating installer metadata")

	if infra == nil {
		return nil, fmt.Errorf("infrastructure must not be nil")
	}
	if infra.Status.InfrastructureName == "" {
		return nil, fmt.Errorf("infrastructure status has no InfrastructureName")
	}

	md := &Metadata{
		ClusterName:       infra.Name,
		ClusterID:         string(infra.UID),
		InfraID:           infra.Status.InfrastructureName,
		TerraformPlatform: "vsphere",
	}

	// Build vCenter entries from failure domains.
	for i := range failureDomains {
		fd := &failureDomains[i]

		username, password := parseCredentials(credentials[fd.Server])

		if i == 0 {
			md.VCenter = fd.Server
			md.Username = username
			md.Password = password
		}

		var network string
		if len(fd.Topology.Networks) > 0 {
			network = fd.Topology.Networks[0]
		}

		vc := VCenters{
			Server:       fd.Server,
			Port:         443,
			Username:     username,
			Password:     password,
			Datacenters:  []string{fd.Topology.Datacenter},
			DefaultDC:    fd.Topology.Datacenter,
			Cluster:      fd.Topology.ComputeCluster,
			Datastore:    fd.Topology.Datastore,
			Network:      network,
			ResourcePool: fd.Topology.ResourcePool,
			Folder:       fd.Topology.Folder,
		}
		md.VCenters = append(md.VCenters, vc)
	}

	log.V(2).Info("generated metadata", "infraID", md.InfraID, "vcenterCount", len(md.VCenters))
	return md, nil
}

// SaveToSecret persists the metadata as JSON in a Secret with the given
// namespace and name. If the Secret already exists, it is updated.
// A Secret is used instead of a ConfigMap because the metadata contains
// vCenter credentials.
func (m *MetadataManager) SaveToSecret(ctx context.Context, md *Metadata, namespace, name string) error {
	log := klog.FromContext(ctx)
	log.V(2).Info("saving metadata to Secret", "namespace", namespace, "name", name)

	data, err := json.Marshal(md)
	if err != nil {
		return fmt.Errorf("marshalling metadata: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"migration.openshift.io/metadata": "true",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			MetadataKey: data,
		},
	}

	existing, err := m.kubeClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		existing.Data = secret.Data
		if _, err := m.kubeClient.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("updating metadata Secret %s/%s: %w", namespace, name, err)
		}
	} else {
		if _, err := m.kubeClient.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("creating metadata Secret %s/%s: %w", namespace, name, err)
		}
	}

	log.V(2).Info("saved metadata to Secret", "namespace", namespace, "name", name)
	return nil
}

// GetMetadataFromSecret reads the metadata JSON from a Secret.
func (m *MetadataManager) GetMetadataFromSecret(ctx context.Context, namespace, name string) ([]byte, error) {
	secret, err := m.kubeClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting metadata Secret %s/%s: %w", namespace, name, err)
	}
	data, ok := secret.Data[MetadataKey]
	if !ok {
		return nil, fmt.Errorf("key %q not found in Secret %s/%s", MetadataKey, namespace, name)
	}
	return data, nil
}

// GetMetadataSecretName returns the conventional Secret name for storing
// migration metadata based on the migration resource name.
func GetMetadataSecretName(migrationName string) string {
	return fmt.Sprintf("%s-metadata", migrationName)
}

// parseCredentials splits a "username:password" string into its components.
// If the format is unexpected, both values are returned empty.
func parseCredentials(cred string) (username, password string) {
	if cred == "" {
		return "", ""
	}
	parts := splitFirst(cred, ':')
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// splitFirst splits s on the first occurrence of sep.
func splitFirst(s string, sep byte) []string {
	idx := -1
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			idx = i
			break
		}
	}
	if idx < 0 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+1:]}
}
