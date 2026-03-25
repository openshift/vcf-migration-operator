/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	machineclient "github.com/openshift/client-go/machine/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/metadata"
	"github.com/openshift/vcf-migration-operator/internal/openshift"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

// VmwareCloudFoundationMigrationReconciler reconciles a VmwareCloudFoundationMigration object.
// It drives a multi-stage migration workflow from one vCenter to another by
// progressing through a series of conditions in order.
type VmwareCloudFoundationMigrationReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	KubeClient          kubernetes.Interface
	ConfigClient        configclient.Interface
	MachineClient       machineclient.Interface
	DynamicClient       dynamic.Interface
	APIExtensionsClient apiextensionsclient.Interface
	Recorder            record.EventRecorder
}

// conditionOrder defines the sequence in which conditions are evaluated.
// The reconciler processes the first condition that is not True and returns.
var conditionOrder = []string{
	migrationv1alpha1.ConditionInfrastructurePrepared,
	migrationv1alpha1.ConditionDestinationInitialized,
	migrationv1alpha1.ConditionMultiSiteConfigured,
	migrationv1alpha1.ConditionWorkloadMigrated,
	migrationv1alpha1.ConditionSourceCleaned,
	migrationv1alpha1.ConditionReady,
}

// +kubebuilder:rbac:groups=migration.openshift.io,resources=vmwarecloudfoundationmigrations,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=migration.openshift.io,resources=vmwarecloudfoundationmigrations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=migration.openshift.io,resources=vmwarecloudfoundationmigrations/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets;configmaps;pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=config.openshift.io,resources=infrastructures;clusteroperators,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=machine.openshift.io,resources=machinesets;machines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=machine.openshift.io,resources=controlplanemachinesets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

// Reconcile drives the migration workflow by checking conditions in order and
// executing the work for the first incomplete condition. It is idempotent and
// safe to restart at any point.
func (r *VmwareCloudFoundationMigrationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := klog.FromContext(ctx)

	migration := &migrationv1alpha1.VmwareCloudFoundationMigration{}
	if err := r.Get(ctx, req.NamespacedName, migration); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("migration resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("getting migration resource: %w", err)
	}

	if migration.Spec.State != migrationv1alpha1.MigrationStateRunning {
		log.V(1).Info("migration not in Running state, skipping", "state", migration.Spec.State)
		return ctrl.Result{}, nil
	}

	// Set start time on first reconcile in Running state.
	if migration.Status.StartTime == nil {
		now := metav1.Now()
		migration.Status.StartTime = &now
		r.Recorder.Event(migration, "Normal", "MigrationStarted", "Migration workflow started")
		if err := r.updateStatus(ctx, migration); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Walk conditions in order; execute work for the first non-True condition.
	type conditionHandler func(context.Context, *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error)
	handlers := map[string]conditionHandler{
		migrationv1alpha1.ConditionInfrastructurePrepared: r.ensureInfrastructurePrepared,
		migrationv1alpha1.ConditionDestinationInitialized: r.ensureDestinationInitialized,
		migrationv1alpha1.ConditionMultiSiteConfigured:    r.ensureMultiSiteConfigured,
		migrationv1alpha1.ConditionWorkloadMigrated:       r.ensureWorkloadMigrated,
		migrationv1alpha1.ConditionSourceCleaned:          r.ensureSourceCleaned,
		migrationv1alpha1.ConditionReady:                  r.ensureReady,
	}

	for _, condType := range conditionOrder {
		if r.isConditionTrue(migration, condType) {
			continue
		}

		handler, ok := handlers[condType]
		if !ok {
			return ctrl.Result{}, fmt.Errorf("no handler for condition %q", condType)
		}

		log.V(1).Info("processing condition", "condition", condType)
		result, err := handler(ctx, migration)
		if err != nil {
			r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonFailed, err.Error())
			r.Recorder.Eventf(migration, "Warning", "ConditionFailed", "Condition %s failed: %v", condType, err)
		}

		// Always persist status after processing a condition.
		if statusErr := r.updateStatus(ctx, migration); statusErr != nil {
			log.Error(statusErr, "failed to update status")
			return ctrl.Result{}, statusErr
		}

		if err != nil {
			return ctrl.Result{}, err
		}
		return result, nil
	}

	// All conditions True: migration complete.
	log.Info("migration complete")
	return ctrl.Result{}, nil
}

