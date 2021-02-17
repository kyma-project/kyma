package subscription

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// reason is the reason of an event.
type reason string

const (
	// reasonCreate is used when an object is successfully created.
	reasonCreate reason = "Create"
	// reasonCreateFailed is used when an object creation fails.
	reasonCreateFailed reason = "CreateFailed"
	// reasonUpdate is used when an object is successfully updated.
	reasonUpdate reason = "Update"
	// reasonUpdateFailed is used when an object update fails.
	reasonUpdateFailed reason = "UpdateFailed"
	// reasonDelete is used when an object is successfully deleted.
	reasonDelete reason = "Delete"
	// reasonDeleteFailed is used when an object delete fails.
	reasonDeleteFailed reason = "DeleteFailed"
	// reasonValidationFailed is used when an object validation fails.
	reasonValidationFailed reason = "ValidationFailed"
)

// eventNormal records a normal event for an API object.
func (r *Reconciler) eventNormal(obj runtime.Object, rn reason, msgFmt string, args ...interface{}) {
	r.recorder.Eventf(obj, corev1.EventTypeNormal, string(rn), msgFmt, args...)
}

// eventWarn records a warning event for an API object.
func (r *Reconciler) eventWarn(obj runtime.Object, rn reason, msgFmt string, args ...interface{}) {
	r.recorder.Eventf(obj, corev1.EventTypeWarning, string(rn), msgFmt, args...)
}
