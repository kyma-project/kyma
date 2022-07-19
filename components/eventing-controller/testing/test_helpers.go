package testing

import (
	"fmt"
	"net"
	"net/http"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

const (
	ApplicationName         = "testapp1023"
	ApplicationNameNotClean = "test-app_1-0+2=3"

	EventSource                              = "/default/kyma/id"
	EventTypePrefix                          = "prefix"
	EventTypePrefixEmpty                     = ""
	OrderCreatedV1Event                      = "order.created.v1"
	OrderCreatedEventType                    = EventTypePrefix + "." + ApplicationName + "." + OrderCreatedV1Event
	OrderCreatedEventTypeNotClean            = EventTypePrefix + "." + ApplicationNameNotClean + "." + OrderCreatedV1Event
	OrderCreatedEventTypePrefixEmpty         = ApplicationName + "." + OrderCreatedV1Event
	OrderCreatedEventTypeNotCleanPrefixEmpty = ApplicationNameNotClean + "." + OrderCreatedV1Event
	EventID                                  = "8945ec08-256b-11eb-9928-acde48001122"
	EventSpecVersion                         = "1.0"
	EventData                                = "test-data"

	CloudEventType        = EventTypePrefix + "." + ApplicationName + ".order.created.v1"
	CloudEventSource      = "/default/sap.kyma/id"
	CloudEventSpecVersion = "1.0"
	CloudEventData        = "{\"foo\":\"bar\"}"

	CeIDHeader          = "ce-id"
	CeTypeHeader        = "ce-type"
	CeSourceHeader      = "ce-source"
	CeSpecVersionHeader = "ce-specversion"

	StructuredCloudEvent = `{
           "id":"` + EventID + `",
           "type":"` + OrderCreatedEventType + `",
           "specversion":"` + EventSpecVersion + `",
           "source":"` + EventSource + `",
           "data":"` + EventData + `"
        }`

	JSStreamName = "kyma"
)

type APIRuleOption func(r *apigatewayv1alpha1.APIRule)

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

// NewAPIRule returns a valid APIRule
func NewAPIRule(subscription *eventingv1alpha1.Subscription, opts ...APIRuleOption) *apigatewayv1alpha1.APIRule {
	apiRule := &apigatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "eventing.kyma-project.io/v1alpha1",
					Kind:       "subscriptions",
					Name:       subscription.Name,
					UID:        subscription.UID,
				},
			},
		},
	}

	for _, opt := range opts {
		opt(apiRule)
	}
	return apiRule
}

func WithService(name, host string) APIRuleOption {
	return func(r *apigatewayv1alpha1.APIRule) {
		port := uint32(443)
		isExternal := true
		r.Spec.Service = &apigatewayv1alpha1.Service{
			Name:       &name,
			Port:       &port,
			Host:       &host,
			IsExternal: &isExternal,
		}
	}
}

func WithPath() APIRuleOption {
	return func(r *apigatewayv1alpha1.APIRule) {
		handlerOAuth := object.OAuthHandlerName
		handler := apigatewayv1alpha1.Handler{
			Name: handlerOAuth,
		}
		authenticator := &apigatewayv1alpha1.Authenticator{
			Handler: &handler,
		}
		r.Spec.Rules = []apigatewayv1alpha1.Rule{
			{
				Path: "/path",
				Methods: []string{
					http.MethodPost,
					http.MethodOptions,
				},
				AccessStrategies: []*apigatewayv1alpha1.Authenticator{
					authenticator,
				},
			},
		}
	}
}

func MarkReady(r *apigatewayv1alpha1.APIRule) {
	statusOK := &apigatewayv1alpha1.APIRuleResourceStatus{
		Code:        apigatewayv1alpha1.StatusOK,
		Description: "",
	}

	r.Status = apigatewayv1alpha1.APIRuleStatus{
		APIRuleStatus:        statusOK,
		VirtualServiceStatus: statusOK,
		AccessRuleStatus:     statusOK,
	}
}

type ProtoOpt func(p *eventingv1alpha1.ProtocolSettings)

func NewProtocolSettings(opts ...ProtoOpt) *eventingv1alpha1.ProtocolSettings {
	protoSettings := &eventingv1alpha1.ProtocolSettings{}
	for _, o := range opts {
		o(protoSettings)
	}
	return protoSettings
}

func WithBinaryContentMode() ProtoOpt {
	return func(p *eventingv1alpha1.ProtocolSettings) {
		p.ContentMode = utils.StringPtr(eventingv1alpha1.ProtocolSettingsContentModeBinary)
	}
}

