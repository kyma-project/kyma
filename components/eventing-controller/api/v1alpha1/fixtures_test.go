package v1alpha1_test

import (
	"fmt"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
)

const (
	eventSource                   = "source"
	orderCreatedEventType         = "prefix." + "noapp." + "order.created.v1"
	orderUpdatedEventType         = "prefix." + "app." + "order.updated.v1"
	orderDeletedEventType         = "prefix." + "noapp." + "order.deleted.v1"
	orderDeletedEventTypeNonClean = "prefix." + "noapp." + "order.deleted_&.v1"
	orderProcessedEventType       = "prefix." + "noapp." + "order.processed.v1"
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

func newDefaultSubscription(opts ...eventingtesting.SubscriptionV1alpha1Opt) *v1alpha1.Subscription {
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

	// remove nats specific field in eventmesh case
	if newSub.Status.EmsSubscriptionStatus != nil {
		newSub.Spec.Config = nil
		newSub.Status.Config = nil
	}

	return newSub
}

// extend the v1 Subscription helpers with Status fields

func v1WithWebhookAuthForBEB() eventingtesting.SubscriptionV1alpha1Opt {
	return func(s *v1alpha1.Subscription) {
		s.Spec.Protocol = "BEB"
		s.Spec.ProtocolSettings = &v1alpha1.ProtocolSettings{
			ContentMode: func() *string {
				contentMode := v1alpha1.ProtocolSettingsContentModeBinary
				return &contentMode
			}(),
			Qos: func() *string {
				qos := "AT_LEAST_ONCE"
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

func v1WithBEBStatusFields() eventingtesting.SubscriptionV1alpha1Opt {
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

func newV2DefaultSubscription(opts ...eventingtesting.SubscriptionOpt) *v1alpha2.Subscription {
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

// extend the v2 Subscription helpers with Status fields

func v2WithBEBStatusFields() eventingtesting.SubscriptionOpt {
	return func(s *v1alpha2.Subscription) {
		s.Status.Backend.Ev2hash = 123
		s.Status.Backend.ExternalSink = "testlink.com"
		s.Status.Backend.FailedActivation = "123156464672"
		s.Status.Backend.APIRuleName = "APIRule"
		s.Status.Backend.EventMeshSubscriptionStatus = &v1alpha2.EventMeshSubscriptionStatus{
			Status:                   "not active",
			StatusReason:             "reason",
			LastSuccessfulDelivery:   "",
			LastFailedDelivery:       "1345613234",
			LastFailedDeliveryReason: "failed",
		}
	}
}

func v2WithStatusTypes(statusTypes []v1alpha2.EventType) eventingtesting.SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		if statusTypes == nil {
			sub.Status.InitializeEventTypes()
			return
		}
		sub.Status.Types = statusTypes
	}
}

func v2WithStatusJetStreamTypes(types []v1alpha2.JetStreamTypes) eventingtesting.SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		sub.Status.Backend.Types = types
	}
}
