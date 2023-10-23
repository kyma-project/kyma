package testing

import (
	"fmt"
	"net"
	"net/http"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"

	apigatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

const (
	ApplicationName         = "testapp1023"
	ApplicationNameNotClean = "test-app_1-0+2=3"

	OrderCreatedUncleanEvent = "order.cre-ä+t*ed.v2"
	OrderCreatedCleanEvent   = "order.cre-ä+ted.v2"
	EventSourceUnclean       = "s>o>*u*r>c.e"
	EventSourceClean         = "source"

	EventMeshProtocol = "BEB"

	EventMeshNamespaceNS        = "/default/ns"
	EventMeshNamespace          = "/default/kyma/id"
	EventSource                 = "/default/kyma/id"
	EventTypePrefix             = "prefix"
	EventMeshPrefix             = "one.two.three"      // three segments
	InvalidEventMeshPrefix      = "one.two.three.four" // four segments
	EventTypePrefixEmpty        = ""
	OrderCreatedV1Event         = "order.created.v1"
	OrderCreatedV2Event         = "order.created.v2"
	OrderCreatedV1EventNotClean = "order.c*r%e&a!te#d.v1"
	JetStreamSubject            = "kyma" + "." + EventSourceClean + "." + OrderCreatedV1Event
	JetStreamSubjectV2          = "kyma" + "." + EventSourceClean + "." + OrderCreatedCleanEvent

	EventMeshExactType          = EventMeshPrefix + "." + ApplicationNameNotClean + "." + OrderCreatedV1EventNotClean
	EventMeshOrderCreatedV1Type = EventMeshPrefix + "." + ApplicationName + "." + OrderCreatedV1Event

	OrderCreatedEventType            = EventTypePrefix + "." + ApplicationName + "." + OrderCreatedV1Event
	OrderCreatedEventTypeNotClean    = EventTypePrefix + "." + ApplicationNameNotClean + "." + OrderCreatedV1Event
	OrderCreatedEventTypePrefixEmpty = ApplicationName + "." + OrderCreatedV1Event

	CloudEventType  = EventTypePrefix + "." + ApplicationName + ".order.created.v1"
	CloudEventData  = "{\"foo\":\"bar\"}"
	CloudEventData2 = "{\"foo\":\"bar2\"}"

	JSStreamName = "kyma"

	EventID = "8945ec08-256b-11eb-9928-acde48001122"

	CloudEventSource      = "/default/sap.kyma/id"
	CloudEventSpecVersion = "1.0"

	CeIDHeader          = "ce-id"
	CeTypeHeader        = "ce-type"
	CeSourceHeader      = "ce-source"
	CeSpecVersionHeader = "ce-specversion"
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

type SubscriptionV1alpha1Opt func(subscription *eventingv1alpha1.Subscription)

func WithStatusCleanEventTypes(cleanEventTypes []string) SubscriptionV1alpha1Opt {
	return func(sub *eventingv1alpha1.Subscription) {
		if cleanEventTypes == nil {
			sub.Status.InitializeCleanEventTypes()
		} else {
			sub.Status.CleanEventTypes = cleanEventTypes
		}
	}
}

func WithV1alpha1ProtocolEventMesh() SubscriptionV1alpha1Opt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Spec.Protocol = EventMeshProtocol
	}
}

func WithV1alpha1ProtocolSettings(p *eventingv1alpha1.ProtocolSettings) SubscriptionV1alpha1Opt {
	return func(s *eventingv1alpha1.Subscription) {
		s.Spec.ProtocolSettings = p
	}
}

