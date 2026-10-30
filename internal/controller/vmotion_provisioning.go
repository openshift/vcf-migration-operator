package controller

import (
	"context"
	"fmt"

	machinev1 "github.com/openshift/api/machine/v1"
	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

type vmotionMachineAPIInventory struct {
	machines      []machinev1beta1.Machine
	sets          []machinev1beta1.MachineSet
	controlPlanes []machinev1.ControlPlaneMachineSet
	checks        []machinev1beta1.MachineHealthCheck
}

func (r *VmwareCloudFoundationMigrationReconciler) readVMotionMachineAPI(ctx context.Context) (vmotionMachineAPIInventory, error) {
	machines, err := r.MachineClient.MachineV1beta1().Machines("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return vmotionMachineAPIInventory{}, fmt.Errorf("listing Machines for VMotion eligibility: %w", err)
	}
	sets, err := r.MachineClient.MachineV1beta1().MachineSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return vmotionMachineAPIInventory{}, fmt.Errorf("listing machinesets for VMotion quiescence: %w", err)
	}
	controlPlanes, err := r.MachineClient.MachineV1().ControlPlaneMachineSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return vmotionMachineAPIInventory{}, fmt.Errorf("listing controlplanemachinesets for VMotion quiescence: %w", err)
	}
	checks, err := r.MachineClient.MachineV1beta1().MachineHealthChecks("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return vmotionMachineAPIInventory{}, fmt.Errorf("listing machinehealthchecks for VMotion quiescence: %w", err)
	}
	return vmotionMachineAPIInventory{machines: machines.Items, sets: sets.Items, controlPlanes: controlPlanes.Items, checks: checks.Items}, nil
}

func (i vmotionMachineAPIInventory) validate(identities []vmotionNodeIdentity) error {
	if err := validateVMotionMachineEligibility(identities, i.machines); err != nil {
		return err
	}
	return validateVMotionProvisioning(i.machines, i.sets, i.controlPlanes, i.checks)
}

// validateVMotionProvisioning observes Machine API intent and children, not just
// phase or replica summaries. It does not fence other controllers or writers.
func validateVMotionProvisioning(machines []machinev1beta1.Machine, sets []machinev1beta1.MachineSet, controlPlanes []machinev1.ControlPlaneMachineSet, checks []machinev1beta1.MachineHealthCheck) error {
	owners := make(map[types.UID]string, len(sets)+len(controlPlanes))
	for i := range sets {
		ms := &sets[i]
		if err := validateVMotionMachineSet(ms, machines); err != nil {
			return err
		}
		if ms.UID != "" {
			owners[ms.UID] = machinev1beta1.GroupVersion.String() + "/MachineSet/" + ms.Namespace + "/" + ms.Name
		}
	}
	for i := range controlPlanes {
		cpms := &controlPlanes[i]
		if err := validateVMotionControlPlane(cpms, machines); err != nil {
			return err
		}
		if cpms.UID != "" {
			owners[cpms.UID] = machinev1.GroupVersion.String() + "/ControlPlaneMachineSet/" + cpms.Namespace + "/" + cpms.Name
		}
	}
	for i := range machines {
		m := &machines[i]
		if owner := metav1.GetControllerOf(m); owner != nil {
			if owner.UID == "" || owners[owner.UID] != owner.APIVersion+"/"+owner.Kind+"/"+m.Namespace+"/"+owner.Name {
				return fmt.Errorf("checking Machine %s/%s: controller reference is unresolved or unsupported; lifecycle is indeterminate", m.Namespace, m.Name)
			}
		}
	}
	return validateVMotionRemediation(machines, checks)
}

func selectedVMotionMachines(namespace string, selector metav1.LabelSelector, machines []machinev1beta1.Machine) ([]*machinev1beta1.Machine, error) {
	match, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return nil, fmt.Errorf("parsing Machine selector: %w", err)
	}
	var selected []*machinev1beta1.Machine
	for i := range machines {
		m := &machines[i]
		if m.Namespace == namespace && match.Matches(labels.Set(m.Labels)) {
			selected = append(selected, m)
		}
	}
	return selected, nil
}