// ensureInfrastructurePrepared validates preflight checks and disables the CVO.
func (r *VmwareCloudFoundationMigrationReconciler) ensureInfrastructurePrepared(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error) {
	log := klog.FromContext(ctx)
	condType := migrationv1alpha1.ConditionInfrastructurePrepared

	r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Validating preflight checks")

	// Validate failure domains are present.
	if len(migration.Spec.FailureDomains) == 0 {
		return ctrl.Result{}, fmt.Errorf("spec.failureDomains must not be empty")
	}

	// Validate target credentials secret exists.
	secretRef := migration.Spec.TargetVCenterCredentialsSecret
	if secretRef.Name == "" {
		return ctrl.Result{}, fmt.Errorf("spec.targetVCenterCredentialsSecret.name must not be empty")
	}
	ns := secretRef.Namespace
	if ns == "" {
		ns = migration.Namespace
	}
	if _, err := r.KubeClient.CoreV1().Secrets(ns).Get(ctx, secretRef.Name, metav1.GetOptions{}); err != nil {
		return ctrl.Result{}, fmt.Errorf("target credentials secret %s/%s not found: %w", ns, secretRef.Name, err)
	}

	// Get source vCenter from Infrastructure CRD.
	infraMgr := openshift.NewInfrastructureManager(r.ConfigClient, r.APIExtensionsClient)
	sourceVC, err := infraMgr.GetSourceVCenter(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting source vCenter: %w", err)
	}

	// Validate source connectivity.
	sm := openshift.NewSecretManager(r.KubeClient)
	srcUser, srcPass, err := sm.GetCredentials(ctx, sourceVC.Server)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting source vCenter credentials: %w", err)
	}

	var srcDC string
	if len(sourceVC.Datacenters) > 0 {
		srcDC = sourceVC.Datacenters[0]
	} else {
		return ctrl.Result{}, fmt.Errorf("source vCenter has no datacenters configured")
	}

	srcSession, err := getVSphereSession(ctx, sourceVC.Server, srcDC, srcUser, srcPass)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("connecting to source vCenter %s: %w", sourceVC.Server, err)
	}
	log.V(1).Info("source vCenter connectivity validated", "server", sourceVC.Server)

	// Validate source datacenter accessible.
	if _, err := srcSession.Finder.Datacenter(ctx, srcDC); err != nil {
		return ctrl.Result{}, fmt.Errorf("source datacenter %q not accessible: %w", srcDC, err)
	}

	// Validate each target failure domain's topology.
	for i := range migration.Spec.FailureDomains {
		fd := &migration.Spec.FailureDomains[i]
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing,
			fmt.Sprintf("Validating target failure domain %q", fd.Name))

		username, password, err := getTargetCredentials(ctx, r.KubeClient, migration, fd.Server)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting credentials for target %s: %w", fd.Server, err)
		}

		session, err := getVSphereSession(ctx, fd.Server, fd.Topology.Datacenter, username, password)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("connecting to target vCenter %s: %w", fd.Server, err)
		}

		// Validate datacenter.
		if _, err := session.Finder.Datacenter(ctx, fd.Topology.Datacenter); err != nil {
			return ctrl.Result{}, fmt.Errorf("target datacenter %q on %s not found: %w", fd.Topology.Datacenter, fd.Server, err)
		}

		// Validate compute cluster.
		if _, err := session.Finder.ClusterComputeResource(ctx, fd.Topology.ComputeCluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("target cluster %q on %s not found: %w", fd.Topology.ComputeCluster, fd.Server, err)
		}

		// Validate datastore.
		if _, err := session.Finder.Datastore(ctx, fd.Topology.Datastore); err != nil {
			return ctrl.Result{}, fmt.Errorf("target datastore %q on %s not found: %w", fd.Topology.Datastore, fd.Server, err)
		}

		// Validate networks.
		for _, net := range fd.Topology.Networks {
			if _, err := session.Finder.Network(ctx, net); err != nil {
				return ctrl.Result{}, fmt.Errorf("target network %q on %s not found: %w", net, fd.Server, err)
			}
		}

		// Validate resource pool.
		if fd.Topology.ResourcePool != "" {
			if _, err := session.Finder.ResourcePool(ctx, fd.Topology.ResourcePool); err != nil {
				return ctrl.Result{}, fmt.Errorf("target resource pool %q on %s not found: %w", fd.Topology.ResourcePool, fd.Server, err)
			}
		}

		// Validate template (folder path).
		if fd.Topology.Template != "" {
			if _, err := session.Finder.VirtualMachine(ctx, fd.Topology.Template); err != nil {
				return ctrl.Result{}, fmt.Errorf("target template %q on %s not found: %w", fd.Topology.Template, fd.Server, err)
			}
		}

		log.V(1).Info("target failure domain validated", "name", fd.Name, "server", fd.Server)
	}

	// Disable CVO.
	r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Disabling Cluster Version Operator")
	if err := disableCVO(ctx, r.KubeClient); err != nil {
		return ctrl.Result{}, fmt.Errorf("disabling CVO: %w", err)
	}

	r.setCondition(migration, condType, metav1.ConditionTrue, migrationv1alpha1.ReasonCompleted, "Infrastructure prepared and CVO disabled")
	r.Recorder.Event(migration, "Normal", "InfrastructurePrepared", "Preflight validation passed, CVO disabled")
	return ctrl.Result{}, nil
}