// AddV1alpha1Filter creates a new Filter from eventSource and eventType and adds it to the subscription.
func AddV1alpha1Filter(eventSource, eventType string, subscription *eventingv1alpha1.Subscription) {
	if subscription.Spec.Filter == nil {
		subscription.Spec.Filter = &eventingv1alpha1.BEBFilters{
			Filters: []*eventingv1alpha1.EventMeshFilter{},
		}
	}

	filter := &eventingv1alpha1.EventMeshFilter{
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

// WithV1alpha1Filter is a SubscriptionOpt for creating a Subscription with a specific event type filter,
// that itself gets created from the passed eventSource and eventType.
func WithV1alpha1Filter(eventSource, eventType string) SubscriptionV1alpha1Opt {
	return func(subscription *eventingv1alpha1.Subscription) {
		AddV1alpha1Filter(eventSource, eventType, subscription)
	}
}

// WithV1alpha1EmptyFilter is a SubscriptionOpt for creating a subscription with an empty event type filter.
// Note that this is different from setting Filter to nil.
func WithV1alpha1EmptyFilter() SubscriptionV1alpha1Opt {
	return func(subscription *eventingv1alpha1.Subscription) {
		subscription.Spec.Filter = &eventingv1alpha1.BEBFilters{
			Filters: []*eventingv1alpha1.EventMeshFilter{},
		}
	}
}

func WithV1alpha1EmptyStatus() SubscriptionV1alpha1Opt {
	return func(subscription *eventingv1alpha1.Subscription) {
		subscription.Status = eventingv1alpha1.SubscriptionStatus{
			CleanEventTypes: []string{},
		}
	}
}

func WithV1alpha1EmptyConfig() SubscriptionV1alpha1Opt {
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

func SubscriptionControllerReadyConditionWith(ready corev1.ConditionStatus,
	reason eventingv1alpha1.ConditionReason) eventingv1alpha1.Condition {
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

// NewAPIRule returns a valid APIRule.
func NewAPIRule(subscription *eventingv1alpha2.Subscription, opts ...APIRuleOption) *apigatewayv1beta1.APIRule {
	apiRule := &apigatewayv1beta1.APIRule{
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
	return func(r *apigatewayv1beta1.APIRule) {
		port := uint32(443) //nolint:gomnd // tests
		isExternal := true
		r.Spec.Host = &host
		r.Spec.Service = &apigatewayv1beta1.Service{
			Name:       &name,
			Port:       &port,
			IsExternal: &isExternal,
		}
	}
}

func WithPath() APIRuleOption {
	return func(r *apigatewayv1beta1.APIRule) {
		handlerOAuth := object.OAuthHandlerNameOAuth2Introspection
		handler := apigatewayv1beta1.Handler{
			Name: handlerOAuth,
		}
		authenticator := &apigatewayv1beta1.Authenticator{
			Handler: &handler,
		}
		r.Spec.Rules = []apigatewayv1beta1.Rule{
			{
				Path: "/path",
				Methods: []string{
					http.MethodPost,
					http.MethodOptions,
				},
				AccessStrategies: []*apigatewayv1beta1.Authenticator{
					authenticator,
				},
			},
		}
	}
}

func MarkReady(r *apigatewayv1beta1.APIRule) {
	statusOK := &apigatewayv1beta1.APIRuleResourceStatus{
		Code:        apigatewayv1beta1.StatusOK,
		Description: "",
	}

	r.Status = apigatewayv1beta1.APIRuleStatus{
		APIRuleStatus:        statusOK,
		VirtualServiceStatus: statusOK,
		AccessRuleStatus:     statusOK,
	}
}

type SubscriptionOpt func(subscription *eventingv1alpha2.Subscription)

func NewSubscription(name, namespace string, opts ...SubscriptionOpt) *eventingv1alpha2.Subscription {
	newSub := &eventingv1alpha2.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventingv1alpha2.SubscriptionSpec{
			Config: map[string]string{},
		},
	}
	for _, o := range opts {
		o(newSub)
	}
	return newSub
}

func NewEventMeshSubscription(name, contentMode string, webhookURL string, events types.Events,
	webhookAuth *types.WebhookAuth) *types.Subscription {
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

func NewSampleEventMeshSubscription() *types.Subscription {
	eventType := []types.Event{
		{
			Source: EventSource,
			Type:   OrderCreatedEventTypeNotClean,
		},
	}

	return NewEventMeshSubscription("ev2subs1", types.ContentModeStructured, "https://webhook.xxx.com",
		eventType, nil)
}

func WithFakeSubscriptionStatus() SubscriptionOpt {
	return func(s *eventingv1alpha2.Subscription) {
		s.Status.Conditions = []eventingv1alpha2.Condition{
			{
				Type:    "foo",
				Status:  "foo",
				Reason:  "foo-reason",
				Message: "foo-message",
			},
		}
	}
}

func WithSource(source string) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.Spec.Source = source
	}
}

func WithTypes(types []string) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.Spec.Types = types
	}
}

func WithSink(sink string) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.Spec.Sink = sink
	}
}
func WithConditions(conditions []eventingv1alpha2.Condition) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.Status.Conditions = conditions
	}
}
func WithStatus(status bool) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.Status.Ready = status
	}
}
func WithFinalizers(finalizers []string) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.ObjectMeta.Finalizers = finalizers
	}
}

