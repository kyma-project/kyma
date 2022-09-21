package v1alpha1

import (
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EventSource                   = "source"
	OrderCreatedEventType         = "prefix." + "noapp." + "order.created.v1"
	OrderUpdatedEventType         = "prefix." + "app." + "order.updated.v1"
	OrderDeletedEventType         = "prefix." + "noapp." + "order.deleted.v1"
	OrderDeletedEventTypeNonClean = "prefix." + "noapp." + "order.deleted_&.v1"
)

// +kubebuilder:object:generate=false
type SubscriptionOpt func(subscription *Subscription)

func NewDefaultSubscription(opts ...SubscriptionOpt) *Subscription {
	var defaultConditions []Condition
	for _, condition := range v1alpha2.DefaultConditions {
		defaultConditions = append(defaultConditions, ConditionV2ToV1(condition))
	}
	newSub := &Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      v1alpha2.DefaultName,
			Namespace: v1alpha2.DefaultNamespace,
		},
		Spec: SubscriptionSpec{
			Sink:   v1alpha2.DefaultSink,
			ID:     v1alpha2.DefaultID,
			Config: &SubscriptionConfig{MaxInFlightMessages: v1alpha2.DefaultMaxInFlight},
		},
		Status: SubscriptionStatus{
			Conditions: defaultConditions,
			Ready:      v1alpha2.DefaultStatusReady,
			Config:     &SubscriptionConfig{MaxInFlightMessages: v1alpha2.DefaultMaxInFlight},
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

func exemptHandshake(val bool) *bool {
	exemptHandshake := val
	return &exemptHandshake
}

func WithStatus(status bool) SubscriptionOpt {
	return func(sub *Subscription) {
		sub.Status.Ready = status
	}
}

func WithStatusCleanEventTypes(cleanEventTypes []string) SubscriptionOpt {
	return func(sub *Subscription) {
		if cleanEventTypes == nil {
			sub.Status.InitializeCleanEventTypes()
		} else {
			sub.Status.CleanEventTypes = cleanEventTypes
		}
	}
}

func WithWebhookAuthForBEB() SubscriptionOpt {
	return func(s *Subscription) {
		s.Spec.Protocol = "BEB"
		s.Spec.ProtocolSettings = &ProtocolSettings{
			ContentMode: func() *string {
				contentMode := ProtocolSettingsContentModeBinary
				return &contentMode
			}(),
			Qos: func() *string {
				qos := "true"
				return &qos
			}(),
			ExemptHandshake: exemptHandshake(true),
			WebhookAuth: &WebhookAuth{
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

func WithProtocolBEB() SubscriptionOpt {
	return func(s *Subscription) {
		s.Spec.Protocol = "BEB"
	}
}

func WithBEBStatusFields() SubscriptionOpt {
	return func(s *Subscription) {
		s.Status.Ev2hash = 123
		s.Status.ExternalSink = "testlink.com"
		s.Status.FailedActivation = "123156464672"
		s.Status.APIRuleName = "APIRule"
		s.Status.EmsSubscriptionStatus = &EmsSubscriptionStatus{
			SubscriptionStatus:       "not active",
			SubscriptionStatusReason: "reason",
			LastSuccessfulDelivery:   "",
			LastFailedDelivery:       "1345613234",
			LastFailedDeliveryReason: "failed",
		}
	}
}

// WithWebhookForNATS is a SubscriptionOpt for creating a Subscription with a webhook set to the NATS protocol.
func WithWebhookForNATS() SubscriptionOpt {
	return func(s *Subscription) {
		s.Spec.Protocol = "NATS"
		s.Spec.ProtocolSettings = &ProtocolSettings{}
	}
}

// WithFilter is a SubscriptionOpt for creating a Subscription with a specific event type filter,
// that itself gets created from the passed eventSource and eventType.
func WithFilter(eventSource, eventType string) SubscriptionOpt {
	return func(subscription *Subscription) { AddFilter(eventSource, eventType, subscription) }
}

// AddFilter creates a new Filter from eventSource and eventType and adds it to the subscription.
func AddFilter(eventSource, eventType string, subscription *Subscription) {
	if subscription.Spec.Filter == nil {
		subscription.Spec.Filter = &BEBFilters{
			Filters: []*BEBFilter{},
		}
	}

	filter := &BEBFilter{
		EventSource: &Filter{
			Type:     "exact",
			Property: "source",
			Value:    eventSource,
		},
		EventType: &Filter{
			Type:     "exact",
			Property: "type",
			Value:    eventType,
		},
	}

	subscription.Spec.Filter.Filters = append(subscription.Spec.Filter.Filters, filter)
}

// WithEmptyFilter is a SubscriptionOpt for creating a subscription with an empty event type filter.
// Note that this is different from setting Filter to nil.
func WithEmptyFilter() SubscriptionOpt {
	return func(subscription *Subscription) {
		subscription.Spec.Filter = &BEBFilters{
			Filters: []*BEBFilter{},
		}
		subscription.Status.InitializeCleanEventTypes()
	}
}
