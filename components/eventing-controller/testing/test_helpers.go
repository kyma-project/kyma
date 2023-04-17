package testing

import (
	"net"
	"net/http"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ApplicationName = "testapp1023"

	EventTypePrefix = "prefix"
	EventID         = "8945ec08-256b-11eb-9928-acde48001122"

	CloudEventType        = EventTypePrefix + "." + ApplicationName + ".order.created.v1"
	CloudEventSource      = "/default/sap.kyma/id"
	CloudEventSpecVersion = "1.0"

	CeIDHeader          = "ce-id"
	CeTypeHeader        = "ce-type"
	CeSourceHeader      = "ce-source"
	CeSpecVersionHeader = "ce-specversion"

	JSStreamName = "kyma"
)

type APIRuleOption func(r *apigatewayv1beta1.APIRule)

// GetFreePort determines a free port on the host. It does so by delegating the job to net.ListenTCP.
// Then providing a port of 0 to net.ListenTCP, it will automatically choose a port for us.
func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			port := l.Addr().(*net.TCPAddr).Port
			err = l.Close()
			return port, err
		}
	}
	return
}

type ProtoOpt func(p *eventingv1alpha1.ProtocolSettings)

func NewProtocolSettings(opts ...ProtoOpt) *eventingv1alpha1.ProtocolSettings {
	protoSettings := &eventingv1alpha1.ProtocolSettings{}
	for _, o := range opts {
		o(protoSettings)
	}
	return protoSettings
}

func WithAtLeastOnceQOS() ProtoOpt {
	return func(p *eventingv1alpha1.ProtocolSettings) {
		p.Qos = utils.StringPtr(string(types.QosAtLeastOnce))
	}
}

func WithRequiredWebhookAuth() ProtoOpt {
	return func(p *eventingv1alpha1.ProtocolSettings) {
		p.WebhookAuth = &eventingv1alpha1.WebhookAuth{
			GrantType:    "client_credentials",
			ClientID:     "xxx",
			ClientSecret: "xxx",
			TokenURL:     "https://oauth2.xxx.com/oauth2/token",
		}
	}
}

type SubscriptionOpt func(subscription *eventingv1alpha1.Subscription)

func WithStatusCleanEventTypes(cleanEventTypes []string) SubscriptionOpt {
	return func(sub *eventingv1alpha1.Subscription) {
		if cleanEventTypes == nil {
			sub.Status.InitializeCleanEventTypes()
		} else {
			sub.Status.CleanEventTypes = cleanEventTypes
		}
	}
}

func WithProtocolBEB() SubscriptionOpt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Spec.Protocol = "BEB"
	}
}

func WithProtocolSettings(p *eventingv1alpha1.ProtocolSettings) SubscriptionOpt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Spec.ProtocolSettings = p
	}
}

// AddFilter creates a new Filter from eventSource and eventType and adds it to the subscription.
func AddFilter(eventSource, eventType string, subscription *eventingv1alpha1.Subscription) {
	if subscription.Spec.Filter == nil {
		subscription.Spec.Filter = &eventingv1alpha1.BEBFilters{
			Filters: []*eventingv1alpha1.BEBFilter{},
		}
	}

	filter := &eventingv1alpha1.BEBFilter{
		EventSource: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "source",
			Value:    eventSource,
		},
		EventType: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "type",
			Value:    eventType,
		},
	}

	subscription.Spec.Filter.Filters = append(subscription.Spec.Filter.Filters, filter)
}

// WithFilter is a SubscriptionOpt for creating a Subscription with a specific event type filter,
// that itself gets created from the passed eventSource and eventType.
func WithFilter(eventSource, eventType string) SubscriptionOpt {
	return func(subscription *eventingv1alpha1.Subscription) { AddFilter(eventSource, eventType, subscription) }
}

// WithEmptyFilter is a SubscriptionOpt for creating a subscription with an empty event type filter.
// Note that this is different from setting Filter to nil.
func WithEmptyFilter() SubscriptionOpt {
	return func(subscription *eventingv1alpha1.Subscription) {
		subscription.Spec.Filter = &eventingv1alpha1.BEBFilters{
			Filters: []*eventingv1alpha1.BEBFilter{},
		}
	}
}

func WithEmptyStatus() SubscriptionOpt {
	return func(subscription *eventingv1alpha1.Subscription) {
		subscription.Status = eventingv1alpha1.SubscriptionStatus{
			CleanEventTypes: []string{},
		}
	}
}

func WithEmptyConfig() SubscriptionOpt {
	return func(subscription *eventingv1alpha1.Subscription) {
		subscription.Spec.Config = nil
	}
}

