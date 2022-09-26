package v2

import (
	"fmt"
	"net"
	"net/http"
	"time"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/kyma-project/kyma/components/eventing-controller/testing/event/cehelper"
)

const (
	ApplicationName         = "testapp1023"
	ApplicationNameNotClean = "test-app_1-0+2=3"

	OrderCreatedUncleanEvent = "order.cre-ä+t*ed.v2"
	OrderCreatedCleanEvent   = "order.cre-ä+ted.v2"
	EventSourceUnclean       = "s>o>*u*r>c.e"
	EventSourceClean         = "source"

	EventMeshNamespace                       = "/default/kyma/id"
	EventSource                              = "/default/kyma/id"
	EventTypePrefix                          = "prefix"
	EventTypePrefixEmpty                     = ""
	OrderCreatedV1Event                      = "order.created.v1"
	OrderCreatedV2Event                      = "order.created.v2"
	OrderCreatedEventType                    = EventTypePrefix + "." + ApplicationName + "." + OrderCreatedV1Event
	NewOrderCreatedEventType                 = EventTypePrefix + "." + ApplicationName + "." + OrderCreatedV2Event
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

// CloudEvent returns the pointer to a simple CloudEvent for testing purpose.
func CloudEvent() (*ce.Event, error) {
	return cehelper.NewEvent(
		cehelper.WithSubject(CloudEventType),
		cehelper.WithSpecVersion(CloudEventSpecVersion),
		cehelper.WithID(CloudEventSpecVersion),
		cehelper.WithSource(CloudEventSource),
		cehelper.WithType(CloudEventType),
		cehelper.WithData(ce.ApplicationJSON, CloudEventData),
		cehelper.WithTime(time.Now()),
	)
}

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
		port := uint32(443)
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
		handlerOAuth := object.OAuthHandlerName
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

type ProtoOpt func(p *eventingv1alpha2.ProtocolSettings)

func NewProtocolSettings(opts ...ProtoOpt) *eventingv1alpha2.ProtocolSettings {
	protoSettings := &eventingv1alpha2.ProtocolSettings{}
	for _, o := range opts {
		o(protoSettings)
	}
	return protoSettings
}

func WithBinaryContentMode() ProtoOpt {
	return func(p *eventingv1alpha2.ProtocolSettings) {
		p.ContentMode = utils.StringPtr(eventingv1alpha2.ProtocolSettingsContentModeBinary)
	}
}

func WithExemptHandshake() ProtoOpt {
	return func(p *eventingv1alpha2.ProtocolSettings) {
		p.ExemptHandshake = func() *bool {
			exemptHandshake := true
			return &exemptHandshake
		}()
	}
}

func WithAtLeastOnceQOS() ProtoOpt {
	return func(p *eventingv1alpha2.ProtocolSettings) {
		p.Qos = utils.StringPtr(string(types.QosAtLeastOnce))
	}
}

func WithDefaultWebhookAuth() ProtoOpt {
	return func(p *eventingv1alpha2.ProtocolSettings) {
		p.WebhookAuth = &eventingv1alpha2.WebhookAuth{
			Type:         "oauth2",
			GrantType:    "client_credentials",
			ClientID:     "xxx",
			ClientSecret: "xxx",
			TokenURL:     "https://oauth2.xxx.com/oauth2/token",
			Scope:        []string{"guid-identifier"},
		}
	}
}

type SubscriptionOpt func(subscription *eventingv1alpha2.Subscription)

func NewSubscription(name, namespace string, opts ...SubscriptionOpt) *eventingv1alpha2.Subscription {
	newSub := &eventingv1alpha2.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventingv1alpha2.SubscriptionSpec{},
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
			sub.Status.InitializeCleanEventTypes()
		} else {
			sub.Status.Types = cleanEventTypes
		}
	}
}

func WithEmsSubscriptionStatus(status string) SubscriptionOpt {
	return func(sub *eventingv1alpha2.Subscription) {
		sub.Status.Backend.EmsSubscriptionStatus = &eventingv1alpha2.EmsSubscriptionStatus{
			Status: status,
		}
	}
}

