package util

import (
	"unicode/utf8"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/events"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
)

const (
	// MaxConditionMessageBytes is the Kubernetes API character limit on condition messages.
	MaxConditionMessageBytes = 32768
	// MaxEventNoteBytes is the Kubernetes API byte limit on event notes.
	MaxEventNoteBytes = 1024
)

// SetCondition is a convenience wrapper around apimeta.SetStatusCondition. It
// reports whether the condition's status, reason, message, or observed
// generation changed, so callers can emit an event only on a transition.
func SetCondition(migration *migrationv1alpha1.VmwareCloudFoundationMigration, conditionType string, status metav1.ConditionStatus, reason, message string) bool {
	return apimeta.SetStatusCondition(&migration.Status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		ObservedGeneration: migration.Generation,
		Reason:             reason,
		Message:            message,
	})
}

// SetReadyRunning marks Ready as progressing and records an event when that
// changes the condition.
func SetReadyRunning(migration *migrationv1alpha1.VmwareCloudFoundationMigration, recorder events.EventRecorder) {
	msg := "Migration is running"
	changed := SetCondition(migration, migrationv1alpha1.ConditionReady, metav1.ConditionFalse, migrationv1alpha1.ReasonProgressing, msg)
	if changed && recorder != nil {
		recorder.Eventf(migration, nil, "Normal", migrationv1alpha1.ReasonProgressing, migrationv1alpha1.ReasonProgressing, "%s", msg)
	}
}

// StatusConditionChanged reports whether two conditions of the same type
// differ in any field this operator writes. It identifies which conditions a
// reconcile actually changed relative to the status it started from.
func StatusConditionChanged(prev, next metav1.Condition) bool {
	return prev.Status != next.Status ||
		prev.Reason != next.Reason ||
		prev.Message != next.Message ||
		prev.ObservedGeneration != next.ObservedGeneration ||
		!prev.LastTransitionTime.Equal(&next.LastTransitionTime)
}

// TruncateToBytes truncates s to at most maxBytes bytes without splitting a UTF-8 rune.
func TruncateToBytes(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}
