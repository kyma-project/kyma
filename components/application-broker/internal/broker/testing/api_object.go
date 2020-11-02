package testing

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"

	securityv1beta1apis "istio.io/api/security/v1beta1"
	istiov1beta1apis "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	FakeChannelName      = "fake-chan"
	FakeSubscriptionName = "fake-sub"
)

// redefine here to avoid cyclic dependency
const (
	integrationNamespace                      = "kyma-integration"
	applicationNameLabelKey                   = "application-name"
	applicationServiceIDLabelKey              = "application-service-id"
	brokerNamespaceLabelKey                   = "broker-namespace"
	knativeEventingInjectionLabelKey          = "knative-eventing-injection"
	knativeEventingInjectionLabelValueEnabled = "enabled"
	knSubscriptionNamePrefix                  = "brokersub"
)

func NewAppSubscription(appNs, appName, appSvcID string, opts ...SubscriptionOption) *messagingv1alpha1.Subscription {
	sub := &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", knSubscriptionNamePrefix),
			Namespace:    integrationNamespace,
			Labels: map[string]string{
				brokerNamespaceLabelKey:      appNs,
				applicationNameLabelKey:      appName,
				applicationServiceIDLabelKey: appSvcID,
			},
		},
	}

	for _, opt := range opts {
		opt(sub)
	}

	return sub
}

// SubscriptionOption is a functional option for Subscription objects.
type SubscriptionOption func(*messagingv1alpha1.Subscription)

// WithSpec sets the spec of a Subscription.
func WithSpec(t *testing.T, subscriberURI string) SubscriptionOption {
	url, err := apis.ParseURL(subscriberURI)
	if err != nil {
		t.Fatalf("error while parsing url: %v, error: %v", subscriberURI, err)
	}
	return func(s *messagingv1alpha1.Subscription) {
		s.Spec = messagingv1alpha1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				Name: FakeChannelName,
			},
			Subscriber: &duckv1.Destination{
				URI: url,
			},
		}
	}
}

// WithNameSuffix generates the name of a Subscription using its GenerateName prefix.
func WithNameSuffix(nameSuffix string) SubscriptionOption {
	return func(s *messagingv1alpha1.Subscription) {
		if s.GenerateName != "" && s.Name == "" {
			s.Name = s.GenerateName + nameSuffix
		}
	}
}

func NewAppNamespace(name string, brokerInjection bool) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if brokerInjection {
		ns.Labels = map[string]string{
			knativeEventingInjectionLabelKey: knativeEventingInjectionLabelValueEnabled,
		}
	}
	return ns
}

func NewDefaultBroker(ns string) *eventingv1alpha1.Broker {
	return &eventingv1alpha1.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: ns,
		},
	}
}

func NewIstioPeerAuthentication(ns, peerAuthName string) *securityv1beta1.PeerAuthentication {
	labels := make(map[string]string)
	labels["eventing.knative.dev/broker"] = "default"
	port := uint32(9090)
	portLevelMtls := map[uint32]*securityv1beta1apis.PeerAuthentication_MutualTLS{
		port: {
			Mode: securityv1beta1apis.PeerAuthentication_MutualTLS_PERMISSIVE,
		},
	}
	return &securityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      peerAuthName,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: securityv1beta1apis.PeerAuthentication{
			Selector: &istiov1beta1apis.WorkloadSelector{
				MatchLabels: labels,
			},
			PortLevelMtls: portLevelMtls,
		},
	}
}

func NewAppChannel(appName string) *messagingv1alpha1.Channel {
	return &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FakeChannelName,
			Namespace: integrationNamespace,
			Labels: map[string]string{
				applicationNameLabelKey: appName,
			},
		},
	}
}

func NewEventActivation(ns, name string) *v1alpha1.EventActivation {
	return &v1alpha1.EventActivation{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1alpha1.EventActivationSpec{
			DisplayName: "DisplayName",
			SourceID:    "source-id",
		},
	}
}