func WithWebhookAuthForBEB() SubscriptionOpt {
	return func(s *eventingv1alpha2.Subscription) {
		s.Spec.Protocol = "BEB"
		s.Spec.ProtocolSettings = &eventingv1alpha2.ProtocolSettings{
			ContentMode: func() *string {
				contentMode := eventingv1alpha2.ProtocolSettingsContentModeBinary
				return &contentMode
			}(),
			ExemptHandshake: exemptHandshake(true),
			Qos:             utils.StringPtr(string(types.QosAtLeastOnce)),
			WebhookAuth: &eventingv1alpha2.WebhookAuth{
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
	return func(s *eventingv1alpha2.Subscription) {
		s.Spec.Protocol = "BEB"
	}
}

func WithProtocolSettings(p *eventingv1alpha2.ProtocolSettings) SubscriptionOpt {
	return func(s *eventingv1alpha2.Subscription) {
		s.Spec.ProtocolSettings = p
	}
}

// WithWebhookForNATS is a SubscriptionOpt for creating a Subscription with a webhook set to the NATS protocol.
func WithWebhookForNATS() SubscriptionOpt {
	return func(s *eventingv1alpha2.Subscription) {
		s.Spec.Protocol = "NATS"
		s.Spec.ProtocolSettings = &eventingv1alpha2.ProtocolSettings{}
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

// WithEventSource is a SubscriptionOpt for creating a Subscription with a specific event source,
func WithEventSource(source string) SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) { subscription.Spec.Source = source }
}

// WithNotCleanFilter initializes subscription filter with a not clean event-type
// A not clean event-type means it contains none-alphanumeric characters.
func WithNotCleanFilter() SubscriptionOpt {
	return WithEventType(OrderCreatedEventTypeNotClean)
}

// WithEmptyTypes is a SubscriptionOpt for creating a subscription with an empty event type filter.
// Note that this is different from setting Types to nil.
func WithEmptyTypes() SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		subscription.Spec.Types = make([]string, 0)
	}
}

func WithOrderCreatedFilter() SubscriptionOpt {
	return WithEventType(OrderCreatedEventType)
}

func WithSinkMissingScheme(svcNamespace, svcName string) SubscriptionOpt {
	return WithSinkURL(fmt.Sprintf("%s.%s.svc.cluster.local", svcName, svcNamespace))
}

func WithDefaultSource() SubscriptionOpt {
	return WithEventSource(ApplicationName)
}

// WithValidSink is a SubscriptionOpt for creating a subscription with a valid sink that itself gets created from
// the svcNamespace and the svcName.
func WithValidSink(svcNamespace, svcName string) SubscriptionOpt {
	return WithSinkURL(ValidSinkURL(svcNamespace, svcName))
}

// WithSinkURLFromSvcAndPath sets a kubernetes service as the sink.
func WithSinkURLFromSvcAndPath(svc *corev1.Service, path string) SubscriptionOpt {
	return WithSinkURL(fmt.Sprintf("%s%s", ValidSinkURL(svc.Namespace, svc.Name), path))
}

// WithSinkURLFromSvc sets a kubernetes service as the sink.
func WithSinkURLFromSvc(svc *corev1.Service) SubscriptionOpt {
	return WithSinkURL(ValidSinkURL(svc.Namespace, svc.Name))
}

// ValidSinkURL converts a namespace and service name to a valid sink url.
func ValidSinkURL(namespace, svcName string) string {
	return fmt.Sprintf("https://%s.%s.svc.cluster.local", svcName, namespace)
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

// WithCleanEventTypeOld is a SubscriptionOpt that initializes subscription with a not clean event type from v1alpha1
func WithCleanEventTypeOld() SubscriptionOpt {
	return WithSourceAndType(EventSourceClean, OrderCreatedEventType)
}

// WithNotCleanEventSourceAndType is a SubscriptionOpt that initializes subscription with a not clean event source and type
func WithNotCleanEventSourceAndType() SubscriptionOpt {
	return WithSourceAndType(EventSourceUnclean, OrderCreatedUncleanEvent)
}

// WithTypeMatchingStandard is a SubscriptionOpt that initializes the subscription with type matching to standard
func WithTypeMatchingStandard() SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		subscription.Spec.TypeMatching = eventingv1alpha2.STANDARD
	}
}

// WithTypeMatchingExact is a SubscriptionOpt that initializes the subscription with type matching to exact
func WithTypeMatchingExact() SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		subscription.Spec.TypeMatching = eventingv1alpha2.EXACT
	}
}

// WithStatusMaxInFlight is a SubscriptionOpt that sets the status with the maxInFlightMessages value
func WithStatusMaxInFlight(maxInFlight int) SubscriptionOpt {
	return func(subscription *eventingv1alpha2.Subscription) {
		subscription.Status.Backend.MaxInFlightMessages = maxInFlight // TODO: change with the new version
	}
}