// ensureDestinationInitialized creates VM folders and region/zone tags on target vCenters.
func (r *VmwareCloudFoundationMigrationReconciler) ensureDestinationInitialized(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error) {
	log := klog.FromContext(ctx)
	condType := migrationv1alpha1.ConditionDestinationInitialized

	r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Initializing destination vCenter")

	infraMgr := openshift.NewInfrastructureManager(r.ConfigClient, r.APIExtensionsClient)
	infraID, err := infraMgr.GetInfrastructureID(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting infrastructure ID: %w", err)
	}

	// Track which server/datacenter pairs have had folders created to avoid duplicates.
	type serverDC struct {
		server, datacenter string
	}
	folderCreated := make(map[serverDC]bool)

	for i := range migration.Spec.FailureDomains {
		fd := &migration.Spec.FailureDomains[i]
		key := serverDC{server: fd.Server, datacenter: fd.Topology.Datacenter}

		username, password, err := getTargetCredentials(ctx, r.KubeClient, migration, fd.Server)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting credentials for %s: %w", fd.Server, err)
		}

		session, err := getVSphereSession(ctx, fd.Server, fd.Topology.Datacenter, username, password)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("connecting to %s/%s: %w", fd.Server, fd.Topology.Datacenter, err)
		}

		// Create VM folder per unique server/datacenter.
		if !folderCreated[key] {
			r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing,
				fmt.Sprintf("Creating VM folder %q on %s/%s", infraID, fd.Server, fd.Topology.Datacenter))

			folder, err := vsphere.CreateVMFolder(ctx, session, infraID)
			if err != nil {
				// Folder may already exist; try to get it.
				folder, getErr := vsphere.GetVMFolder(ctx, session, infraID)
				if getErr != nil {
					return ctrl.Result{}, fmt.Errorf("creating VM folder %q on %s/%s: %w", infraID, fd.Server, fd.Topology.Datacenter, err)
				}
				log.V(1).Info("VM folder already exists", "path", folder.InventoryPath)
			} else {
				log.V(1).Info("created VM folder", "path", folder.InventoryPath)
			}

			// Verify folder is accessible.
			if _, err := vsphere.GetVMFolder(ctx, session, infraID); err != nil {
				return ctrl.Result{}, fmt.Errorf("verifying VM folder %q on %s/%s: %w", infraID, fd.Server, fd.Topology.Datacenter, err)
			}
			folderCreated[key] = true
		}

		// Create region and zone tags.
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing,
			fmt.Sprintf("Creating tags for failure domain %q", fd.Name))

		regionTagID, zoneTagID, err := vsphere.CreateRegionAndZoneTags(ctx, session, fd.Region, fd.Zone)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("creating tags for failure domain %q: %w", fd.Name, err)
		}

		// Attach tags to datacenter and cluster.
		dc, err := session.Finder.Datacenter(ctx, fd.Topology.Datacenter)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("finding datacenter %q for tag attachment: %w", fd.Topology.Datacenter, err)
		}

		cluster, err := session.Finder.ClusterComputeResource(ctx, fd.Topology.ComputeCluster)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("finding cluster %q for tag attachment: %w", fd.Topology.ComputeCluster, err)
		}

		if err := vsphere.AttachFailureDomainTags(ctx, session, regionTagID, zoneTagID, dc, cluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("attaching tags for failure domain %q: %w", fd.Name, err)
		}

		log.V(1).Info("failure domain initialized", "name", fd.Name)
	}

	r.setCondition(migration, condType, metav1.ConditionTrue, migrationv1alpha1.ReasonCompleted, "Destination vCenter initialized with folders and tags")
	r.Recorder.Event(migration, "Normal", "DestinationInitialized", "VM folders and tags created on target vCenter")
	return ctrl.Result{}, nil
}

