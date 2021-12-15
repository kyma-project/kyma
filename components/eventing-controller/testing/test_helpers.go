package testing

import (
	"fmt"
	"net/http"

	k8sresource "k8s.io/apimachinery/pkg/api/resource"

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
	OrderUpdatedV1Event                      = "order.updated.v1"
	OrderCreatedEventType                    = EventTypePrefix + "." + ApplicationName + "." + OrderCreatedV1Event
	OrderUpdatedEventType                    = EventTypePrefix + "." + ApplicationName + "." + OrderUpdatedV1Event
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

	StructuredCloudEventUpdated = `{
           "id":"` + EventID + `",
           "type":"` + OrderUpdatedEventType + `",
           "specversion":"` + EventSpecVersion + `",
           "source":"` + EventSource + `",
           "data":"` + EventData + `"
        }`
)

type APIRuleOption func(rule *apigatewayv1alpha1.APIRule)

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

func WithService(host, svcName string, apiRule *apigatewayv1alpha1.APIRule) {
	port := uint32(443)
	isExternal := true
	apiRule.Spec.Service = &apigatewayv1alpha1.Service{
		Name:       &svcName,
		Port:       &port,
		Host:       &host,
		IsExternal: &isExternal,
	}
}

func WithPath(apiRule *apigatewayv1alpha1.APIRule) {
	handlerOAuth := object.OAuthHandlerName
	handler := apigatewayv1alpha1.Handler{
		Name: handlerOAuth,
	}
	authenticator := &apigatewayv1alpha1.Authenticator{
		Handler: &handler,
	}
	apiRule.Spec.Rules = []apigatewayv1alpha1.Rule{
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

func WithStatusReady(apiRule *apigatewayv1alpha1.APIRule) {
	statusOK := &apigatewayv1alpha1.APIRuleResourceStatus{
		Code:        apigatewayv1alpha1.StatusOK,
		Description: "",
	}

	apiRule.Status = apigatewayv1alpha1.APIRuleStatus{
		APIRuleStatus:        statusOK,
		VirtualServiceStatus: statusOK,
		AccessRuleStatus:     statusOK,
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

type protoOpt func(proto *eventingv1alpha1.ProtocolSettings)

func NewProtocolSettings(opts ...protoOpt) *eventingv1alpha1.ProtocolSettings {
	protoSettings := &eventingv1alpha1.ProtocolSettings{}
	for _, o := range opts {
		o(protoSettings)
	}
	return protoSettings
}

func WithBinaryContentMode(protoSettings *eventingv1alpha1.ProtocolSettings) {
	protoSettings.ContentMode = func() *string {
		contentMode := eventingv1alpha1.ProtocolSettingsContentModeBinary
		return &contentMode
	}()
}

func WithExemptHandshake(protoSettings *eventingv1alpha1.ProtocolSettings) {
	protoSettings.ExemptHandshake = func() *bool {
		exemptHandshake := true
		return &exemptHandshake
	}()
}

func WithAtLeastOnceQOS(protoSettings *eventingv1alpha1.ProtocolSettings) {
	protoSettings.Qos = func() *string {
		qos := "AT-LEAST_ONCE"
		return &qos
	}()
}

func WithDefaultWebhookAuth(protoSettings *eventingv1alpha1.ProtocolSettings) {
	protoSettings.WebhookAuth = &eventingv1alpha1.WebhookAuth{
		Type:         "oauth2",
		GrantType:    "client_credentials",
		ClientID:     "xxx",
		ClientSecret: "xxx",
		TokenURL:     "https://oauth2.xxx.com/oauth2/token",
		Scope:        []string{"guid-identifier"},
	}
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

func qos(qos string) *string {
	q := qos
	return &q
}

func WithFakeSubscriptionStatus(s *eventingv1alpha1.Subscription) {
	s.Status.Conditions = []eventingv1alpha1.Condition{
		{
			Type:    "foo",
			Status:  "foo",
			Reason:  "foo-reason",
			Message: "foo-message",
		},
	}
}

func WithWebhookAuthForBEB(s *eventingv1alpha1.Subscription) {
	s.Spec.Protocol = "BEB"
	s.Spec.ProtocolSettings = &eventingv1alpha1.ProtocolSettings{
		ContentMode: func() *string {
			contentMode := eventingv1alpha1.ProtocolSettingsContentModeBinary
			return &contentMode
		}(),
		ExemptHandshake: exemptHandshake(true),
		Qos:             qos("AT_LEAST_ONCE"),
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

func WithWebhookForNats(s *eventingv1alpha1.Subscription) {
	s.Spec.Protocol = "NATS"
	s.Spec.ProtocolSettings = &eventingv1alpha1.ProtocolSettings{}
}

// WithNotCleanEventTypeFilter initializes subscription filter with a not clean event-type
// A not clean event-type means it contains none-alphanumeric characters
func WithNotCleanEventTypeFilter(s *eventingv1alpha1.Subscription) {
	s.Spec.Filter = &eventingv1alpha1.BEBFilters{
		Filters: []*eventingv1alpha1.BEBFilter{
			{
				EventSource: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "source",
					Value:    EventSource,
				},
				EventType: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "type",
					Value:    OrderCreatedEventTypeNotClean,
				},
			},
		},
	}
}

// WithFilter sets the Subscription filter with the given event source and type.
func WithFilter(eventSource, eventType string) SubscriptionOpt {
	return func(subscription *eventingv1alpha1.Subscription) {
		subscription.Spec.Filter = &eventingv1alpha1.BEBFilters{
			Filters: []*eventingv1alpha1.BEBFilter{
				{
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
				},
			},
		}
	}
}

func WithEventTypeFilter(s *eventingv1alpha1.Subscription) {
	s.Spec.Filter = &eventingv1alpha1.BEBFilters{
		Filters: []*eventingv1alpha1.BEBFilter{
			{
				EventSource: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "source",
					Value:    EventSource,
				},
				EventType: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "type",
					Value:    OrderCreatedEventType,
				},
			},
		},
	}
}

func WithValidSink(svcNs, svcName string, s *eventingv1alpha1.Subscription) {
	s.Spec.Sink = GetValidSink(svcNs, svcName)
}

func GetValidSink(svcNs, svcName string) string {
	return fmt.Sprintf("https://%s.%s.svc.cluster.local", svcName, svcNs)
}

func NewSubscriberSvc(name, ns string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
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

func WithBEBMessagingSecret(name, ns string) *corev1.Secret {
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
			Namespace: ns,
		},
		StringData: map[string]string{
			"messaging": messagingValue,
			"namespace": "test/ns",
		},
	}
}

func WithNamespace(name string) *corev1.Namespace {
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

func WithEventingBackend(name, ns string) *eventingv1alpha1.EventingBackend {
	return &eventingv1alpha1.EventingBackend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec:   eventingv1alpha1.EventingBackendSpec{},
		Status: eventingv1alpha1.EventingBackendStatus{},
	}
}

func WithEventingControllerDeployment() *appsv1.Deployment {
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
func WithEventingControllerPod(backend string) *corev1.Pod {
	labels := map[string]string{
		deployment.AppLabelKey: deployment.PublisherName,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eventing-publisher-proxy-fffff",
			Namespace: deployment.ControllerNamespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  deployment.PublisherName,
					Image: "container-image1",
					Env: []corev1.EnvVar{{
						Name:  "BACKEND",
						Value: backend,
					}},
					Resources: corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]k8sresource.Quantity{
							corev1.ResourceCPU:    k8sresource.MustParse("50m"),
							corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
						},
						Requests: map[corev1.ResourceName]k8sresource.Quantity{
							corev1.ResourceCPU:    k8sresource.MustParse("20m"),
							corev1.ResourceMemory: k8sresource.MustParse("20Mi"),
						},
					},
				},
			},
		},
	}
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
