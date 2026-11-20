package storage

import (
	"context"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	configfake "github.com/openshift/client-go/config/clientset/versioned/fake"
	"github.com/vmware/govmomi/find"
	_ "github.com/vmware/govmomi/pbm/simulator"
	"github.com/vmware/govmomi/simulator"
	_ "github.com/vmware/govmomi/vapi/simulator"
	"github.com/vmware/govmomi/vim25"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	fakekube "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/events"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/storageplan"
)

func storageFixture(t *testing.T) (*StorageController, *migrationv1alpha1.VmwareCloudFoundationMigration) {
	t.Helper()
	m := &migrationv1alpha1.VmwareCloudFoundationMigration{
		ObjectMeta: metav1.ObjectMeta{
			Name:       migrationv1alpha1.SingletonName,
			Namespace:  migrationv1alpha1.SingletonNamespace,
			Finalizers: []string{migrationv1alpha1.FinalizerName},
			UID:        "12345678-migration",
			Generation: 1,
		},
		Spec: migrationv1alpha1.VmwareCloudFoundationMigrationSpec{
			State: migrationv1alpha1.MigrationStatePlanning,
			TargetVCenterCredentialsSecret: migrationv1alpha1.SecretReference{
				Name: "target-creds",
			},
			MigrateStorage: true,
			Storage:        &migrationv1alpha1.StorageMigrationSpec{},
			FailureDomains: []configv1.VSpherePlatformFailureDomainSpec{
				{
					Name:   "target",
					Region: "new-region",
					Zone:   "new-zone",
					Server: "target.example.com",
					Topology: configv1.VSpherePlatformTopology{
						Datacenter:     "dc",
						ComputeCluster: "/dc/host/cluster",
						Datastore:      "/dc/datastore/ds",
						Networks:       []string{"network"},
						ResourcePool:   "/dc/host/cluster/Resources",
						Template:       "/dc/vm/template",
					},
				},
			},
		},
	}
	scheme := runtime.NewScheme()
	if err := migrationv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := clientfake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(m).
		WithStatusSubresource(m, &migrationv1alpha1.StorageWorkloadMigration{}).
		Build()
	kubeClient := fakekube.NewClientset(
		&corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{Name: "pv", UID: "pv-uid"},
			Spec: corev1.PersistentVolumeSpec{
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{
						Driver:       storageplan.Driver,
						VolumeHandle: "disk",
					},
				},
				ClaimRef: &corev1.ObjectReference{
					Namespace: "app",
					Name:      "data",
					UID:       "pvc-uid",
				},
			},
			Status: corev1.PersistentVolumeStatus{Phase: corev1.VolumeBound},
		},
		&corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: "data", Namespace: "app", UID: "pvc-uid"},
			Spec:       corev1.PersistentVolumeClaimSpec{VolumeName: "pv", StorageClassName: ptr.To("thin-csi")},
		},
		&storagev1.StorageClass{
			ObjectMeta:        metav1.ObjectMeta{Name: "thin-csi"},
			VolumeBindingMode: ptr.To(storagev1.VolumeBindingWaitForFirstConsumer),
			Parameters:        map[string]string{"storagepolicyname": "openshift-storage-policy-infra"},
		},
	)
	configClient := configfake.NewClientset(&configv1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Spec: configv1.InfrastructureSpec{
			PlatformSpec: configv1.PlatformSpec{
				VSphere: &configv1.VSpherePlatformSpec{
					FailureDomains: []configv1.VSpherePlatformFailureDomainSpec{
						{Name: "source", Region: "old-region", Zone: "old-zone"},
					},
				},
			},
		},
		Status: configv1.InfrastructureStatus{InfrastructureName: "infra"},
	})
	dynClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.openshift.io/v1",
			"kind":       "ClusterCSIDriver",
			"metadata":   map[string]interface{}{"name": storageplan.Driver},
			"spec":       map[string]interface{}{"managementState": "Managed"},
		},
	})
	return NewStorageController(c, nil, kubeClient, configClient, dynClient, events.NewFakeRecorder(100)), m
}