// ensureMultiSiteConfigured adds target vCenter to cluster configuration
// (secrets, Infrastructure CRD, cloud-provider-config) and restarts affected pods.
// Once config is applied and pods are restarted, it only waits for vSphere pods to
// become ready on subsequent reconciles (avoids re-restarting pods every requeue).
func (r *VmwareCloudFoundationMigrationReconciler) ensureMultiSiteConfigured(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error) {
	log := klog.FromContext(ctx)
	condType := migrationv1alpha1.ConditionMultiSiteConfigured

	currentMsg := ""
	if cond := apimeta.FindStatusCondition(migration.Status.Conditions, condType); cond != nil {
		currentMsg = cond.Message
	}

	// If we are already waiting for vSphere pods, only run the readiness check.
	// Otherwise we would re-apply config and restart pods on every requeue, preventing pods from ever becoming ready.
	waitingForPods := strings.HasPrefix(currentMsg, "Waiting for vSphere pods to become ready")

	if !waitingForPods {
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Configuring multi-site vCenter")

		sm := openshift.NewSecretManager(r.KubeClient)

		// Get vsphere-creds secret and add target credentials.
		credsSecret, err := sm.GetVSphereCredsSecret(ctx)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting vsphere-creds secret: %w", err)
		}

		// Track servers already processed to avoid duplicate updates.
		processedServers := make(map[string]bool)
		for i := range migration.Spec.FailureDomains {
			fd := &migration.Spec.FailureDomains[i]
			if processedServers[fd.Server] {
				continue
			}

			username, password, err := getTargetCredentials(ctx, r.KubeClient, migration, fd.Server)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("getting credentials for %s: %w", fd.Server, err)
			}

			credsSecret, err = sm.AddTargetVCenterCreds(ctx, credsSecret, fd.Server, username, password)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("adding target credentials for %s: %w", fd.Server, err)
			}
			processedServers[fd.Server] = true
		}
		log.V(1).Info("target vCenter credentials added to vsphere-creds")

		// Update Infrastructure CRD with target vCenter and failure domains.
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Updating Infrastructure CRD")
		infraMgr := openshift.NewInfrastructureManager(r.ConfigClient, r.APIExtensionsClient)
		infra, err := infraMgr.Get(ctx)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting infrastructure: %w", err)
		}

		if _, err := infraMgr.AddTargetVCenterWithCRDModification(ctx, infra, migration.Spec.FailureDomains); err != nil {
			return ctrl.Result{}, fmt.Errorf("adding target vCenter to infrastructure: %w", err)
		}
		log.V(1).Info("Infrastructure CRD updated with target vCenter")

		// Update cloud-provider-config.
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Updating cloud-provider-config")
		cmMgr := openshift.NewConfigMapManager(r.KubeClient)
		cm, err := cmMgr.GetCloudProviderConfig(ctx)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting cloud-provider-config: %w", err)
		}

		if _, err := cmMgr.AddTargetVCenterToConfig(ctx, cm, migration.Spec.FailureDomains); err != nil {
			return ctrl.Result{}, fmt.Errorf("adding target vCenter to cloud-provider-config: %w", err)
		}
		log.V(1).Info("cloud-provider-config updated")

		// Restart MCO pods.
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Restarting MCO pods")
		if err := syncControllerConfig(ctx, r.KubeClient); err != nil {
			return ctrl.Result{}, fmt.Errorf("syncing controller config (MCO restart): %w", err)
		}

		// Restart vSphere pods.
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Restarting vSphere pods")
		podMgr := openshift.NewPodManager(r.KubeClient)
		if err := podMgr.RestartVSpherePods(ctx); err != nil {
			return ctrl.Result{}, fmt.Errorf("restarting vSphere pods: %w", err)
		}

		// Set message so next reconcile only runs readiness check.
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Waiting for vSphere pods to become ready")
		// Fall through to readiness check below (no return).
	}

	podMgr := openshift.NewPodManager(r.KubeClient)
	r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Waiting for vSphere pods to become ready")
	podStatus, err := podMgr.CheckVSpherePodsReady(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking vSphere pods readiness: %w", err)
	}
	if !podStatus.AllReady {
		log.V(1).Info("vSphere pods not ready, requeueing", "reason", podStatus.NotReadyReason)
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	r.setCondition(migration, condType, metav1.ConditionTrue, migrationv1alpha1.ReasonCompleted, "Multi-site vCenter configured and pods ready")
	r.Recorder.Event(migration, "Normal", "MultiSiteConfigured", "Cluster configured for both source and target vCenters")
	return ctrl.Result{}, nil
}

