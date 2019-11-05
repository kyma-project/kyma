/*
Copyright 2019 The Kyma Authors.

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
	failedCreateReason reason = "CreateUpdate"
	// failedUpdateReason is used when an object update fails.
	failedUpdateReason reason = "FailedUpdate"
)

// event records a normal event for an API object.
func (r *Reconciler) event(obj runtime.Object, rn reason, msgFmt string, args ...interface{}) {
	r.Recorder.Eventf(obj, corev1.EventTypeNormal, string(rn), msgFmt, args...)
}

// eventWarn records a warning event for an API object.
func (r *Reconciler) eventWarn(obj runtime.Object, rn reason, msgFmt string, args ...interface{}) {
	r.Recorder.Eventf(obj, corev1.EventTypeWarning, string(rn), msgFmt, args...)
}