func WithStatusTypes(cleanEventTypes []eventingv1alpha2.EventType) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		if cleanEventTypes == nil {
			sub.Status.InitializeEventTypes()
		} else {
			sub.Status.Types = cleanEventTypes
		}
	}
}

func WithStatusJSBackendTypes(types []eventingv1alpha2.JetStreamTypes) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.Status.Backend.Types = types
	}
}

func WithEmsSubscriptionStatus(status string) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.Status.Backend.EventMeshSubscriptionStatus = &eventingv1alpha2.EventMeshSubscriptionStatus{
			Status: status,
		}
	}
}

func WithWebhookAuthForEventMesh() SubscriptionOpt {
	return func(s *eventingv1alpha2.Subscription) {
		s.Spec.Config = map[string]string{
			eventingv1alpha2.Protocol:                        EventMeshProtocol,
			eventingv1alpha2.ProtocolSettingsContentMode:     "BINARY",
			eventingv1alpha2.ProtocolSettingsExemptHandshake: "true",
			eventingv1alpha2.ProtocolSettingsQos:             "AT_LEAST_ONCE",
			eventingv1alpha2.WebhookAuthType:                 "oauth2",
			eventingv1alpha2.WebhookAuthGrantType:            "client_credentials",
			eventingv1alpha2.WebhookAuthClientID:             "xxx",
			eventingv1alpha2.WebhookAuthClientSecret:         "xxx",
			eventingv1alpha2.WebhookAuthTokenURL:             "https://oauth2.xxx.com/oauth2/token",
			eventingv1alpha2.WebhookAuthScope:                "guid-identifier,root",
		}
	}
}

func WithInvalidProtocolSettingsQos() SubscriptionOpt {
	return func(s *eventingv1alpha2.Subscription) {
		if s.Spec.Config == nil {
			s.Spec.Config = map[string]string{}
		}
		s.Spec.Config[eventingv1alpha2.ProtocolSettingsQos] = "AT_INVALID_ONCE"
	}
}

func WithInvalidWebhookAuthType() SubscriptionOpt {
	return func(s *eventingv1alpha2.Subscription) {
		if s.Spec.Config == nil {
			s.Spec.Config = map[string]string{}
		}
		s.Spec.Config[eventingv1alpha2.WebhookAuthType] = "abcd"
	}
}

func WithInvalidWebhookAuthGrantType() SubscriptionOpt {
	return func(s *eventingv1alpha2.Subscription) {
		if s.Spec.Config == nil {
			s.Spec.Config = map[string]string{}
		}
		s.Spec.Config[eventingv1alpha2.WebhookAuthGrantType] = "invalid"
	}
}

func WithProtocolEventMesh() SubscriptionOpt {
	return func(s *eventingv1alpha2.Subscription) {
		if s.Spec.Config == nil {
			s.Spec.Config = map[string]string{}
		}
		s.Spec.Config[eventingv1alpha2.Protocol] = EventMeshProtocol
	}
}

// AddEventType adds a new type to the subscription.
func AddEventType(eventType string, subscription *eventingv1alpha2.Subscription) {
	subscription.Spec.Types = append(subscription.Spec.Types, eventType)
}