// ensureWorkloadMigrated creates new worker MachineSets on target, rolls out
// control plane changes, and drains old MachineSets. Progress is derived from
// cluster state (what exists and is ready) so the controller is idempotent
// and safe to restart at any step.
func (r *VmwareCloudFoundationMigrationReconciler) ensureWorkloadMigrated(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error) {
	log := klog.FromContext(ctx)
	condType := migrationv1alpha1.ConditionWorkloadMigrated

	if len(migration.Spec.FailureDomains) == 0 {
		return ctrl.Result{}, fmt.Errorf("spec.failureDomains must not be empty")
	}
	for i := range migration.Spec.FailureDomains {
		if migration.Spec.FailureDomains[i].Topology.Template == "" {
			return ctrl.Result{}, fmt.Errorf("spec.failureDomains[%d].topology.template is required (failure domain %q)", i, migration.Spec.FailureDomains[i].Name)
		}
	}

	// If we are past Step 3 (CPMS updated), run Steps 4–6 (rollout and scale-down) from cluster state.
	if c := apimeta.FindStatusCondition(migration.Status.Conditions, condType); c != nil {
		pastCPMSUpdate := strings.HasPrefix(c.Message, "CPMS updated") || strings.Contains(c.Message, "Control plane rollout") || strings.Contains(c.Message, "Old workers")
		if pastCPMSUpdate {
			return r.ensureWorkloadMigratedRolloutAndScaleDown(ctx, migration)
		}
	}

	infraMgr := openshift.NewInfrastructureManager(r.ConfigClient, r.APIExtensionsClient)
	infraID, err := infraMgr.GetInfrastructureID(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting infrastructure ID: %w", err)
	}

	sourceVC, err := infraMgr.GetSourceVCenter(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting source vCenter: %w", err)
	}

	machineMgr := openshift.NewMachineManager(r.KubeClient, r.MachineClient, r.DynamicClient)

	// Step 1: Ensure target worker MachineSets exist (idempotent: create only missing ones).
	allTargetMSExist := true
	for i := range migration.Spec.FailureDomains {
		msName := fmt.Sprintf("%s-worker-%s", infraID, migration.Spec.FailureDomains[i].Name)
		if _, err := machineMgr.GetMachineSet(ctx, msName); err != nil {
			allTargetMSExist = false
			break
		}
	}
	if !allTargetMSExist {
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Creating new worker MachineSets")
		existingMachineSets, err := machineMgr.GetMachineSetsByVCenter(ctx, sourceVC.Server)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting existing MachineSets: %w", err)
		}
		if len(existingMachineSets) == 0 {
			return ctrl.Result{}, fmt.Errorf("no existing MachineSets found for source vCenter %s", sourceVC.Server)
		}
		var totalReplicas int32
		for _, ms := range existingMachineSets {
			if ms.Spec.Replicas != nil {
				totalReplicas += *ms.Spec.Replicas
			}
		}
		replicasPerFD := totalReplicas / int32(len(migration.Spec.FailureDomains))
		if replicasPerFD < 1 {
			replicasPerFD = 1
		}
		template := existingMachineSets[0]
		createdAny := false
		for i := range migration.Spec.FailureDomains {
			fd := &migration.Spec.FailureDomains[i]
			msName := fmt.Sprintf("%s-worker-%s", infraID, fd.Name)
			if _, err := machineMgr.GetMachineSet(ctx, msName); err == nil {
				log.V(1).Info("worker MachineSet already exists, skipping", "name", msName)
				continue
			}
			replicas := replicasPerFD
			if i == len(migration.Spec.FailureDomains)-1 {
				replicas = totalReplicas - replicasPerFD*int32(len(migration.Spec.FailureDomains)-1)
				if replicas < 1 {
					replicas = 1
				}
			}
			if _, err := machineMgr.CreateWorkerMachineSet(ctx, msName, replicas, fd, template, infraID); err != nil {
				return ctrl.Result{}, fmt.Errorf("creating worker MachineSet %q: %w", msName, err)
			}
			createdAny = true
			log.V(1).Info("created worker MachineSet", "name", msName, "replicas", replicas)
		}
		if createdAny {
			r.Recorder.Event(migration, "Normal", "WorkersCreated", "New worker MachineSets created on target vCenter")
		}
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Workers created, waiting for machines ready")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Step 2: Wait for target worker machines and nodes to be ready (cluster state).
	allReady, err := checkWorkerReadiness(ctx, machineMgr, migration.Spec.FailureDomains, infraID)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !allReady {
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Workers created, waiting for machines ready")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Step 3: Update CPMS with target failure domains and set state to Active.
	// The CPMS is updated in place — no delete/recreate needed. The CPMS operator
	// resolves failure domain topology from the Infrastructure resource and triggers
	// a rolling replacement of control plane machines.
	targetFDNames := failureDomainNames(migration.Spec.FailureDomains)
	if err := machineMgr.UpdateCPMSFailureDomain(ctx, targetFDNames); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating CPMS failure domains: %w", err)
	}
	r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "CPMS updated, waiting for generation observed")
	r.Recorder.Event(migration, "Normal", "CPMSUpdated", fmt.Sprintf("CPMS updated with failure domains %v", targetFDNames))
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// ensureWorkloadMigratedRolloutAndScaleDown runs Steps 5–7: wait for control plane
// rollout, scale old MachineSets to 0, wait for old machines/nodes to be deleted.
// Progress is derived from cluster state so it is idempotent. Call when condition
// message indicates we are past "CPMS updated" (e.g. "Control plane rollout" or
// "Old workers" or we have observed generation and rollout complete).
func (r *VmwareCloudFoundationMigrationReconciler) ensureWorkloadMigratedRolloutAndScaleDown(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error) {
	log := klog.FromContext(ctx)
	condType := migrationv1alpha1.ConditionWorkloadMigrated
	infraMgr := openshift.NewInfrastructureManager(r.ConfigClient, r.APIExtensionsClient)
	sourceVC, err := infraMgr.GetSourceVCenter(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting source vCenter: %w", err)
	}
	machineMgr := openshift.NewMachineManager(r.KubeClient, r.MachineClient, r.DynamicClient)

	// Step 5: Wait for CPMS generation observed and rollout complete.
	observed, err := machineMgr.IsCPMSGenerationObserved(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking CPMS generation: %w", err)
	}
	if !observed {
		log.V(1).Info("CPMS generation not yet observed")
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "CPMS updated, waiting for generation observed")
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}
	complete, replicas, updated, ready, err := machineMgr.CheckControlPlaneRolloutStatus(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking control plane rollout: %w", err)
	}
	if !complete {
		log.V(1).Info("control plane rollout in progress", "replicas", replicas, "updated", updated, "ready", ready)
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing,
			fmt.Sprintf("CPMS updated, control plane rolling out (%d/%d ready)", ready, replicas))
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Step 6: Scale old MachineSets to 0 (idempotent).
	oldMachineSets, err := machineMgr.GetMachineSetsByVCenter(ctx, sourceVC.Server)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting old MachineSets: %w", err)
	}
	scaledAny := false
	for _, ms := range oldMachineSets {
		if ms.Spec.Replicas != nil && *ms.Spec.Replicas > 0 {
			if err := machineMgr.ScaleMachineSet(ctx, ms.Name, 0); err != nil {
				return ctrl.Result{}, fmt.Errorf("scaling old MachineSet %q to 0: %w", ms.Name, err)
			}
			scaledAny = true
			log.V(1).Info("scaled old MachineSet to 0", "name", ms.Name)
		}
	}
	if scaledAny {
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Old workers scaled down, waiting for deletion")
		r.Recorder.Event(migration, "Normal", "ControlPlaneRolledOut", "Control plane rollout complete on target vCenter")
		r.Recorder.Event(migration, "Normal", "OldWorkersScaledDown", "Old worker MachineSets scaled to 0")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Step 7: Wait for old machines and nodes to be deleted.
	allDeleted := true
	for _, ms := range oldMachineSets {
		machinesGone, remaining, err := machineMgr.CheckMachinesDeleted(ctx, ms.Name)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("checking machine deletion for %q: %w", ms.Name, err)
		}
		if !machinesGone {
			log.V(1).Info("old machines still present", "machineSet", ms.Name, "remaining", remaining)
			allDeleted = false
			continue
		}
		nodesGone, nodeRemaining, err := machineMgr.CheckNodesDeletedForMachines(ctx, ms.Name)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("checking node deletion for %q: %w", ms.Name, err)
		}
		if !nodesGone {
			log.V(1).Info("old nodes still present", "machineSet", ms.Name, "remaining", nodeRemaining)
			allDeleted = false
		}
	}
	if !allDeleted {
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Old workers scaled down, waiting for deletion")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	r.setCondition(migration, condType, metav1.ConditionTrue, migrationv1alpha1.ReasonCompleted, "Workload migrated to target vCenter")
	r.Recorder.Event(migration, "Normal", "WorkloadMigrated", "All workloads migrated to target vCenter")
	return ctrl.Result{}, nil
}

