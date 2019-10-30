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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	// MQTTConditionReady has status True when the MQTTSource is ready to
	// send events.
	MQTTConditionReady = apis.ConditionReady

	// MQTTConditionSinkProvided has status True when the MQTTSource has
	// been configured with a sink target.
	MQTTConditionSinkProvided apis.ConditionType = "SinkProvided"

	// MQTTConditionDeployed has status True when the MQTTSource adapter
	// has been successfully deployed.
	MQTTConditionDeployed apis.ConditionType = "Deployed"
)

var mqttCondSet = apis.NewLivingConditionSet(
	MQTTConditionSinkProvided,
	MQTTConditionDeployed,
)

// MQTTSourceGVK returns a GroupVersionKind for the MQTTSource type.
func MQTTSourceGVK() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("MQTTSource")
}

// ToOwner return a OwnerReference corresponding to the given MQTTSource.
func (s *MQTTSource) ToOwner() *metav1.OwnerReference {
	return metav1.NewControllerRef(s, MQTTSourceGVK())
}

// ToKey returns a key corresponding to the MQTTSource.
func (s *MQTTSource) ToKey() string {
	return s.Namespace + "/" + s.Name
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *MQTTSourceStatus) InitializeConditions() {
	mqttCondSet.Manage(s).InitializeConditions()
}

// MarkSink sets the SinkProvided condition to True using the given URI.
func (s *MQTTSourceStatus) MarkSink(uri string) {
	s.SinkURI = uri
	if uri == "" {
		mqttCondSet.Manage(s).MarkUnknown(MQTTConditionSinkProvided,
			"SinkEmpty", "The sink has no URI")
		return
	}
	mqttCondSet.Manage(s).MarkTrue(MQTTConditionSinkProvided)
}

// MarkNoSink sets the SinkProvided condition to False with the given reason
// and message.
func (s *MQTTSourceStatus) MarkNoSink(reason, msg string) {
	mqttCondSet.Manage(s).MarkFalse(MQTTConditionSinkProvided,
		reason, msg)
}

// PropagateServiceReady uses the readiness of the provided Knative Service to
// determine whether the Deployed condition should be marked as true or false.
func (s *MQTTSourceStatus) PropagateServiceReady(ksvc *servingv1.Service) {
	if ksvc.Status.IsReady() {
		mqttCondSet.Manage(s).MarkTrue(MQTTConditionDeployed)
		return
	}

	msg := "The adapter Service is not yet ready"
	if ksvcCondReady := ksvc.Status.GetCondition(servingv1.ServiceConditionReady); ksvcCondReady != nil {
		msg += ": " + ksvcCondReady.Message
	}
	mqttCondSet.Manage(s).MarkFalse(MQTTConditionDeployed, "KnativeServiceNotReady", msg)
}

// IsReady returns whether the MQTTSource is ready to serve the requested
// configuration.
func (s *MQTTSourceStatus) IsReady() bool {
	return mqttCondSet.Manage(s).IsHappy()
}