func TestCheckStorageManagement(t *testing.T) {
	for _, state := range []string{"Managed", "Removed", "Unmanaged"} {
		t.Run(state, func(t *testing.T) {
			s, _ := storageFixture(t)
			obj, err := s.DynamicClient.Resource(ClusterCSIDriverGVR).Get(context.Background(), storageplan.Driver, metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if err := unstructured.SetNestedField(obj.Object, state, "spec", "managementState"); err != nil {
				t.Fatal(err)
			}
			if _, err := s.DynamicClient.Resource(ClusterCSIDriverGVR).Update(context.Background(), obj, metav1.UpdateOptions{}); err != nil {
				t.Fatal(err)
			}
			_, err = s.CheckStorageManagement(context.Background(), true /* migrateStorage */)
			if (err == nil) != (state == "Managed") {
				t.Fatalf("CheckStorageManagement(%s): %v", state, err)
			}
		})
	}
}

func TestStaleStoragePlans(t *testing.T) {
	s, m := storageFixture(t)
	if _, err := s.EnsureStoragePlanned(context.Background(), m); err != nil {
		t.Fatalf("EnsureStoragePlanned: %v", err)
	}
	swm := &migrationv1alpha1.StorageWorkloadMigration{}
	key := client.ObjectKey{Namespace: "app", Name: PlanName(m)}
	if err := s.Client.Get(context.Background(), key, swm); err != nil {
		t.Fatal(err)
	}
	// Simulate an execution phase that is not yet part of the API.
	swm.Status.Phase = migrationv1alpha1.StorageWorkloadMigrationPhase("PreparingTargets")
	if err := s.Client.Status().Update(context.Background(), swm); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteStaleStoragePlans(context.Background(), m, &storageplan.Plan{}); err != nil {
		t.Fatal(err)
	}
	if err := s.Client.Get(context.Background(), key, swm); err != nil {
		t.Fatalf("advanced SWM deleted: %v", err)
	}
	swm.Status.Phase = migrationv1alpha1.StorageWorkloadMigrationPhasePlanning
	if err := s.Client.Status().Update(context.Background(), swm); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteStaleStoragePlans(context.Background(), m, &storageplan.Plan{}); err != nil {
		t.Fatal(err)
	}
	if err := s.Client.Get(context.Background(), key, swm); err == nil {
		t.Fatal("stale Planning SWM survived")
	}
}

func TestPVCEventsEnqueueOnlyWhileStoragePlanRefreshes(t *testing.T) {
	ctx := context.Background()
	for _, tt := range []struct {
		name    string
		change  func(*migrationv1alpha1.VmwareCloudFoundationMigration)
		enqueue bool
	}{
		{name: "planning with storage", change: func(*migrationv1alpha1.VmwareCloudFoundationMigration) {}, enqueue: true},
		{name: "running with storage", change: func(m *migrationv1alpha1.VmwareCloudFoundationMigration) {
			m.Spec.State = migrationv1alpha1.MigrationStateRunning
		}, enqueue: true},
		{name: "storage disabled", change: func(m *migrationv1alpha1.VmwareCloudFoundationMigration) {
			m.Spec.MigrateStorage = false
			m.Spec.Storage = nil
		}},
		{name: "paused", change: func(m *migrationv1alpha1.VmwareCloudFoundationMigration) {
			m.Spec.State = migrationv1alpha1.MigrationStatePaused
		}},
		{name: "started", change: func(m *migrationv1alpha1.VmwareCloudFoundationMigration) {
			now := metav1.Now()
			m.Status.StartTime = &now
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s, m := storageFixture(t)
			tt.change(m)
			if err := s.Client.Update(ctx, m); err != nil {
				t.Fatal(err)
			}
			// Update returns the stored status, so reapply status changes.
			tt.change(m)
			if err := s.Client.Status().Update(ctx, m); err != nil {
				t.Fatal(err)
			}
			got := s.mapPVCToStoragePlanning(ctx, &metav1.PartialObjectMetadata{})
			if (len(got) == 1) != tt.enqueue {
				t.Fatalf("mapPVCToStoragePlanning = %+v, want enqueue %v", got, tt.enqueue)
			}
		})
	}

	s, m := storageFixture(t)
	if err := s.Client.Delete(ctx, m); err != nil {
		t.Fatal(err)
	}
	// Remove the finalizer so the object is gone rather than terminating.
	latest := &migrationv1alpha1.VmwareCloudFoundationMigration{}
	if err := s.Client.Get(ctx, client.ObjectKeyFromObject(m), latest); err != nil {
		t.Fatal(err)
	}
	latest.Finalizers = nil
	if err := s.Client.Update(ctx, latest); err != nil {
		t.Fatal(err)
	}
	if got := s.mapPVCToStoragePlanning(ctx, &metav1.PartialObjectMetadata{}); len(got) != 0 {
		t.Fatalf("mapPVCToStoragePlanning without a migration = %+v, want none", got)
	}
}

func TestPVCMetadataChanged(t *testing.T) {
	base := &metav1.PartialObjectMetadata{ObjectMeta: metav1.ObjectMeta{
		Name: "data", Namespace: "app", ResourceVersion: "1",
		Annotations: map[string]string{"a": "b"},
	}}
	for _, tt := range []struct {
		name    string
		change  func(*metav1.PartialObjectMetadata)
		changed bool
	}{
		{name: "resource version only", change: func(p *metav1.PartialObjectMetadata) { p.ResourceVersion = "2" }},
		{name: "managed fields only", change: func(p *metav1.PartialObjectMetadata) {
			p.ManagedFields = []metav1.ManagedFieldsEntry{{Manager: "kubelet"}}
		}},
		{name: "labels only", change: func(p *metav1.PartialObjectMetadata) { p.Labels = map[string]string{"x": "y"} }},
		{name: "target annotation", change: func(p *metav1.PartialObjectMetadata) {
			p.Annotations[storageplan.TargetFailureDomainAnnotation] = "override"
		}, changed: true},
		{name: "owner references", change: func(p *metav1.PartialObjectMetadata) {
			p.OwnerReferences = []metav1.OwnerReference{{Kind: "Pod", Name: "pod"}}
		}, changed: true},
		{name: "deletion", change: func(p *metav1.PartialObjectMetadata) {
			now := metav1.Now()
			p.DeletionTimestamp = &now
		}, changed: true},
		{name: "generation", change: func(p *metav1.PartialObjectMetadata) { p.Generation = 2 }, changed: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			updated := base.DeepCopy()
			tt.change(updated)
			if got := isPVCMetadataChanged(base, updated); got != tt.changed {
				t.Fatalf("isPVCMetadataChanged = %v, want %v", got, tt.changed)
			}
		})
	}
}