// ensureSourceCleaned removes the source vCenter from all cluster configuration,
// generates metadata, and re-enables the CVO.
func (r *VmwareCloudFoundationMigrationReconciler) ensureSourceCleaned(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error) {
	log := klog.FromContext(ctx)
	condType := migrationv1alpha1.ConditionSourceCleaned

	r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Cleaning up source vCenter")

	infraMgr := openshift.NewInfrastructureManager(r.ConfigClient, r.APIExtensionsClient)
	sourceVC, err := infraMgr.GetSourceVCenter(ctx)
	if err != nil {
		// Source may already have been removed on a previous attempt.
		log.V(1).Info("source vCenter not found in infrastructure, may already be removed", "err", err)
	}

	var sourceServer string
	if sourceVC != nil {
		sourceServer = sourceVC.Server
	}

	// Complete all vCenter list changes (Infrastructure, config, secrets, pods) before
	// re-enabling CVO. CVO must not be re-enabled until the cluster config reflects
	// only the target vCenter(s).
	if sourceServer != "" {
		// Remove source from Infrastructure CRD (requires temporary CRD modification;
		// platform validation forbids adding/removing vCenters once set).
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Removing source vCenter from Infrastructure CRD")
		infra, err := infraMgr.Get(ctx)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting infrastructure: %w", err)
		}
		if _, err := infraMgr.RemoveSourceVCenterWithCRDModification(ctx, infra, sourceServer); err != nil {
			return ctrl.Result{}, fmt.Errorf("removing source vCenter from infrastructure: %w", err)
		}
		log.V(1).Info("removed source vCenter from infrastructure", "server", sourceServer)

		// Remove source from cloud-provider-config.
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Removing source vCenter from cloud-provider-config")
		cmMgr := openshift.NewConfigMapManager(r.KubeClient)
		cm, err := cmMgr.GetCloudProviderConfig(ctx)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting cloud-provider-config: %w", err)
		}

		if _, err := cmMgr.RemoveSourceVCenterFromConfig(ctx, cm, sourceServer); err != nil {
			return ctrl.Result{}, fmt.Errorf("removing source vCenter from cloud-provider-config: %w", err)
		}
		log.V(1).Info("removed source vCenter from cloud-provider-config", "server", sourceServer)

		// Remove source creds from vsphere-creds.
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Removing source credentials from vsphere-creds")
		sm := openshift.NewSecretManager(r.KubeClient)
		credsSecret, err := sm.GetVSphereCredsSecret(ctx)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting vsphere-creds secret: %w", err)
		}

		if _, err := sm.RemoveSourceVCenterCreds(ctx, credsSecret, sourceServer); err != nil {
			return ctrl.Result{}, fmt.Errorf("removing source credentials: %w", err)
		}
		log.V(1).Info("removed source credentials from vsphere-creds", "server", sourceServer)

		// Restart vSphere pods to pick up new config.
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Restarting vSphere pods after source removal")
		podMgr := openshift.NewPodManager(r.KubeClient)
		if err := podMgr.RestartVSpherePods(ctx); err != nil {
			return ctrl.Result{}, fmt.Errorf("restarting vSphere pods: %w", err)
		}
	}

	// Generate metadata (after vCenter list is updated).
	r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Generating migration metadata")
	infra, err := infraMgr.Get(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting infrastructure for metadata: %w", err)
	}

	// Build credentials map for metadata generation.
	sm := openshift.NewSecretManager(r.KubeClient)
	creds := make(map[string]string)
	for i := range migration.Spec.FailureDomains {
		fd := &migration.Spec.FailureDomains[i]
		if _, exists := creds[fd.Server]; exists {
			continue
		}
		username, password, err := getTargetCredentials(ctx, r.KubeClient, migration, fd.Server)
		if err != nil {
			// Fall back to vsphere-creds.
			username, password, err = sm.GetCredentials(ctx, fd.Server)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("getting credentials for metadata for %s: %w", fd.Server, err)
			}
		}
		creds[fd.Server] = fmt.Sprintf("%s:%s", username, password)
	}

	metadataMgr := metadata.NewMetadataManager(r.KubeClient)
	md, err := metadataMgr.GenerateMetadata(ctx, migration.Spec.FailureDomains, infra, creds)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("generating metadata: %w", err)
	}

	secretName := metadata.GetMetadataSecretName(migration.Name)
	if err := metadataMgr.SaveToSecret(ctx, md, migration.Namespace, secretName); err != nil {
		return ctrl.Result{}, fmt.Errorf("saving metadata: %w", err)
	}
	log.V(1).Info("metadata saved", "secret", secretName)

	// Re-enable CVO only after vCenter list and config have been updated (above).
	r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Re-enabling Cluster Version Operator")
	if err := enableCVO(ctx, r.KubeClient); err != nil {
		return ctrl.Result{}, fmt.Errorf("enabling CVO: %w", err)
	}

	// Wait for CVO to become ready.
	ready, err := isCVOReady(ctx, r.KubeClient)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking CVO readiness: %w", err)
	}
	if !ready {
		log.V(1).Info("CVO not yet ready, requeueing")
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Waiting for CVO to become ready")
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	r.setCondition(migration, condType, metav1.ConditionTrue, migrationv1alpha1.ReasonCompleted, "Source vCenter cleaned and CVO re-enabled")
	r.Recorder.Event(migration, "Normal", "SourceCleaned", "Source vCenter removed from cluster configuration, CVO re-enabled")
	return ctrl.Result{}, nil
}

