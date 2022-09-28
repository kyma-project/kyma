package v1alpha1_test

import (
	"fmt"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	eventSource                   = "source"
	orderCreatedEventType         = "prefix." + "noapp." + "order.created.v1"
	orderUpdatedEventType         = "prefix." + "app." + "order.updated.v1"
	orderDeletedEventType         = "prefix." + "noapp." + "order.deleted.v1"
	orderDeletedEventTypeNonClean = "prefix." + "noapp." + "order.deleted_&.v1"
)

const (
	defaultName        = "test"
	defaultNamespace   = "test-namespace"
	defaultSink        = "https://svc2.test.local"
	defaultID          = "id"
	defaultMaxInFlight = 10
	defaultStatusReady = true
)

var (
	v2DefaultConditions = []v1alpha2.Condition{
		{
			Type:   v1alpha2.ConditionSubscriptionActive,
			Status: "true",
		},
		{
			Type:   v1alpha2.ConditionSubscribed,
			Status: "false",
		}}
)

type subscriptionOpt = eventingtesting.SubscriptionOpt

func newDefaultSubscription(opts ...subscriptionOpt) *v1alpha1.Subscription {
	var defaultConditions []v1alpha1.Condition
	for _, condition := range v2DefaultConditions {
		defaultConditions = append(defaultConditions, v1alpha1.ConditionV2ToV1(condition))
	}
	newSub := &v1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultName,
			Namespace: defaultNamespace,
		},
		Spec: v1alpha1.SubscriptionSpec{
			Sink:   defaultSink,
			ID:     defaultID,
			Config: &v1alpha1.SubscriptionConfig{MaxInFlightMessages: defaultMaxInFlight},
		},
		Status: v1alpha1.SubscriptionStatus{
			Conditions: defaultConditions,
			Ready:      defaultStatusReady,
			Config:     &v1alpha1.SubscriptionConfig{MaxInFlightMessages: defaultMaxInFlight},
		},
	}
	for _, o := range opts {
		o(newSub)
	}

	// remove nats specific field in beb case
	if newSub.Status.EmsSubscriptionStatus != nil {
		newSub.Spec.Config = nil
		newSub.Status.Config = nil
	}

	return newSub
}

var withStatusCleanEventTypes = eventingtesting.WithStatusCleanEventTypes

// TODO:
// var withWebhookAuthForBEB = eventingtesting.WithWebhookAuthForBEB
func withWebhookAuthForBEB() subscriptionOpt {
	return func(s *v1alpha1.Subscription) {
		s.Spec.Protocol = "BEB"
		s.Spec.ProtocolSettings = &v1alpha1.ProtocolSettings{
			ContentMode: func() *string {
				contentMode := v1alpha1.ProtocolSettingsContentModeBinary
				return &contentMode
			}(),
			Qos: func() *string {
				qos := "true"
				return &qos
			}(),
			ExemptHandshake: utils.BoolPtr(true),
			WebhookAuth: &v1alpha1.WebhookAuth{
				Type:         "oauth2",
				GrantType:    "client_credentials",
				ClientID:     "xxx",
				ClientSecret: "xxx",
				TokenURL:     "https://oauth2.xxx.com/oauth2/token",
				Scope:        []string{"guid-identifier", "root"},
			},
		}
	}
}

var withProtocolBEB = eventingtesting.WithProtocolBEB

func withBEBStatusFields() subscriptionOpt {
	return func(s *v1alpha1.Subscription) {
		s.Status.Ev2hash = 123
		s.Status.ExternalSink = "testlink.com"
		s.Status.FailedActivation = "123156464672"
		s.Status.APIRuleName = "APIRule"
		s.Status.EmsSubscriptionStatus = &v1alpha1.EmsSubscriptionStatus{
			SubscriptionStatus:       "not active",
			SubscriptionStatusReason: "reason",
			LastSuccessfulDelivery:   "",
			LastFailedDelivery:       "1345613234",
			LastFailedDeliveryReason: "failed",
		}
	}
}

var withFilter = eventingtesting.WithFilter