func TestCustomerStoragePolicyPlanning(t *testing.T) {
	simulator.Test(func(ctx context.Context, c *vim25.Client) {
		s, m := storageFixture(t)
		finder := find.NewFinder(c, true)
		dc, err := finder.DefaultDatacenter(ctx)
		if err != nil {
			t.Fatal(err)
		}
		finder.SetDatacenter(dc)
		ds, err := finder.DefaultDatastore(ctx)
		if err != nil {
			t.Fatal(err)
		}
		fd := &m.Spec.FailureDomains[0]
		fd.Server = c.URL().Host
		fd.Topology.Datacenter = dc.Name()
		fd.Topology.Datastore = ds.InventoryPath
		password, _ := simulator.DefaultLogin.Password()
		_, err = s.KubeClient.CoreV1().Secrets(m.Namespace).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: m.Spec.TargetVCenterCredentialsSecret.Name},
			Data: map[string][]byte{
				fd.Server + ".username": []byte(simulator.DefaultLogin.Username()),
				fd.Server + ".password": []byte(password),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatal(err)
		}
		sc, err := s.KubeClient.StorageV1().StorageClasses().Get(ctx, "thin-csi", metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for _, tt := range []struct {
			policy  string
			blocked bool
		}{{"vSAN Default Storage Policy", false}, {"missing-policy", true}} {
			sc.Parameters["storagepolicyname"] = tt.policy
			if _, err := s.KubeClient.StorageV1().StorageClasses().Update(ctx, sc, metav1.UpdateOptions{}); err != nil {
				t.Fatal(err)
			}
			_, err := s.EnsureStoragePlanned(ctx, m)
			if err != nil {
				t.Fatalf("EnsureStoragePlanned: %v", err)
			}
			condition := apimeta.FindStatusCondition(m.Status.Conditions, migrationv1alpha1.ConditionStoragePlanned)
			if (condition.Status == metav1.ConditionFalse) != tt.blocked {
				t.Fatalf("policy %s: %+v", tt.policy, condition)
			}
		}
	})
}