// ensureReady verifies all operators are healthy and only target vCenters remain
// in the Infrastructure resource.
func (r *VmwareCloudFoundationMigrationReconciler) ensureReady(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error) {
	log := klog.FromContext(ctx)
	condType := migrationv1alpha1.ConditionReady

	r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, "Verifying final cluster state")

	// Check all operators healthy.
	opMgr := openshift.NewOperatorManager(r.ConfigClient)
	healthy, unhealthy, err := opMgr.CheckAllOperatorsHealthy(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking operator health: %w", err)
	}
	if !healthy {
		msg := fmt.Sprintf("Unhealthy operators: %s", strings.Join(unhealthy, ", "))
		log.V(1).Info("operators not all healthy", "unhealthy", unhealthy)
		r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, msg)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Verify only target vCenters remain in Infrastructure.
	infraMgr := openshift.NewInfrastructureManager(r.ConfigClient, r.APIExtensionsClient)
	infra, err := infraMgr.Get(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting infrastructure for readiness check: %w", err)
	}

	if infra.Spec.PlatformSpec.VSphere != nil {
		targetServers := make(map[string]bool)
		for i := range migration.Spec.FailureDomains {
			targetServers[migration.Spec.FailureDomains[i].Server] = true
		}

		for _, vc := range infra.Spec.PlatformSpec.VSphere.VCenters {
			if !targetServers[vc.Server] {
				msg := fmt.Sprintf("Non-target vCenter %q still present in Infrastructure", vc.Server)
				log.V(1).Info("unexpected vCenter in infrastructure", "server", vc.Server)
				r.setCondition(migration, condType, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, msg)
				return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
			}
		}
	}

	// Set completion time.
	now := metav1.Now()
	migration.Status.CompletionTime = &now

	r.setCondition(migration, condType, metav1.ConditionTrue, migrationv1alpha1.ReasonCompleted, "Migration complete, all operators healthy")
	r.Recorder.Event(migration, "Normal", "MigrationComplete", "Migration completed successfully")
	return ctrl.Result{}, nil
}