func WithExemptHandshake() ProtoOpt {
	return func(p *eventingv1alpha1.ProtocolSettings) {
		p.ExemptHandshake = func() *bool {
			exemptHandshake := true
			return &exemptHandshake
		}()
	}
}

func WithAtLeastOnceQOS() ProtoOpt {
	return func(p *eventingv1alpha1.ProtocolSettings) {
		p.Qos = utils.StringPtr(string(types.QosAtLeastOnce))
	}
}

func WithDefaultWebhookAuth() ProtoOpt {
	return func(p *eventingv1alpha1.ProtocolSettings) {
		p.WebhookAuth = &eventingv1alpha1.WebhookAuth{
			Type:         "oauth2",
			GrantType:    "client_credentials",
			ClientID:     "xxx",
			ClientSecret: "xxx",
			TokenURL:     "https://oauth2.xxx.com/oauth2/token",
			Scope:        []string{"guid-identifier"},
		}
	}
}

type SubscriptionOpt func(subscription *eventingv1alpha1.Subscription)

func NewSubscription(name, namespace string, opts ...SubscriptionOpt) *eventingv1alpha1.Subscription {
	newSub := &eventingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventingv1alpha1.SubscriptionSpec{},
	}
	for _, o := range opts {
		o(newSub)
	}
	return newSub
}

func NewBEBSubscription(name, contentMode string, webhookURL string, events types.Events, webhookAuth *types.WebhookAuth) *types.Subscription {
	return &types.Subscription{
		Name:            name,
		ContentMode:     contentMode,
		Qos:             types.QosAtLeastOnce,
		ExemptHandshake: true,
		Events:          events,
		WebhookAuth:     webhookAuth,
		WebhookURL:      webhookURL,
	}
}

func exemptHandshake(val bool) *bool {
	exemptHandshake := val
	return &exemptHandshake
}

func WithFakeSubscriptionStatus() SubscriptionOpt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Status.Conditions = []eventingv1alpha1.Condition{
			{
				Type:    "foo",
				Status:  "foo",
				Reason:  "foo-reason",
				Message: "foo-message",
			},
		}
	}
}

func WithSink(sink string) SubscriptionOpt {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.Spec.Sink = sink
	}
}
func WithConditions(conditions []eventingv1alpha1.Condition) SubscriptionOpt {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.Status.Conditions = conditions
	}
}
func WithStatus(status bool) SubscriptionOpt {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.Status.Ready = status
	}
}
func WithFinalizers(finalizers []string) SubscriptionOpt {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.ObjectMeta.Finalizers = finalizers
	}
}

func WithStatusConfig(defaultConfig env.DefaultSubscriptionConfig) SubscriptionOpt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Status.Config = eventingv1alpha1.MergeSubsConfigs(nil, &defaultConfig)
	}
}

func WithSpecConfig(defaultConfig env.DefaultSubscriptionConfig) SubscriptionOpt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Spec.Config = eventingv1alpha1.MergeSubsConfigs(nil, &defaultConfig)
	}
}

func WithStatusCleanEventTypes(cleanEventTypes []string) SubscriptionOpt {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.Status.CleanEventTypes = cleanEventTypes
	}
}

func WithEmsSubscriptionStatus(status string) SubscriptionOpt {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.Status.EmsSubscriptionStatus = &eventingv1alpha1.EmsSubscriptionStatus{
			SubscriptionStatus: status,
		}
	}
}