func NewBEBMessagingSecret(name, namespace string) *corev1.Secret {
	messagingValue := `
				[{
					"broker": {
						"type": "sapmgw"
					},
					"oa2": {
						"clientid": "clientid",
						"clientsecret": "clientsecret",
						"granttype": "client_credentials",
						"tokenendpoint": "https://token"
					},
					"protocol": ["amqp10ws"],
					"uri": "wss://amqp"
				}, {
					"broker": {
						"type": "sapmgw"
					},
					"oa2": {
						"clientid": "clientid",
						"clientsecret": "clientsecret",
						"granttype": "client_credentials",
						"tokenendpoint": "https://token"
					},
					"protocol": ["amqp10ws"],
					"uri": "wss://amqp"
				}, {
					"broker": {
						"type": "saprestmgw"
					},
					"oa2": {
						"clientid": "rest-clientid",
						"clientsecret": "rest-client-secret",
						"granttype": "client_credentials",
						"tokenendpoint": "https://rest-token"
					},
					"protocol": ["httprest"],
					"uri": "https://rest-messaging"
				}]`

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"messaging": messagingValue,
			"namespace": "test/ns",
		},
	}
}

func NewNamespace(name string) *corev1.Namespace {
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

func NewEventingBackend(name, namespace string) *eventingv1alpha1.EventingBackend {
	return &eventingv1alpha1.EventingBackend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec:   eventingv1alpha1.EventingBackendSpec{},
		Status: eventingv1alpha1.EventingBackendStatus{},
	}
}

func NewEventingControllerDeployment() *appsv1.Deployment {
	labels := map[string]string{
		"app.kubernetes.io/name": "value",
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.ControllerName,
			Namespace: deployment.ControllerNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: metav1.SetAsLabelSelector(labels),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   deployment.ControllerName,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  deployment.ControllerName,
							Image: "eventing-controller-pod-image",
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{},
	}
}

// ToSubscription converts an unstructured subscription into a typed one.
func ToSubscription(unstructuredSub *unstructured.Unstructured) (*eventingv1alpha1.Subscription, error) {
	sub := new(eventingv1alpha1.Subscription)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredSub.Object, sub)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// SubscriptionGroupVersionResource returns the GVR of a subscription.
func SubscriptionGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}

func GetStructuredMessageHeaders() http.Header {
	return http.Header{"Content-Type": []string{"application/cloudevents+json"}}
}

func GetBinaryMessageHeaders() http.Header {
	headers := make(http.Header)
	headers.Add(CeIDHeader, EventID)
	headers.Add(CeTypeHeader, CloudEventType)
	headers.Add(CeSourceHeader, CloudEventSource)
	headers.Add(CeSpecVersionHeader, CloudEventSpecVersion)
	return headers
}

func PublisherProxyDefaultReadyCondition() eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionPublisherProxyReady,
		eventingv1alpha1.ConditionReasonPublisherDeploymentReady,
		corev1.ConditionTrue, "")
}

func PublisherProxyDefaultNotReadyCondition() eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionPublisherProxyReady,
		eventingv1alpha1.ConditionReasonPublisherDeploymentNotReady,
		corev1.ConditionFalse, "")
}

func SubscriptionControllerDefaultReadyCondition() eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionControllerReady,
		eventingv1alpha1.ConditionReasonSubscriptionControllerReady,
		corev1.ConditionTrue, "")
}

func SubscriptionControllerReadyConditionWith(ready corev1.ConditionStatus, reason eventingv1alpha1.ConditionReason) eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionControllerReady, reason, ready, "")
}

func SubscriptionControllerReadyEvent() corev1.Event {
	return corev1.Event{
		Reason: string(eventingv1alpha1.ConditionReasonSubscriptionControllerReady),
		Type:   corev1.EventTypeNormal,
	}
}

func SubscriptionControllerNotReadyEvent() corev1.Event {
	return corev1.Event{
		Reason: string(eventingv1alpha1.ConditionReasonSubscriptionControllerNotReady),
		Type:   corev1.EventTypeWarning,
	}
}

func PublisherDeploymentReadyEvent() corev1.Event {
	return corev1.Event{
		Reason: string(eventingv1alpha1.ConditionReasonPublisherDeploymentReady),
		Type:   corev1.EventTypeNormal,
	}
}

func PublisherDeploymentNotReadyEvent() corev1.Event {
	return corev1.Event{
		Reason: string(eventingv1alpha1.ConditionReasonPublisherDeploymentNotReady),
		Type:   corev1.EventTypeWarning,
	}
}