// WithEventType is a SubscriptionOpt for creating a Subscription with a specific event type,
// that itself gets created from the passed eventType.
func WithEventType(eventType string) SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) { AddEventType(eventType, subscription) }
}

// WithEventSource is a SubscriptionOpt for creating a Subscription with a specific event source,.
func WithEventSource(source string) SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) { subscription.Spec.Source = source }
}

// WithExactTypeMatching is a SubscriptionOpt for creating a Subscription with an exact type matching.
func WithExactTypeMatching() SubscriptionOpt {
	return WithTypeMatching(eventingv1alpha2.TypeMatchingExact)
}

// WithStandardTypeMatching is a SubscriptionOpt for creating a Subscription with a standard type matching.
func WithStandardTypeMatching() SubscriptionOpt {
	return WithTypeMatching(eventingv1alpha2.TypeMatchingStandard)
}

// WithTypeMatching is a SubscriptionOpt for creating a Subscription with a specific type matching,.
func WithTypeMatching(typeMatching eventingv1alpha2.TypeMatching) SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) { subscription.Spec.TypeMatching = typeMatching }
}

// WithNotCleanType initializes subscription with a not clean event-type
// A not clean event-type means it contains none-alphanumeric characters.
func WithNotCleanType() SubscriptionOpt {
	return WithEventType(OrderCreatedV1EventNotClean)
}

func WithEmptyStatus() SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		subscription.Status = eventingv1alpha2.SubscriptionStatus{}
	}
}

func WithEmptyConfig() SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		subscription.Spec.Config = map[string]string{}
	}
}

func WithConfigValue(key, value string) SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		if subscription.Spec.Config == nil {
			subscription.Spec.Config = map[string]string{}
		}
		subscription.Spec.Config[key] = value
	}
}

func WithOrderCreatedFilter() SubscriptionOpt {
	return WithEventType(OrderCreatedEventType)
}

func WithEventMeshExactType() SubscriptionOpt {
	return WithEventType(EventMeshExactType)
}

func WithOrderCreatedV1Event() SubscriptionOpt {
	return WithEventType(OrderCreatedV1Event)
}

func WithDefaultSource() SubscriptionOpt {
	return WithEventSource(ApplicationName)
}

func WithNotCleanSource() SubscriptionOpt {
	return WithEventSource(ApplicationNameNotClean)
}

// WithValidSink is a SubscriptionOpt for creating a subscription with a valid sink that itself gets created from
// the svcNamespace and the svcName.
func WithValidSink(svcNamespace, svcName string) SubscriptionOpt {
	return WithSinkURL(ValidSinkURL(svcNamespace, svcName))
}

// WithSinkURLFromSvc sets a kubernetes service as the sink.
func WithSinkURLFromSvc(svc *corev1.Service) SubscriptionOpt {
	return WithSinkURL(ValidSinkURL(svc.Namespace, svc.Name))
}

// ValidSinkURL converts a namespace and service name to a valid sink url.
func ValidSinkURL(namespace, svcName string) string {
	return fmt.Sprintf("https://%s.%s.svc.cluster.local", svcName, namespace)
}

// ValidSinkURLWithPath converts a namespace and service name to a valid sink url with path.
func ValidSinkURLWithPath(namespace, svcName, path string) string {
	return fmt.Sprintf("https://%s.%s.svc.cluster.local/%s", svcName, namespace, path)
}

// WithSinkURL is a SubscriptionOpt for creating a subscription with a specific sink.
func WithSinkURL(sinkURL string) SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) { subscription.Spec.Sink = sinkURL }
}

// WithNonZeroDeletionTimestamp sets the deletion timestamp of the subscription to Now().
func WithNonZeroDeletionTimestamp() SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		now := metav1.Now()
		subscription.DeletionTimestamp = &now
	}
}