func validateVMotionMachineSet(ms *machinev1beta1.MachineSet, machines []machinev1beta1.Machine) error {
	if !ms.DeletionTimestamp.IsZero() || ms.Spec.Replicas == nil || *ms.Spec.Replicas < 0 {
		return fmt.Errorf("checking MachineSet %s/%s: deletion or unknown desired capacity is unsettled", ms.Namespace, ms.Name)
	}
	if len(ms.Spec.Selector.MatchLabels)+len(ms.Spec.Selector.MatchExpressions) == 0 {
		return fmt.Errorf("checking MachineSet %s/%s: empty selector makes provisioning intent indeterminate", ms.Namespace, ms.Name)
	}
	selector, err := metav1.LabelSelectorAsSelector(&ms.Spec.Selector)
	if err != nil {
		return fmt.Errorf("checking MachineSet %s/%s selector: %w", ms.Namespace, ms.Name, err)
	}
	if !selector.Matches(labels.Set(ms.Spec.Template.Labels)) {
		return fmt.Errorf("checking MachineSet %s/%s: template does not match selector; provisioning is unsettled", ms.Namespace, ms.Name)
	}
	selected, err := selectedVMotionMachines(ms.Namespace, ms.Spec.Selector, machines)
	if err != nil {
		return err
	}
	actual := int32(0)
	for _, m := range selected {
		if owner := metav1.GetControllerOf(m); owner != nil && owner.UID != ms.UID {
			continue
		}
		if !m.DeletionTimestamp.IsZero() || m.Status.Phase == nil || *m.Status.Phase != machinev1beta1.PhaseRunning {
			return fmt.Errorf("checking MachineSet %s/%s: child Machine %s has an unsettled lifecycle", ms.Namespace, ms.Name, m.Name)
		}
		actual++
	}
	desired := *ms.Spec.Replicas
	if ms.Status.ObservedGeneration != ms.Generation || actual != desired || ms.Status.Replicas != desired ||
		ms.Status.ReadyReplicas != desired || ms.Status.AvailableReplicas != desired ||
		(ms.Status.ErrorReason != nil && *ms.Status.ErrorReason != "") || (ms.Status.ErrorMessage != nil && *ms.Status.ErrorMessage != "") {
		return fmt.Errorf("checking MachineSet %s/%s: provisioning or scale-down has not settled (desired %d, observed children %d); require current generation and matching ready/available counts", ms.Namespace, ms.Name, desired, actual)
	}
	return nil
}

