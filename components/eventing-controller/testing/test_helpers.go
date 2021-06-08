package testing

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"

	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
)

const (
	ApplicationName         = "testapp1023"
	ApplicationNameNotClean = "test-app_1-0+2=3"

	// event properties
	EventSource       = "/default/kyma/id"
	EventTypePrefix   = "sap.kyma"
	EventType         = EventTypePrefix + "." + ApplicationName + ".order.created.v1"
	EventTypeNotClean = EventTypePrefix + "." + ApplicationNameNotClean + ".order.created.v1"
	EventID           = "8945ec08-256b-11eb-9928-acde48001122"
	EventSpecVersion  = "1.0"
	EventData         = "test-data"

	StructuredCloudEvent = `{
           "id":"` + EventID + `",
           "type":"` + EventType + `",
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
	handler := oryv1alpha1.Handler{
		Name: handlerOAuth,
	}
	authenticator := &oryv1alpha1.Authenticator{
		Handler: &handler,
	}
	apiRule.Spec.Rules = []apigatewayv1alpha1.Rule{
		{
			Path: "/path",
			Methods: []string{
				http.MethodPost,
				http.MethodOptions,
			},
			AccessStrategies: []*oryv1alpha1.Authenticator{
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

type subOpt func(subscription *eventingv1alpha1.Subscription)

func NewSubscription(name, namespace string, opts ...subOpt) *eventingv1alpha1.Subscription {
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

func exemptHandshake(val bool) *bool {
	exemptHandshake := val
	return &exemptHandshake
}

func qos(qos string) *string {
	q := qos
	return &q
}

func WithWebhookAuthForBEB(s *eventingv1alpha1.Subscription) {
	s.Spec.Protocol = "BEB"
	s.Spec.ProtocolSettings = &eventingv1alpha1.ProtocolSettings{
		ContentMode: func() *string {
			contentMode := eventingv1alpha1.ProtocolSettingsContentModeBinary
			return &contentMode
		}(),
		ExemptHandshake: exemptHandshake(true),
		Qos:             qos("AT-LEAST_ONCE"),
		WebhookAuth: &eventingv1alpha1.WebhookAuth{
			Type:         "oauth2",
			GrantType:    "client_credentials",
			ClientId:     "xxx",
			ClientSecret: "xxx",
			TokenUrl:     "https://oauth2.xxx.com/oauth2/token",
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
	s.Spec.Filter = &eventingv1alpha1.BebFilters{
		Filters: []*eventingv1alpha1.BebFilter{
			{
				EventSource: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "source",
					Value:    EventSource,
				},
				EventType: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "type",
					Value:    EventTypeNotClean,
				},
			},
		},
	}
}

func WithEmptyEventTypeFilter(s *eventingv1alpha1.Subscription) {
	s.Spec.Filter = &eventingv1alpha1.BebFilters{
		Filters: []*eventingv1alpha1.BebFilter{
			{
				EventSource: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "source",
					Value:    EventSource,
				},
				EventType: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "type",
					Value:    "",
				},
			},
		},
	}
}

func WithEventTypeFilter(s *eventingv1alpha1.Subscription) {
	s.Spec.Filter = &eventingv1alpha1.BebFilters{
		Filters: []*eventingv1alpha1.BebFilter{
			{
				EventSource: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "source",
					Value:    EventSource,
				},
				EventType: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "type",
					Value:    EventType,
				},
			},
		},
	}
}

func WithEmptySourceEventType(s *eventingv1alpha1.Subscription) {
	s.Spec.Filter = &eventingv1alpha1.BebFilters{
		Filters: []*eventingv1alpha1.BebFilter{
			{
				EventSource: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "source",
					Value:    "",
				},
				EventType: &eventingv1alpha1.Filter{
					Type:     "exact",
					Property: "type",
					Value:    EventType,
				},
			},
		},
	}
}

func WithValidSink(svcNs, svcName string, s *eventingv1alpha1.Subscription) {
	s.Spec.Sink = fmt.Sprintf("https://%s.%s.svc.cluster.local", svcName, svcNs)
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
