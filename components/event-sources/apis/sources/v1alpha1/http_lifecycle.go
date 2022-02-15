package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

const (
	// HTTPConditionSinkProvided has status True when the HTTPSource has
	// been configured with a sink target.
	HTTPConditionSinkProvided apis.ConditionType = "SinkProvided"

	// HTTPConditionDeployed has status True when the HTTPSource adapter
	// has been successfully deployed.
	HTTPConditionDeployed apis.ConditionType = "Deployed"

	// HTTPConditionPeerAuthenticationCreated has status True when the PeerAuthentication for
	// Deployment has been successfully created.
	HTTPConditionPeerAuthenticationCreated apis.ConditionType = "PeerAuthenticationCreated"

	// HTTPConditionServiceCreated has status True when the Service for
	// Deployment has been successfully created.
	HTTPConditionServiceCreated apis.ConditionType = "ServiceCreated"
)

var httpCondSet = apis.NewLivingConditionSet(
	HTTPConditionSinkProvided,
	HTTPConditionDeployed,
	HTTPConditionPeerAuthenticationCreated,
	HTTPConditionServiceCreated,
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
	HTTPSourceReasonSinkNotFound                 = "SinkNotFound"
	HTTPSourceReasonSinkEmpty                    = "EmptySinkURI"
	HTTPSourceReasonServiceNotReady              = "ServiceNotReady"
	HTTPSourceReasonPeerAuthenticationNotCreated = "PeerAuthenticationNotCreated"
	HTTPSourceReasonServiceNotCreated            = "ServiceNotCreated"
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

// MarkServiceCreated sets the ServiceCreated condition to True once a Service is created.
func (s *HTTPSourceStatus) MarkServiceCreated(service *corev1.Service) {
	if service == nil {
		httpCondSet.Manage(s).MarkUnknown(HTTPConditionServiceCreated,
			HTTPSourceReasonServiceNotCreated, "The Service is not created")
		return
	}
	httpCondSet.Manage(s).MarkTrue(HTTPConditionServiceCreated)
}

// MarkPeerAuthenticationCreated sets the PeerAuthenticationCreated condition to True once a PeerAuthentication is created.
func (s *HTTPSourceStatus) MarkPeerAuthenticationCreated(peerAuthentication *securityv1beta1.PeerAuthentication) {
	if peerAuthentication == nil {
		httpCondSet.Manage(s).MarkUnknown(HTTPConditionPeerAuthenticationCreated,
			HTTPSourceReasonPeerAuthenticationNotCreated, "The Istio PeerAuthentication is not created")
		return
	}
	httpCondSet.Manage(s).MarkTrue(HTTPConditionPeerAuthenticationCreated)
}

// MarkNoSink sets the SinkProvided condition to False with the given reason
// and message.
func (s *HTTPSourceStatus) MarkNoSink() {
	s.SinkURI = ""
	httpCondSet.Manage(s).MarkFalse(HTTPConditionSinkProvided,
		HTTPSourceReasonSinkNotFound, "The sink does not exist or its URL is not set")
}

// PropagateDeploymentReady uses the readiness of the provided Deployment to
// determine whether the Deployed condition should be marked as true or false.
func (s *HTTPSourceStatus) PropagateDeploymentReady(deployment *appsv1.Deployment) {
	if deployment.Status.AvailableReplicas > 0 {
		httpCondSet.Manage(s).MarkTrue(HTTPConditionDeployed)
		return
	}

	msg := "The adapter Service is not yet ready"
	httpCondSet.Manage(s).MarkFalse(HTTPConditionDeployed,
		HTTPSourceReasonServiceNotReady, msg)
}

// IsReady returns whether the HTTPSource is ready to serve the requested
// configuration.
func (s *HTTPSourceStatus) IsReady() bool {
	return httpCondSet.Manage(s).IsHappy()
}
