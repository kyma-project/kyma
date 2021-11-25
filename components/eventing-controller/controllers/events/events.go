package events

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

// reason is the reason of an event.
type reason string

const (
	// ReasonCreate is used when an object is successfully created.
	ReasonCreate reason = "Create"
	// ReasonCreateFailed is used when an object creation fails.
	ReasonCreateFailed reason = "CreateFailed"
	// ReasonUpdate is used when an object is successfully updated.
	ReasonUpdate reason = "Update"
	// ReasonUpdateFailed is used when an object update fails.
	ReasonUpdateFailed reason = "UpdateFailed"
	// ReasonValidationFailed is used when an object validation fails.
	ReasonValidationFailed reason = "ValidationFailed"
)

// eventNormal records a normal event for an API object.
func EventNormal(recorder record.EventRecorder, obj runtime.Object, rn reason, msgFmt string, args ...interface{}) {
	recorder.Eventf(obj, corev1.EventTypeNormal, string(rn), msgFmt, args...)
}

// eventWarn records a warning event for an API object.
func EventWarn(recorder record.EventRecorder, obj runtime.Object, rn reason, msgFmt string, args ...interface{}) {
	recorder.Eventf(obj, corev1.EventTypeWarning, string(rn), msgFmt, args...)
}