func WithWebhookAuthForBEB() SubscriptionOpt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Spec.Protocol = "BEB"
		s.Spec.ProtocolSettings = &eventingv1alpha1.ProtocolSettings{
			ContentMode: func() *string {
				contentMode := eventingv1alpha1.ProtocolSettingsContentModeBinary
				return &contentMode
			}(),
			ExemptHandshake: exemptHandshake(true),
			Qos:             utils.StringPtr(string(types.QosAtLeastOnce)),
			WebhookAuth: &eventingv1alpha1.WebhookAuth{
				Type:         "oauth2",
				GrantType:    "client_credentials",
				ClientID:     "xxx",
				ClientSecret: "xxx",
				TokenURL:     "https://oauth2.xxx.com/oauth2/token",
				Scope:        []string{"guid-identifier"},
			},
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

// WithWebhookForNATS is a SubscriptionOpt for creating a Subscription with a webhook set to the NATS protocol.
func WithWebhookForNATS() SubscriptionOpt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Spec.Protocol = "NATS"
		s.Spec.ProtocolSettings = &eventingv1alpha1.ProtocolSettings{}
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

// WithNotCleanFilter initializes subscription filter with a not clean event-type
// A not clean event-type means it contains none-alphanumeric characters
func WithNotCleanFilter() SubscriptionOpt {
	return WithFilter(EventSource, OrderCreatedEventTypeNotClean)
}

// WithEmptyFilter is a SubscriptionOpt for creating a subscription with an empty event type filter.
//  Note that this is different from setting Filter to nil.
func WithEmptyFilter() SubscriptionOpt {
	return func(subscription *eventingv1alpha1.Subscription) {
		subscription.Spec.Filter = &eventingv1alpha1.BEBFilters{
			Filters: []*eventingv1alpha1.BEBFilter{},
		}
	}
}

func WithOrderCreatedFilter() SubscriptionOpt {
	return WithFilter(EventSource, OrderCreatedEventType)
}

func WithSinkMissingScheme(svcNamespace, svcName string) SubscriptionOpt {
	return WithSinkURL(fmt.Sprintf("%s.%s.svc.cluster.local", svcName, svcNamespace))
}

// WithValidSink is a SubscriptionOpt for creating a subscription with a valid sink that itself gets created from
// the svcNamespace and the svcName.
func WithValidSink(svcNamespace, svcName string) SubscriptionOpt {
	return WithSinkURL(ValidSinkURL(svcNamespace, svcName))
}

// WithSinkURLFromSvcAndPath sets a kubernetes service as the sink
func WithSinkURLFromSvcAndPath(svc *corev1.Service, path string) SubscriptionOpt {
	return WithSinkURL(fmt.Sprintf("%s%s", ValidSinkURL(svc.Namespace, svc.Name), path))
}

// WithSinkURLFromSvc sets a kubernetes service as the sink
func WithSinkURLFromSvc(svc *corev1.Service) SubscriptionOpt {
	return WithSinkURL(ValidSinkURL(svc.Namespace, svc.Name))
}

// ValidSinkURL converts a namespace and service name to a valid sink url
func ValidSinkURL(namespace, svcName string) string {
	return fmt.Sprintf("https://%s.%s.svc.cluster.local", svcName, namespace)
}

// WithSinkURL is a SubscriptionOpt for creating a subscription with a specific sink.
func WithSinkURL(sinkURL string) SubscriptionOpt {
	return func(subscription *eventingv1alpha1.Subscription) { subscription.Spec.Sink = sinkURL }
}

// WithNonZeroDeletionTimestamp sets the deletion timestamp of the subscription to Now()
func WithNonZeroDeletionTimestamp() SubscriptionOpt {
	return func(subscription *eventingv1alpha1.Subscription) {
		now := metav1.Now()
		subscription.DeletionTimestamp = &now
	}
}

// SetSink sets the subscription's sink to a valid sink created from svcNameSpace and svcName.
func SetSink(svcNamespace, svcName string, subscription *eventingv1alpha1.Subscription) {
	subscription.Spec.Sink = ValidSinkURL(svcNamespace, svcName)
}

func NewSubscriberSvc(name, namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol: "TCP",
					Port:     443,
					TargetPort: intstr.IntOrString{
						IntVal: 8080,
					},
				},
			},
			Selector: map[string]string{
				"test": "test",
			},
		},
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

// WithMultipleConditions is a SubscriptionOpt for creating Subscriptions with multiple conditions.
func WithMultipleConditions() SubscriptionOpt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Status.Conditions = MultipleDefaultConditions()
	}
}

func MultipleDefaultConditions() []eventingv1alpha1.Condition {
	return []eventingv1alpha1.Condition{CustomReadyCondition("One"), CustomReadyCondition("Two")}
}

func CustomReadyCondition(msg string) eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(
		eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionTrue, msg)
}

func DefaultReadyCondition() eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(
		eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionTrue, "")
}

// ToSubscription converts an unstructured subscription into a typed one
func ToSubscription(unstructuredSub *unstructured.Unstructured) (*eventingv1alpha1.Subscription, error) {
	subscription := new(eventingv1alpha1.Subscription)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredSub.Object, subscription)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// ToUnstructuredAPIRule converts an APIRule object into a unstructured APIRule
func ToUnstructuredAPIRule(obj interface{}) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	u.Object = unstructuredObj
	return u, nil
}

// SetupSchemeOrDie add a scheme to eventing API schemes
func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

// SubscriptionGroupVersionResource returns the GVR of a subscription
func SubscriptionGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}

// NewFakeSubscriptionClient returns a fake dynamic subscription client
func NewFakeSubscriptionClient(sub *eventingv1alpha1.Subscription) (dynamic.Interface, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return nil, err
	}

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, sub)
	return dynamicClient, nil
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
