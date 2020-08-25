package httpsource

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// reason is the reason of an event.
type reason string

const (
	// createReason is used when an object is successfully created.
	createReason reason = "Create"
	// updateReason is used when an object is successfully updated.
	updateReason reason = "Update"
	// failedCreateReason is used when an object creation fails.
	failedCreateReason reason = "FailedCreate"
	// failedUpdateReason is used when an object update fails.
	failedUpdateReason reason = "FailedUpdate"
	// deleteReason is used when an object is successfully deleted.
	deleteReason reason = "Delete"
)

// event records a normal event for an API object.
func (r *Reconciler) event(obj runtime.Object, rn reason, msgFmt string, args ...interface{}) {
	r.Recorder.Eventf(obj, corev1.EventTypeNormal, string(rn), msgFmt, args...)
}

// eventWarn records a warning event for an API object.
func (r *Reconciler) eventWarn(obj runtime.Object, rn reason, msgFmt string, args ...interface{}) {
	r.Recorder.Eventf(obj, corev1.EventTypeWarning, string(rn), msgFmt, args...)
}
