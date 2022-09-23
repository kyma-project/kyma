package v1alpha2

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ProtocolSettingsContentModeBinary = "BINARY"
	DefaultName                       = "test"
	DefaultNamespace                  = "test-namespace"
	DefaultSink                       = "https://svc2.test.local"
	DefaultID                         = "id"
	DefaultMaxInFlight                = 10
	DefaultStatusReady                = true
	DefaultConditions                 = []Condition{
		{
			Type:   ConditionSubscriptionActive,
			Status: "true",
		},
		{
			Type:   ConditionSubscribed,
			Status: "false",
		}}
)

// +kubebuilder:object:generate=false
type SubscriptionOpt func(subscription *Subscription)

func NewDefaultSubscription(opts ...SubscriptionOpt) *Subscription {
	newSub := &Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultName,
			Namespace: DefaultNamespace,
		},
		Spec: SubscriptionSpec{
			TypeMatching: TypeMatchingExact,
			Sink:         DefaultSink,
			ID:           DefaultID,
			Config: map[string]string{
				MaxInFlightMessages: fmt.Sprint(DefaultMaxInFlight),
			},
		},
		Status: SubscriptionStatus{
			Ready:      DefaultStatusReady,
			Conditions: DefaultConditions,
		},
	}
	for _, o := range opts {
		o(newSub)
	}

	return newSub
}

func WithMaxInFlight(maxInFlight string) SubscriptionOpt {
	return func(sub *Subscription) {
		sub.Spec.Config = map[string]string{
			MaxInFlightMessages: fmt.Sprint(maxInFlight),
		}
	}
}

func WithSource(source string) SubscriptionOpt {
	return func(sub *Subscription) {
		sub.Spec.Source = source
	}
}

func WithTypes(types []string) SubscriptionOpt {
	return func(sub *Subscription) {
		sub.Spec.Types = types
	}
}
func WithStatusJetStreamTypes(types []JetStreamTypes) SubscriptionOpt {
	return func(sub *Subscription) {
		sub.Status.Backend.Types = types
	}
}

func WithStatusTypes(statusTypes []EventType) SubscriptionOpt {
	return func(sub *Subscription) {
		if statusTypes == nil {
			sub.Status.InitializeEventTypes()
			return
		}
		sub.Status.Types = statusTypes
	}
}

func WithWebhookAuthForBEB() SubscriptionOpt {
	return func(s *Subscription) {
		s.Spec.Config = map[string]string{
			Protocol:                        "BEB",
			ProtocolSettingsContentMode:     ProtocolSettingsContentModeBinary,
			ProtocolSettingsExemptHandshake: "true",
			ProtocolSettingsQos:             "true",
			WebhookAuthType:                 "oauth2",
			WebhookAuthGrantType:            "client_credentials",
			WebhookAuthClientID:             "xxx",
			WebhookAuthClientSecret:         "xxx",
			WebhookAuthTokenURL:             "https://oauth2.xxx.com/oauth2/token",
			WebhookAuthScope:                "guid-identifier,root",
		}
	}
}

func WithProtocolBEB() SubscriptionOpt {
	return func(s *Subscription) {
		if s.Spec.Config == nil {
			s.Spec.Config = map[string]string{}
		}
		s.Spec.Config[Protocol] = "BEB"
	}
}

func WithBEBStatusFields() SubscriptionOpt {
	return func(s *Subscription) {
		s.Status.Backend.Ev2hash = 123
		s.Status.Backend.ExternalSink = "testlink.com"
		s.Status.Backend.FailedActivation = "123156464672"
		s.Status.Backend.APIRuleName = "APIRule"
		s.Status.Backend.EmsSubscriptionStatus = &EmsSubscriptionStatus{
			Status:                   "not active",
			StatusReason:             "reason",
			LastSuccessfulDelivery:   "",
			LastFailedDelivery:       "1345613234",
			LastFailedDeliveryReason: "failed",
		}
	}
}
