package openshift

import (
	"context"
	"encoding/json"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	// InfrastructureName is the well-known name of the cluster Infrastructure resource.
	InfrastructureName = "cluster"

	// infrastructureCRDName is the CRD name for the Infrastructure resource.
	infrastructureCRDName = "infrastructures.config.openshift.io"
)

// InfrastructureManager manages the OpenShift Infrastructure custom resource.
type InfrastructureManager struct {
	client              configclient.Interface
	apiextensionsClient apiextensionsclient.Interface
}

// NewInfrastructureManager creates a new InfrastructureManager with the given clients.
func NewInfrastructureManager(client configclient.Interface, apiextensionsClient apiextensionsclient.Interface) *InfrastructureManager {
	return &InfrastructureManager{
		client:              client,
		apiextensionsClient: apiextensionsClient,
	}
}

// Get retrieves the cluster Infrastructure resource.
func (m *InfrastructureManager) Get(ctx context.Context) (*configv1.Infrastructure, error) {
	infra, err := m.client.ConfigV1().Infrastructures().Get(ctx, InfrastructureName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting infrastructure %q: %w", InfrastructureName, err)
	}
	return infra, nil
}

// GetSourceVCenter returns the first VCenter entry from the Infrastructure spec.
// This is typically the source vCenter in a single-vCenter cluster.
func (m *InfrastructureManager) GetSourceVCenter(ctx context.Context) (*configv1.VSpherePlatformVCenterSpec, error) {
	infra, err := m.Get(ctx)
	if err != nil {
		return nil, err
	}

	if infra.Spec.PlatformSpec.VSphere == nil || len(infra.Spec.PlatformSpec.VSphere.VCenters) == 0 {
		return nil, fmt.Errorf("no vCenters configured in infrastructure spec")
	}

	return &infra.Spec.PlatformSpec.VSphere.VCenters[0], nil
}

// GetSourceFailureDomain returns the first failure domain from the Infrastructure spec.
func (m *InfrastructureManager) GetSourceFailureDomain(ctx context.Context) (*configv1.VSpherePlatformFailureDomainSpec, error) {
	infra, err := m.Get(ctx)
	if err != nil {
		return nil, err
	}

	if infra.Spec.PlatformSpec.VSphere == nil || len(infra.Spec.PlatformSpec.VSphere.FailureDomains) == 0 {
		return nil, fmt.Errorf("no failure domains configured in infrastructure spec")
	}

	return &infra.Spec.PlatformSpec.VSphere.FailureDomains[0], nil
}

// AddTargetVCenter adds target vCenter entries and failure domains to the Infrastructure
// resource. VCenter entries are derived from the failure domains and are only added if
// not already present. The updated Infrastructure is persisted and returned.
func (m *InfrastructureManager) AddTargetVCenter(ctx context.Context, infra *configv1.Infrastructure, failureDomains []configv1.VSpherePlatformFailureDomainSpec) (*configv1.Infrastructure, error) {
	log := klog.FromContext(ctx)
	log.V(2).Info("adding target vCenter to infrastructure")

	if infra.Spec.PlatformSpec.VSphere == nil {
		return nil, fmt.Errorf("infrastructure spec has no vSphere platform configuration")
	}

	// Build set of existing vCenter servers for deduplication.
	existingServers := make(map[string]bool)
	for _, vc := range infra.Spec.PlatformSpec.VSphere.VCenters {
		existingServers[vc.Server] = true
	}

	// Build vCenter entries from failure domains.
	serverDatacenters := make(map[string]map[string]bool)
	for i := range failureDomains {
		fd := &failureDomains[i]
		if _, ok := serverDatacenters[fd.Server]; !ok {
			serverDatacenters[fd.Server] = make(map[string]bool)
		}
		serverDatacenters[fd.Server][fd.Topology.Datacenter] = true
	}

	for server, dcSet := range serverDatacenters {
		if existingServers[server] {
			log.V(2).Info("vCenter already exists in infrastructure, skipping", "server", server)
			continue
		}

		var datacenters []string
		for dc := range dcSet {
			datacenters = append(datacenters, dc)
		}

		infra.Spec.PlatformSpec.VSphere.VCenters = append(infra.Spec.PlatformSpec.VSphere.VCenters, configv1.VSpherePlatformVCenterSpec{
			Server:      server,
			Port:        443,
			Datacenters: datacenters,
		})
		log.V(2).Info("added vCenter entry", "server", server, "datacenters", datacenters)
	}

	// Add failure domains.
	existingFDs := make(map[string]bool)
	for _, fd := range infra.Spec.PlatformSpec.VSphere.FailureDomains {
		existingFDs[fd.Name] = true
	}
	for i := range failureDomains {
		if existingFDs[failureDomains[i].Name] {
			log.V(2).Info("failure domain already exists, skipping", "name", failureDomains[i].Name)
			continue
		}
		infra.Spec.PlatformSpec.VSphere.FailureDomains = append(infra.Spec.PlatformSpec.VSphere.FailureDomains, failureDomains[i])
		log.V(2).Info("added failure domain", "name", failureDomains[i].Name)
	}

	updated, err := m.client.ConfigV1().Infrastructures().Update(ctx, infra, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("updating infrastructure with target vCenter: %w", err)
	}

	return updated, nil
}

