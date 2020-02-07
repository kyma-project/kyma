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

	authenticationv1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	"knative.dev/pkg/apis"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

const (
	// HTTPConditionSinkProvided has status True when the HTTPSource has
	// been configured with a sink target.
	HTTPConditionSinkProvided apis.ConditionType = "SinkProvided"

	// HTTPConditionDeployed has status True when the HTTPSource adapter
	// has been successfully deployed.
	HTTPConditionDeployed apis.ConditionType = "Deployed"

	// HTTPConditionPolicyCreated has status True when the Policy for
	// Knative service has been successfully created.
	HTTPConditionPolicyCreated apis.ConditionType = "PolicyCreated"
)

var httpCondSet = apis.NewLivingConditionSet(
	HTTPConditionSinkProvided,
	HTTPConditionDeployed,
)

// HTTPSourceGVK returns a GroupVersionKind for the HTTPSource type.
func HTTPSourceGVK() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("HTTPSource")
}

// ToOwner return a OwnerReference corresponding to the given HTTPSource.
func (s *HTTPSource) ToOwner() *metav1.OwnerReference {
	return metav1.NewControllerRef(s, HTTPSourceGVK())
}

// ToKey returns a key corresponding to the HTTPSource.
func (s *HTTPSource) ToKey() string {
	return s.Namespace + "/" + s.Name
}

const (
	HTTPSourceReasonSinkNotFound    = "SinkNotFound"
	HTTPSourceReasonSinkEmpty       = "EmptySinkURI"
	HTTPSourceReasonServiceNotReady = "ServiceNotReady"
	HTTPSourcePolicyNotCreated      = "PolicyNotCreated"
)

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *HTTPSourceStatus) InitializeConditions() {
	httpCondSet.Manage(s).InitializeConditions()
}

// MarkSink sets the SinkProvided condition to True using the given URI.
func (s *HTTPSourceStatus) MarkSink(uri string) {
	s.SinkURI = uri
	if uri == "" {
		httpCondSet.Manage(s).MarkUnknown(HTTPConditionSinkProvided,
			HTTPSourceReasonSinkEmpty, "The sink has no URI")
		return
	}
	httpCondSet.Manage(s).MarkTrue(HTTPConditionSinkProvided)
}

// MarkPolicyCreated sets the PolicyCreated condition to True once a Policy is created.
func (s *HTTPSourceStatus) MarkPolicyCreated(policy *authenticationv1alpha1.Policy) {
	if policy == nil {
		httpCondSet.Manage(s).MarkUnknown(HTTPConditionPolicyCreated,
			HTTPSourcePolicyNotCreated, "The Istio policy is not created")
		return
	}
	httpCondSet.Manage(s).MarkTrue(HTTPConditionPolicyCreated)
}

// MarkNoSink sets the SinkProvided condition to False with the given reason
// and message.
func (s *HTTPSourceStatus) MarkNoSink() {
	s.SinkURI = ""
	httpCondSet.Manage(s).MarkFalse(HTTPConditionSinkProvided,
		HTTPSourceReasonSinkNotFound, "The sink does not exist or its URL is not set")
}

// PropagateServiceReady uses the readiness of the provided Knative Service to
// determine whether the Deployed condition should be marked as true or false.
func (s *HTTPSourceStatus) PropagateServiceReady(ksvc *servingv1alpha1.Service) {
	if ksvc.Status.IsReady() {
		httpCondSet.Manage(s).MarkTrue(HTTPConditionDeployed)
		return
	}

	msg := "The adapter Service is not yet ready"
	ksvcCondReady := ksvc.Status.GetCondition(servingv1alpha1.ServiceConditionReady)
	if ksvcCondReady != nil && ksvcCondReady.Message != "" {
		msg += ": " + ksvcCondReady.Message
	}
	httpCondSet.Manage(s).MarkFalse(HTTPConditionDeployed,
		HTTPSourceReasonServiceNotReady, msg)
}

// IsReady returns whether the HTTPSource is ready to serve the requested
// configuration.
func (s *HTTPSourceStatus) IsReady() bool {
	return httpCondSet.Manage(s).IsHappy()
}
