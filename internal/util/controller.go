package util

import (
	"context"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/openshift"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

// PermissionReader returns apiReader when configured, falling back to the
// standard controller-runtime client (used in unit tests with a non-caching fake).
func PermissionReader(apiReader client.Reader, fallback client.Client) client.Reader {
	if apiReader != nil {
		return apiReader
	}
	return fallback
}

// FindFailureDomainByName returns the failure domain with the given name, or
// nil when absent.
func FindFailureDomainByName(fds []configv1.VSpherePlatformFailureDomainSpec, name string) *configv1.VSpherePlatformFailureDomainSpec {
	for i := range fds {
		if fds[i].Name == name {
			return &fds[i]
		}
	}
	return nil
}

// GetVSphereSession creates a vSphere session for the given server and datacenter
// using the provided credentials.
func GetVSphereSession(ctx context.Context, server, datacenter, username, password string) (*vsphere.Session, error) {
	return vsphere.GetOrCreate(ctx, vsphere.Params{
		Server:     server,
		Datacenter: datacenter,
		Username:   username,
		Password:   password,
		Insecure:   true,
	})
}

// GetTargetCredentials resolves the username and password for a target vCenter server
// from the migration's target credentials secret. The secret is expected to have keys
// in the format {server}.username and {server}.password.
func GetTargetCredentials(ctx context.Context, kubeClient kubernetes.Interface, migration *migrationv1alpha1.VmwareCloudFoundationMigration, server string) (username, password string, err error) {
	log := klog.FromContext(ctx)
	log.V(2).Info("resolving target credentials", "server", server)

	secretRef := migration.Spec.TargetVCenterCredentialsSecret
	if secretRef.Name == "" {
		return "", "", fmt.Errorf("spec.targetVCenterCredentialsSecret.name must not be empty")
	}
	ns := secretRef.Namespace
	if ns == "" {
		ns = migration.Namespace
	}

	sm := openshift.NewSecretManager(kubeClient)
	username, password, err = sm.GetVCenterCredsFromSecret(ctx, ns, secretRef.Name, server)
	if err != nil {
		return "", "", fmt.Errorf("getting target credentials for %s from secret %s/%s: %w", server, ns, secretRef.Name, err)
	}

	return username, password, nil
}