// RemoveSourceVCenter removes the vCenter and associated failure domains for the
// given source server from the Infrastructure resource. The updated Infrastructure
// is persisted and returned.
func (m *InfrastructureManager) RemoveSourceVCenter(ctx context.Context, infra *configv1.Infrastructure, sourceServer string) (*configv1.Infrastructure, error) {
	log := klog.FromContext(ctx)
	log.V(2).Info("removing source vCenter from infrastructure", "server", sourceServer)

	if infra.Spec.PlatformSpec.VSphere == nil {
		return nil, fmt.Errorf("infrastructure spec has no vSphere platform configuration")
	}

	// Remove the vCenter entry.
	var vcenters []configv1.VSpherePlatformVCenterSpec
	for _, vc := range infra.Spec.PlatformSpec.VSphere.VCenters {
		if vc.Server != sourceServer {
			vcenters = append(vcenters, vc)
		}
	}
	infra.Spec.PlatformSpec.VSphere.VCenters = vcenters

	// Remove failure domains that reference the source server.
	var fds []configv1.VSpherePlatformFailureDomainSpec
	for _, fd := range infra.Spec.PlatformSpec.VSphere.FailureDomains {
		if fd.Server != sourceServer {
			fds = append(fds, fd)
		}
	}
	infra.Spec.PlatformSpec.VSphere.FailureDomains = fds

	updated, err := m.client.ConfigV1().Infrastructures().Update(ctx, infra, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("updating infrastructure after removing source vCenter: %w", err)
	}

	log.V(2).Info("removed source vCenter from infrastructure", "server", sourceServer)
	return updated, nil
}

// RemoveSourceVCenterWithCRDModification temporarily modifies the Infrastructure CRD
// to allow vCenter changes (same as when adding), removes the source vCenter and
// its failure domains, then restores the CRD. Use this when the platform
// validation forbids adding/removing vCenters once set.
func (m *InfrastructureManager) RemoveSourceVCenterWithCRDModification(ctx context.Context, infra *configv1.Infrastructure, sourceServer string) (*configv1.Infrastructure, error) {
	log := klog.FromContext(ctx)

	originalCRD, err := m.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, infrastructureCRDName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting infrastructure CRD for backup: %w", err)
	}
	originalSchemas := make([]json.RawMessage, len(originalCRD.Spec.Versions))
	for i := range originalCRD.Spec.Versions {
		if originalCRD.Spec.Versions[i].Schema != nil {
			data, err := json.Marshal(originalCRD.Spec.Versions[i].Schema)
			if err != nil {
				return nil, fmt.Errorf("marshalling original CRD schema for version %s: %w", originalCRD.Spec.Versions[i].Name, err)
			}
			originalSchemas[i] = data
		}
	}

	if err := m.ModifyInfrastructureCRDToAllowVCenterChanges(ctx); err != nil {
		return nil, fmt.Errorf("modifying infrastructure CRD: %w", err)
	}

	updated, updateErr := m.RemoveSourceVCenter(ctx, infra, sourceServer)

	restoreCRD, err := m.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, infrastructureCRDName, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "failed to get infrastructure CRD for schema restoration")
	} else {
		for i := range restoreCRD.Spec.Versions {
			if i < len(originalSchemas) && originalSchemas[i] != nil {
				schema := &apiextensionsv1.CustomResourceValidation{}
				if err := json.Unmarshal(originalSchemas[i], schema); err != nil {
					log.Error(err, "failed to unmarshal original schema", "version", restoreCRD.Spec.Versions[i].Name)
					continue
				}
				restoreCRD.Spec.Versions[i].Schema = schema
			}
		}
		if _, err := m.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, restoreCRD, metav1.UpdateOptions{}); err != nil {
			log.Error(err, "failed to restore infrastructure CRD schema")
		} else {
			log.V(2).Info("restored infrastructure CRD schema")
		}
	}

	if updateErr != nil {
		return nil, fmt.Errorf("updating infrastructure after removing source vCenter (CRD restored): %w", updateErr)
	}
	return updated, nil
}

// GetInfrastructureID returns the infrastructure ID (InfrastructureName) from the
// Infrastructure status.
func (m *InfrastructureManager) GetInfrastructureID(ctx context.Context) (string, error) {
	infra, err := m.Get(ctx)
	if err != nil {
		return "", err
	}

	if infra.Status.InfrastructureName == "" {
		return "", fmt.Errorf("infrastructure status has no InfrastructureName")
	}

	return infra.Status.InfrastructureName, nil
}

