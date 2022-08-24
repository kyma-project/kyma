package v1alpha1

import (
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this Subscription to the Hub version (v2).
func (src *Subscription) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.Subscription)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.ID = src.Spec.ID
	dst.Spec.Protocol = src.Spec.Protocol
	dst.Spec.ProtocolSettings = &v1alpha2.ProtocolSettings{
		ContentMode:     src.Spec.ProtocolSettings.ContentMode,
		ExemptHandshake: src.Spec.ProtocolSettings.ExemptHandshake,
		Qos:             src.Spec.ProtocolSettings.Qos,
		WebhookAuth: &v1alpha2.WebhookAuth{
			Type:         src.Spec.ProtocolSettings.WebhookAuth.Type,
			GrantType:    src.Spec.ProtocolSettings.WebhookAuth.GrantType,
			ClientID:     src.Spec.ProtocolSettings.WebhookAuth.ClientID,
			ClientSecret: src.Spec.ProtocolSettings.WebhookAuth.ClientSecret,
			TokenURL:     src.Spec.ProtocolSettings.WebhookAuth.TokenURL,
			Scope:        src.Spec.ProtocolSettings.WebhookAuth.Scope,
		},
	}
	dst.Spec.Sink = src.Spec.Sink
	dst.Spec.TypeMatching = v1alpha2.EXACT

	for _, filter := range src.Spec.Filter.Filters {
		if dst.Spec.Source == "" {
			dst.Spec.Source = filter.EventSource.Value
		}
		dst.Spec.Types = append(dst.Spec.Types, filter.EventType.Value)
	}

	if src.Spec.Config != nil {
		dst.Spec.Config = map[string]string{
			"maxInFlightMessages": string(src.Spec.Config.MaxInFlightMessages),
		}
	}

	// Status
	var conditions []v1alpha2.Condition
	for _, condition := range src.Status.Conditions {
		newCondition := v1alpha2.Condition{
			Type:               v1alpha2.ConditionType(condition.Type),
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime,
			Reason:             v1alpha2.ConditionReason(condition.Reason),
			Message:            condition.Message,
		}
		conditions = append(conditions, newCondition)
	}
	dst.Status.Conditions = conditions
	dst.Status.Ready = src.Status.Ready
	dst.Status.Types = src.Status.CleanEventTypes
	dst.Status.Ev2hash = src.Status.Ev2hash
	dst.Status.Emshash = src.Status.Emshash
	dst.Status.ExternalSink = src.Status.ExternalSink
	dst.Status.FailedActivation = src.Status.FailedActivation
	dst.Status.APIRuleName = src.Status.APIRuleName
	dst.Status.EmsSubscriptionStatus = &v1alpha2.EmsSubscriptionStatus{
		SubscriptionStatus:       src.Status.EmsSubscriptionStatus.SubscriptionStatus,
		SubscriptionStatusReason: src.Status.EmsSubscriptionStatus.SubscriptionStatusReason,
		LastSuccessfulDelivery:   src.Status.EmsSubscriptionStatus.LastSuccessfulDelivery,
		LastFailedDelivery:       src.Status.EmsSubscriptionStatus.LastFailedDelivery,
		LastFailedDeliveryReason: src.Status.EmsSubscriptionStatus.LastFailedDeliveryReason,
	}

	return nil
}