func validateVMotionControlPlane(cpms *machinev1.ControlPlaneMachineSet, machines []machinev1beta1.Machine) error {
	if !cpms.DeletionTimestamp.IsZero() {
		return fmt.Errorf("checking ControlPlaneMachineSet %s/%s: deletion has not settled", cpms.Namespace, cpms.Name)
	}
	if cpms.Spec.State == machinev1.ControlPlaneMachineSetStateInactive {
		return nil
	}
	if cpms.Spec.State != machinev1.ControlPlaneMachineSetStateActive || cpms.Spec.Replicas == nil || *cpms.Spec.Replicas <= 0 ||
		len(cpms.Spec.Selector.MatchLabels)+len(cpms.Spec.Selector.MatchExpressions) == 0 {
		return fmt.Errorf("checking ControlPlaneMachineSet %s/%s: state, desired capacity or selector is indeterminate", cpms.Namespace, cpms.Name)
	}
	selected, err := selectedVMotionMachines(cpms.Namespace, cpms.Spec.Selector, machines)
	if err != nil {
		return fmt.Errorf("checking ControlPlaneMachineSet %s/%s: %w", cpms.Namespace, cpms.Name, err)
	}
	for _, m := range selected {
		if !m.DeletionTimestamp.IsZero() || m.Status.Phase == nil || *m.Status.Phase != machinev1beta1.PhaseRunning {
			return fmt.Errorf("checking ControlPlaneMachineSet %s/%s: child Machine %s has an unsettled lifecycle", cpms.Namespace, cpms.Name, m.Name)
		}
	}
	desired := *cpms.Spec.Replicas
	if cpms.Status.ObservedGeneration != cpms.Generation || int32(len(selected)) != desired || cpms.Status.Replicas != desired ||
		cpms.Status.ReadyReplicas != desired || cpms.Status.UpdatedReplicas != desired || cpms.Status.UnavailableReplicas != 0 {
		return fmt.Errorf("checking ControlPlaneMachineSet %s/%s: provisioning or replacement has not settled", cpms.Namespace, cpms.Name)
	}
	for _, want := range []struct {
		name   string
		status metav1.ConditionStatus
	}{{"Available", metav1.ConditionTrue}, {"Progressing", metav1.ConditionFalse}, {"Degraded", metav1.ConditionFalse}} {
		cond := apimeta.FindStatusCondition(cpms.Status.Conditions, want.name)
		if cond == nil || cond.Status != want.status || cond.ObservedGeneration != cpms.Generation {
			return fmt.Errorf("checking ControlPlaneMachineSet %s/%s: %s condition does not establish current settled state", cpms.Namespace, cpms.Name, want.name)
		}
	}
	if cond := apimeta.FindStatusCondition(cpms.Status.Conditions, "Error"); cond != nil && (cond.Status != metav1.ConditionFalse || cond.ObservedGeneration != cpms.Generation) {
		return fmt.Errorf("checking ControlPlaneMachineSet %s/%s: controller error state is unresolved", cpms.Namespace, cpms.Name)
	}
	return nil
}

func validateVMotionRemediation(machines []machinev1beta1.Machine, checks []machinev1beta1.MachineHealthCheck) error {
	for i := range checks {
		check := &checks[i]
		selected, err := selectedVMotionMachines(check.Namespace, check.Spec.Selector, machines)
		if err != nil {
			return fmt.Errorf("checking MachineHealthCheck %s/%s: %w", check.Namespace, check.Name, err)
		}
		for _, m := range selected {
			// Native internal remediation skips ownerless Failed Machines. External
			// templates and controller-owned Machines do not have that guarantee;
			// a pause annotation alone cannot prove prior remediation has ended.
			if check.Spec.RemediationTemplate != nil || metav1.GetControllerOf(m) != nil || m.Status.Phase == nil || *m.Status.Phase != machinev1beta1.PhaseFailed {
				return fmt.Errorf("checking MachineHealthCheck %s/%s: remediation for Machine %s may still change inventory; establish inactivity before VMotion", check.Namespace, check.Name, m.Name)
			}
		}
	}
	return nil
}

func validateVMotionMachinePhase(m *machinev1beta1.Machine) error {
	phase := ""
	if m.Status.Phase != nil {
		phase = *m.Status.Phase
	}
	if !m.DeletionTimestamp.IsZero() || phase == machinev1beta1.PhaseProvisioning || phase == machinev1beta1.PhaseProvisioned || phase == machinev1beta1.PhaseDeleting {
		return fmt.Errorf("machine %s/%s is in active provisioning phase %q; wait for provisioning or deletion to settle before VMotion", m.Namespace, m.Name, phase)
	}
	if phase == machinev1beta1.PhaseFailed {
		return nil
	}
	if phase == machinev1beta1.PhaseRunning {
		return fmt.Errorf("checking Machine %s/%s: Running without a verified association to a current Node is inconsistent; lifecycle is indeterminate", m.Namespace, m.Name)
	}
	return fmt.Errorf("checking Machine %s/%s: phase %q is uninitialized or unrecognized; lifecycle is indeterminate", m.Namespace, m.Name, phase)
}