// ModifyInfrastructureCRDToAllowVCenterChanges removes x-kubernetes-validations
// from the vcenters field in the Infrastructure CRD to allow adding or removing
// vCenter entries.
func (m *InfrastructureManager) ModifyInfrastructureCRDToAllowVCenterChanges(ctx context.Context) error {
	log := klog.FromContext(ctx)
	log.V(2).Info("modifying infrastructure CRD to allow vCenter changes")

	crd, err := m.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, infrastructureCRDName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting infrastructure CRD: %w", err)
	}

	modified := false
	for vi := range crd.Spec.Versions {
		version := &crd.Spec.Versions[vi]
		if version.Schema == nil || version.Schema.OpenAPIV3Schema == nil {
			continue
		}

		if err := removeVCentersValidations(version.Schema.OpenAPIV3Schema, &modified); err != nil {
			return fmt.Errorf("removing vcenters validations from CRD version %s: %w", version.Name, err)
		}
	}

	if !modified {
		log.V(2).Info("no vcenters validations found to remove")
		return nil
	}

	if _, err := m.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("updating infrastructure CRD: %w", err)
	}

	log.V(2).Info("modified infrastructure CRD to allow vCenter changes")
	return nil
}

// AddTargetVCenterWithCRDModification temporarily modifies the Infrastructure CRD to
// allow vCenter changes, adds the target vCenter and failure domains, then restores
// the CRD schema. This is a convenience method that wraps ModifyInfrastructureCRDToAllowVCenterChanges
// and AddTargetVCenter.
func (m *InfrastructureManager) AddTargetVCenterWithCRDModification(ctx context.Context, infra *configv1.Infrastructure, failureDomains []configv1.VSpherePlatformFailureDomainSpec) (*configv1.Infrastructure, error) {
	log := klog.FromContext(ctx)

	// Save the original CRD for restoration.
	originalCRD, err := m.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, infrastructureCRDName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting infrastructure CRD for backup: %w", err)
	}

	// Deep copy the original schema for restoration.
	originalSchemas := make([]json.RawMessage, len(originalCRD.Spec.Versions))
	for i := range originalCRD.Spec.Versions {
		if originalCRD.Spec.Versions[i].Schema != nil {
			data, err := json.Marshal(originalCRD.Spec.Versions[i].Schema)
			if err != nil {
				return nil, fmt.Errorf("marshalling original CRD schema for version %s: %w", originalCRD.Spec.Versions[i].Name, err)
			}
			originalSchemas[i] = data
		}
	}

	// Modify CRD to allow vCenter changes.
	if err := m.ModifyInfrastructureCRDToAllowVCenterChanges(ctx); err != nil {
		return nil, fmt.Errorf("modifying infrastructure CRD: %w", err)
	}

	// Update infrastructure with target vCenter.
	updated, updateErr := m.AddTargetVCenter(ctx, infra, failureDomains)

	// Restore the CRD schema regardless of whether the update succeeded.
	restoreCRD, err := m.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, infrastructureCRDName, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "failed to get infrastructure CRD for schema restoration")
	} else {
		for i := range restoreCRD.Spec.Versions {
			if i < len(originalSchemas) && originalSchemas[i] != nil {
				schema := &apiextensionsv1.CustomResourceValidation{}
				if err := json.Unmarshal(originalSchemas[i], schema); err != nil {
					log.Error(err, "failed to unmarshal original schema", "version", restoreCRD.Spec.Versions[i].Name)
					continue
				}
				restoreCRD.Spec.Versions[i].Schema = schema
			}
		}

		if _, err := m.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, restoreCRD, metav1.UpdateOptions{}); err != nil {
			log.Error(err, "failed to restore infrastructure CRD schema")
		} else {
			log.V(2).Info("restored infrastructure CRD schema")
		}
	}

	if updateErr != nil {
		return nil, fmt.Errorf("updating infrastructure with target vCenter (CRD restored): %w", updateErr)
	}

	return updated, nil
}

// removeVCentersValidations walks the OpenAPI schema to find the vcenters field and
// removes x-kubernetes-validations from it and its parent VSpherePlatformSpec.
func removeVCentersValidations(schema *apiextensionsv1.JSONSchemaProps, modified *bool) error {
	if schema == nil {
		return nil
	}

	// Walk: spec -> platformSpec -> vsphere -> vcenters
	spec, ok := schema.Properties["spec"]
	if !ok {
		return nil
	}

	platformSpec, ok := spec.Properties["platformSpec"]
	if !ok {
		return nil
	}

	vsphere, ok := platformSpec.Properties["vsphere"]
	if !ok {
		return nil
	}

	// Remove x-kubernetes-validations from the vsphere level (VSpherePlatformSpec).
	if len(vsphere.XValidations) > 0 {
		vsphere.XValidations = nil
		*modified = true
	}

	vcenters, ok := vsphere.Properties["vcenters"]
	if !ok {
		return nil
	}

	// Remove x-kubernetes-validations from the vcenters field.
	if len(vcenters.XValidations) > 0 {
		vcenters.XValidations = nil
		*modified = true
	}

	// Write back the modified schema.
	vsphere.Properties["vcenters"] = vcenters
	platformSpec.Properties["vsphere"] = vsphere
	spec.Properties["platformSpec"] = platformSpec
	schema.Properties["spec"] = spec

	return nil
}