// SetSink sets the subscription's sink to a valid sink created from svcNameSpace and svcName.
func SetSink(svcNamespace, svcName string, subscription *eventingv1alpha2.Subscription) {
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
					Port:     443, //nolint:gomnd // tests
					TargetPort: intstr.IntOrString{
						IntVal: 8080, //nolint:gomnd // tests
					},
				},
			},
			Selector: map[string]string{
				"test": "test",
			},
		},
	}
}

// ToSubscription converts an unstructured subscription into a typed one.
func ToSubscription(unstructuredSub *unstructured.Unstructured) (*eventingv1alpha2.Subscription, error) {
	sub := new(eventingv1alpha2.Subscription)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredSub.Object, sub)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// ToUnstructuredAPIRule converts an APIRule object into a unstructured APIRule.
func ToUnstructuredAPIRule(obj interface{}) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	u.Object = unstructuredObj
	return u, nil
}

// SetupSchemeOrDie add a scheme to eventing API schemes.
func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := eventingv1alpha2.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

// SubscriptionGroupVersionResource returns the GVR of a subscription.
func SubscriptionGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha2.GroupVersion.Version,
		Group:    eventingv1alpha2.GroupVersion.Group,
		Resource: "subscriptions",
	}
}

// NewFakeSubscriptionClient returns a fake dynamic subscription client.
func NewFakeSubscriptionClient(sub *eventingv1alpha2.Subscription) (dynamic.Interface, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return nil, err
	}

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, sub)
	return dynamicClient, nil
}

// AddSource adds the source value to the subscription.
func AddSource(source string, subscription *eventingv1alpha2.Subscription) {
	subscription.Spec.Source = source
}

// WithSourceAndType is a SubscriptionOpt for creating a Subscription with a specific eventSource and eventType.
func WithSourceAndType(eventSource, eventType string) SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		AddSource(eventSource, subscription)
		AddEventType(eventType, subscription)
	}
}

// WithCleanEventTypeOld is a SubscriptionOpt that initializes subscription with a not clean event type from v1alpha2.
func WithCleanEventTypeOld() SubscriptionOpt {
	return WithSourceAndType(EventSourceClean, OrderCreatedEventType)
}

// WithCleanEventSourceAndType is a SubscriptionOpt that initializes subscription with a not clean event source and
// type.
func WithCleanEventSourceAndType() SubscriptionOpt {
	return WithSourceAndType(EventSourceClean, OrderCreatedV1Event)
}

// WithNotCleanEventSourceAndType is a SubscriptionOpt that initializes subscription with a not clean event source
// and type.
func WithNotCleanEventSourceAndType() SubscriptionOpt {
	return WithSourceAndType(EventSourceUnclean, OrderCreatedUncleanEvent)
}

// WithTypeMatchingStandard is a SubscriptionOpt that initializes the subscription with type matching to standard.
func WithTypeMatchingStandard() SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		subscription.Spec.TypeMatching = eventingv1alpha2.TypeMatchingStandard
	}
}

// WithTypeMatchingExact is a SubscriptionOpt that initializes the subscription with type matching to exact.
func WithTypeMatchingExact() SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		subscription.Spec.TypeMatching = eventingv1alpha2.TypeMatchingExact
	}
}

// WithMaxInFlight is a SubscriptionOpt that sets the status with the maxInFlightMessages int value.
func WithMaxInFlight(maxInFlight int) SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		subscription.Spec.Config = map[string]string{
			eventingv1alpha2.MaxInFlightMessages: fmt.Sprint(maxInFlight),
		}
	}
}

// WithMaxInFlightMessages is a SubscriptionOpt that sets the status with the maxInFlightMessages string value.
func WithMaxInFlightMessages(maxInFlight string) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		if sub.Spec.Config == nil {
			sub.Spec.Config = map[string]string{}
		}
		sub.Spec.Config[eventingv1alpha2.MaxInFlightMessages] = maxInFlight
	}
}

// WithBackend is a SubscriptionOpt that sets the status with the Backend value.
func WithBackend(backend eventingv1alpha2.Backend) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.Status.Backend = backend
	}
}