// TODO:
// var withEmptyFilter = eventingtesting.WithEmptyFilter

// withEmptyFilter is a subscriptionOpt for creating a subscription with an empty event type filter.
// Note that this is different from setting Filter to nil.
func withEmptyFilter() subscriptionOpt {
	return func(subscription *v1alpha1.Subscription) {
		subscription.Spec.Filter = &v1alpha1.BEBFilters{
			Filters: []*v1alpha1.BEBFilter{},
		}
		subscription.Status.InitializeCleanEventTypes()
	}
}

type v2SubscriptionOpt func(subscription *v1alpha2.Subscription)

func v2WithMaxInFlight(maxInFlight string) v2SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		sub.Spec.Config = map[string]string{
			v1alpha2.MaxInFlightMessages: fmt.Sprint(maxInFlight),
		}
	}
}

func v2WithSource(source string) v2SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		sub.Spec.Source = source
	}
}

func v2WithTypes(types []string) v2SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		sub.Spec.Types = types
	}
}

func v2WithStatusJetStreamTypes(types []v1alpha2.JetStreamTypes) v2SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		sub.Status.Backend.Types = types
	}
}

func v2WithStatusTypes(statusTypes []v1alpha2.EventType) v2SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		if statusTypes == nil {
			sub.Status.InitializeEventTypes()
			return
		}
		sub.Status.Types = statusTypes
	}
}

func v2WithWebhookAuthForBEB() v2SubscriptionOpt {
	return func(s *v1alpha2.Subscription) {
		s.Spec.Config = map[string]string{
			v1alpha2.Protocol:                        "BEB",
			v1alpha2.ProtocolSettingsContentMode:     v1alpha1.ProtocolSettingsContentModeBinary,
			v1alpha2.ProtocolSettingsExemptHandshake: "true",
			v1alpha2.ProtocolSettingsQos:             "true",
			v1alpha2.WebhookAuthType:                 "oauth2",
			v1alpha2.WebhookAuthGrantType:            "client_credentials",
			v1alpha2.WebhookAuthClientID:             "xxx",
			v1alpha2.WebhookAuthClientSecret:         "xxx",
			v1alpha2.WebhookAuthTokenURL:             "https://oauth2.xxx.com/oauth2/token",
			v1alpha2.WebhookAuthScope:                "guid-identifier,root",
		}
	}
}

func v2WithProtocolBEB() v2SubscriptionOpt {
	return func(s *v1alpha2.Subscription) {
		if s.Spec.Config == nil {
			s.Spec.Config = map[string]string{}
		}
		s.Spec.Config[v1alpha2.Protocol] = "BEB"
	}
}

func v2WithBEBStatusFields() v2SubscriptionOpt {
	return func(s *v1alpha2.Subscription) {
		s.Status.Backend.Ev2hash = 123
		s.Status.Backend.ExternalSink = "testlink.com"
		s.Status.Backend.FailedActivation = "123156464672"
		s.Status.Backend.APIRuleName = "APIRule"
		s.Status.Backend.EmsSubscriptionStatus = &v1alpha2.EmsSubscriptionStatus{
			Status:                   "not active",
			StatusReason:             "reason",
			LastSuccessfulDelivery:   "",
			LastFailedDelivery:       "1345613234",
			LastFailedDeliveryReason: "failed",
		}
	}
}
func newV2DefaultSubscription(opts ...v2SubscriptionOpt) *v1alpha2.Subscription {
	newSub := &v1alpha2.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultName,
			Namespace: defaultNamespace,
		},
		Spec: v1alpha2.SubscriptionSpec{
			TypeMatching: v1alpha2.TypeMatchingExact,
			Sink:         defaultSink,
			ID:           defaultID,
			Config: map[string]string{
				v1alpha2.MaxInFlightMessages: fmt.Sprint(defaultMaxInFlight),
			},
		},
		Status: v1alpha2.SubscriptionStatus{
			Ready:      defaultStatusReady,
			Conditions: v2DefaultConditions,
		},
	}
	for _, o := range opts {
		o(newSub)
	}

	return newSub
}
