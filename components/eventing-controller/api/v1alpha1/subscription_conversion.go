package v1alpha1

import (
	"fmt"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"strconv"
)

// ConvertTo converts this Subscription to the Hub version (v2).
func (src *Subscription) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.Subscription)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// TypeMeta
	dst.TypeMeta = src.TypeMeta

	// SPEC fields

	dst.Spec.ID = src.Spec.ID
	dst.Spec.Sink = src.Spec.Sink
	src.setTypeMatching(dst)

	// Types
	src.setSpecTypes(dst)

	// BEB
	src.checkBEBSpecConfig(dst)
	src.checkNATSSpecConfig(dst)

	// STATUS Fields

	// Ready
	dst.Status.Ready = src.Status.Ready

	// Conditions
	for _, condition := range src.Status.Conditions {
		dst.Status.Conditions = append(dst.Status.Conditions, ConditionToAlpha2Version(condition))
	}

	// Types
	for _, cleanEventType := range dst.Spec.Types {
		// todo we need to retrieve the original type
		originalType := cleanEventType
		eventType := v1alpha2.EventType{
			OriginalType: originalType,
			CleanType:    cleanEventType,
		}
		dst.Status.Types = append(dst.Status.Types, eventType)
	}

	// Backend-specific status

	// BEB
	if src.Status.EmsSubscriptionStatus != nil {
		src.setBEBBackendStatus(dst)
	} else { // NATS
		src.setNATSBackendStatus(dst)
	}

	return nil
}

func (src *Subscription) setTypeMatching(dst *v1alpha2.Subscription) {
	dst.Spec.TypeMatching = v1alpha2.STANDARD
}

func (src *Subscription) setSpecTypes(dst *v1alpha2.Subscription) {
	for _, filter := range src.Spec.Filter.Filters {
		if dst.Spec.Source == "" {
			dst.Spec.Source = filter.EventSource.Value
		}
		dst.Spec.Types = append(dst.Spec.Types, filter.EventType.Value)
	}
}

func (src *Subscription) checkBEBSpecConfig(dst *v1alpha2.Subscription) {
	dst.Spec.Protocol = src.Spec.Protocol
	if dst.Spec.ProtocolSettings != nil {
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
	}
}

func (src *Subscription) checkNATSSpecConfig(dst *v1alpha2.Subscription) {
	if src.Spec.Config != nil {
		dst.Spec.Config = map[string]string{
			"maxInFlightMessages": fmt.Sprint(src.Spec.Config.MaxInFlightMessages),
		}
	}
}

func (src *Subscription) setNATSBackendStatus(dst *v1alpha2.Subscription) {

}
func (src *Subscription) setBEBBackendStatus(dst *v1alpha2.Subscription) {
	dst.Status.Backend.Ev2hash = src.Status.Ev2hash
	dst.Status.Backend.Emshash = src.Status.Emshash
	dst.Status.Backend.ExternalSink = src.Status.ExternalSink
	dst.Status.Backend.FailedActivation = src.Status.FailedActivation
	dst.Status.Backend.APIRuleName = src.Status.APIRuleName
	dst.Status.Backend.EmsSubscriptionStatus = &v1alpha2.EmsSubscriptionStatus{
		Status:                   src.Status.EmsSubscriptionStatus.SubscriptionStatus,
		StatusReason:             src.Status.EmsSubscriptionStatus.SubscriptionStatusReason,
		LastSuccessfulDelivery:   src.Status.EmsSubscriptionStatus.LastSuccessfulDelivery,
		LastFailedDelivery:       src.Status.EmsSubscriptionStatus.LastFailedDelivery,
		LastFailedDeliveryReason: src.Status.EmsSubscriptionStatus.LastFailedDeliveryReason,
	}
	dst.Status.Backend.EmsSubscriptionStatus = &v1alpha2.EmsSubscriptionStatus{}

}

func (dst *Subscription) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha2.Subscription)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	if dst.Spec.Filter == nil {
		dst.Spec.Filter = &BEBFilters{
			Filters: []*BEBFilter{},
		}
	}
	for _, eventType := range src.Spec.Types {
		filter := &BEBFilter{
			EventSource: &Filter{
				Type:     "exact",
				Property: "source",
				Value:    "",
			},
			EventType: &Filter{
				Type:     "type",
				Property: "exact",
				Value:    eventType,
			},
		}
		dst.Spec.Filter.Filters = append(dst.Spec.Filter.Filters, filter)
	}
	dst.Spec.Sink = src.Spec.Sink
	dst.Spec.ID = src.Spec.ID
	dst.Spec.Protocol = src.Spec.Protocol
	if src.Spec.Config != nil {
		config := src.Spec.Config
		intVal, err := strconv.Atoi(config["maxInFlightMessages"])
		if err == nil {
			dst.Spec.Config = &SubscriptionConfig{}
			dst.Spec.Config.MaxInFlightMessages = intVal
		}
	}

	var conditions []Condition
	for _, condition := range src.Status.Conditions {
		newCondition := Condition{
			Type:               ConditionType(condition.Type),
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime,
			Reason:             ConditionReason(condition.Reason),
			Message:            condition.Message,
		}
		conditions = append(conditions, newCondition)
	}
	dst.Status.Conditions = conditions
	dst.Status.Ready = src.Status.Ready
	// todo add logic to retrieve the cleanEventTypes
	dst.Status.CleanEventTypes = src.Spec.Types
	dst.Status.Ev2hash = src.Status.Backend.Ev2hash
	dst.Status.Emshash = src.Status.Backend.Emshash
	dst.Status.ExternalSink = src.Status.Backend.ExternalSink
	dst.Status.FailedActivation = src.Status.Backend.FailedActivation
	dst.Status.APIRuleName = src.Status.Backend.APIRuleName

	return nil
}