// setCondition is a convenience wrapper around apimeta.SetStatusCondition.
func (r *VmwareCloudFoundationMigrationReconciler) setCondition(migration *migrationv1alpha1.VmwareCloudFoundationMigration, conditionType string, status metav1.ConditionStatus, reason, message string) {
	apimeta.SetStatusCondition(&migration.Status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		ObservedGeneration: migration.Generation,
		Reason:             reason,
		Message:            message,
	})
}

// isConditionTrue checks whether the named condition has status True.
func (r *VmwareCloudFoundationMigrationReconciler) isConditionTrue(migration *migrationv1alpha1.VmwareCloudFoundationMigration, conditionType string) bool {
	cond := apimeta.FindStatusCondition(migration.Status.Conditions, conditionType)
	return cond != nil && cond.Status == metav1.ConditionTrue
}

// updateStatus persists the migration status subresource. It retries on conflict by
// re-fetching the latest resource and re-applying the desired status.
func (r *VmwareCloudFoundationMigrationReconciler) updateStatus(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration) error {
	desiredStatus := migration.Status.DeepCopy()
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &migrationv1alpha1.VmwareCloudFoundationMigration{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(migration), latest); err != nil {
			return err
		}
		latest.Status = *desiredStatus
		return r.Status().Update(ctx, latest)
	})
	if err != nil {
		return fmt.Errorf("updating migration status: %w", err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VmwareCloudFoundationMigrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&migrationv1alpha1.VmwareCloudFoundationMigration{}).
		Named("vmwarecloudfoundationmigration").
		Complete(r)
}

// failureDomainNames extracts the name from each failure domain spec.
func failureDomainNames(fds []configv1.VSpherePlatformFailureDomainSpec) []string {
	names := make([]string, len(fds))
	for i := range fds {
		names[i] = fds[i].Name
	}
	return names
}

// checkWorkerReadiness verifies that all machines and nodes for the target worker
// MachineSets are in a ready state. It returns true when every MachineSet's machines
// are Running with a NodeRef and the corresponding nodes have condition Ready=True.
func checkWorkerReadiness(ctx context.Context, machineMgr *openshift.MachineManager, fds []configv1.VSpherePlatformFailureDomainSpec, infraID string) (bool, error) {
	log := klog.FromContext(ctx)
	for i := range fds {
		msName := fmt.Sprintf("%s-worker-%s", infraID, fds[i].Name)
		machinesReady, readyCount, totalCount, err := machineMgr.CheckMachinesReady(ctx, msName)
		if err != nil {
			return false, fmt.Errorf("checking machines for %q: %w", msName, err)
		}
		if !machinesReady {
			log.V(1).Info("machines not ready", "machineSet", msName, "ready", readyCount, "total", totalCount)
			return false, nil
		}
		nodesReady, nodeReadyCount, nodeTotalCount, err := machineMgr.CheckNodesReady(ctx, msName)
		if err != nil {
			return false, fmt.Errorf("checking nodes for %q: %w", msName, err)
		}
		if !nodesReady {
			log.V(1).Info("nodes not ready", "machineSet", msName, "ready", nodeReadyCount, "total", nodeTotalCount)
			return false, nil
		}
	}
	return true, nil
}
