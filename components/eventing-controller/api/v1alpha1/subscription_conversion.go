package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

const (
	ErrorHubVersionMsg     = "hub version is not the expected v1alpha2 version"
	ErrorMultipleSourceMsg = "subscription contains more than 1 eventSource"
)

// ConvertTo converts this Subscription in version v1 to the Hub version v2.
func (src *Subscription) ConvertTo(dstRaw conversion.Hub) error { //nolint:revive
	dst, ok := dstRaw.(*v1alpha2.Subscription)
	if !ok {
		return errors.Errorf(ErrorHubVersionMsg)
	}
	return V1ToV2(src, dst)
}

// V1ToV2 copies the v1alpha1-type field values into v1alpha2-type field values.
func V1ToV2(src *Subscription, dst *v1alpha2.Subscription) error {

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// SPEC fields

	dst.Spec.ID = src.Spec.ID
	dst.Spec.Sink = src.Spec.Sink
	dst.Spec.Source = ""

	src.setV2TypeMatching(dst)

	// protocol fields
	src.setV2ProtocolFields(dst)

	// Types
	if err := src.setV2SpecTypes(dst); err != nil {
		return err
	}

	// Config
	src.natsSpecConfigToV2(dst)

	// STATUS Fields

	// Ready
	dst.Status.Ready = src.Status.Ready

	// Conditions
	for _, condition := range src.Status.Conditions {
		dst.Status.Conditions = append(dst.Status.Conditions, ConditionV1ToV2(condition))
	}

	// event types
	src.setV2StatusTypes(dst)

	// Backend-specific status

	// BEB
	src.bebBackendStatusToV2(dst)

	// NATS
	src.natsBackendStatusToV2(dst)

	return nil
}

// ConvertFrom converts this Subscription from the Hub version (v2) to v1.
func (dst *Subscription) ConvertFrom(srcRaw conversion.Hub) error { //nolint:revive
	src, ok := srcRaw.(*v1alpha2.Subscription)
	if !ok {
		return errors.Errorf(ErrorHubVersionMsg)
	}
	return V2ToV1(dst, src)
}

// V2ToV1 copies the v1alpha2-type field values into v1alpha1-type field values.
func V2ToV1(dst *Subscription, src *v1alpha2.Subscription) error {

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.ID = src.Spec.ID
	dst.Spec.Sink = src.Spec.Sink

	dst.setV1ProtocolFields(src)

	dst.Spec.Filter = &BEBFilters{
		Filters: []*BEBFilter{},
	}

	for _, eventType := range src.Spec.Types {
		filter := &BEBFilter{
			EventSource: &Filter{
				Property: "source",
				Type:     fmt.Sprint(v1alpha2.TypeMatchingExact),
				Value:    src.Spec.Source,
			},
			EventType: &Filter{
				Type:     fmt.Sprint(v1alpha2.TypeMatchingExact),
				Property: "type",
				Value:    eventType,
			},
		}
		dst.Spec.Filter.Filters = append(dst.Spec.Filter.Filters, filter)
	}

	if src.Spec.Config != nil {
		if err := dst.natsSpecConfigToV1(src); err != nil {
			return err
		}
	}

	// Conditions
	for _, condition := range src.Status.Conditions {
		dst.Status.Conditions = append(dst.Status.Conditions, ConditionV2ToV1(condition))
	}

	dst.Status.Ready = src.Status.Ready

	dst.setV1CleanEvenTypes(src)
	dst.bebBackendStatusToV1(src)
	dst.natsBackendStatusToV1(src)

	return nil
}

// setV2TypeMatching sets the default typeMatching on the v1alpha2 Subscription version.
func (src *Subscription) setV2TypeMatching(dst *v1alpha2.Subscription) {
	dst.Spec.TypeMatching = v1alpha2.TypeMatchingExact
}

// setV2ProtocolFields converts the protocol-related fields from v1alpha1 to v1alpha2 Subscription version.
func (src *Subscription) setV2ProtocolFields(dst *v1alpha2.Subscription) {
	dst.Spec.Config = map[string]string{}
	if src.Spec.Protocol != "" {
		dst.Spec.Config[v1alpha2.Protocol] = src.Spec.Protocol
	}
	// protocol settings
	if src.Spec.ProtocolSettings != nil {
		if src.Spec.ProtocolSettings.ContentMode != nil {
			dst.Spec.Config[v1alpha2.ProtocolSettingsContentMode] = *src.Spec.ProtocolSettings.ContentMode
		}
		dst.Spec.Config[v1alpha2.ProtocolSettingsExemptHandshake] = fmt.Sprint(*src.Spec.ProtocolSettings.ExemptHandshake)
		if src.Spec.ProtocolSettings.Qos != nil {
			dst.Spec.Config[v1alpha2.ProtocolSettingsQos] = *src.Spec.ProtocolSettings.Qos
		}
		// webhookAuth fields
		if src.Spec.ProtocolSettings.WebhookAuth != nil {
			dst.Spec.Config[v1alpha2.WebhookAuthType] = src.Spec.ProtocolSettings.WebhookAuth.Type
			dst.Spec.Config[v1alpha2.WebhookAuthGrantType] = src.Spec.ProtocolSettings.WebhookAuth.GrantType
			dst.Spec.Config[v1alpha2.WebhookAuthClientID] = src.Spec.ProtocolSettings.WebhookAuth.ClientID
			dst.Spec.Config[v1alpha2.WebhookAuthClientSecret] = src.Spec.ProtocolSettings.WebhookAuth.ClientSecret
			dst.Spec.Config[v1alpha2.WebhookAuthTokenURL] = src.Spec.ProtocolSettings.WebhookAuth.TokenURL
			dst.Spec.Config[v1alpha2.WebhookAuthScope] = strings.Join(src.Spec.ProtocolSettings.WebhookAuth.Scope, ",")
		}
	}
}

// setV1ProtocolFields converts the protocol-related fields from v1alpha1 to v1alpha2 Subscription version.
func (src *Subscription) setV1ProtocolFields(dst *v1alpha2.Subscription) {
	if protocol, ok := dst.Spec.Config[v1alpha2.Protocol]; ok {
		src.Spec.Protocol = protocol
	}

	if currentMode, ok := dst.Spec.Config[v1alpha2.ProtocolSettingsContentMode]; ok {
		src.Spec.ProtocolSettings = &ProtocolSettings{}
		src.Spec.ProtocolSettings.ContentMode = &currentMode
	}
	if qos, ok := dst.Spec.Config[v1alpha2.ProtocolSettingsQos]; ok {
		src.Spec.ProtocolSettings.Qos = &qos
	}
	if exemptHandshake, ok := dst.Spec.Config[v1alpha2.ProtocolSettingsExemptHandshake]; ok {
		handshake, err := strconv.ParseBool(exemptHandshake)
		if err != nil {
			handshake = true
		}
		src.Spec.ProtocolSettings.ExemptHandshake = &handshake
	}
	if src.Spec.ProtocolSettings != nil {
		if authType, ok := dst.Spec.Config[v1alpha2.WebhookAuthType]; ok {
			src.Spec.ProtocolSettings.WebhookAuth = &WebhookAuth{}
			src.Spec.ProtocolSettings.WebhookAuth.Type = authType
		}
		if grantType, ok := dst.Spec.Config[v1alpha2.WebhookAuthGrantType]; ok {
			src.Spec.ProtocolSettings.WebhookAuth.GrantType = grantType
		}
		if clientID, ok := dst.Spec.Config[v1alpha2.WebhookAuthClientID]; ok {
			src.Spec.ProtocolSettings.WebhookAuth.ClientID = clientID
		}
		if secret, ok := dst.Spec.Config[v1alpha2.WebhookAuthClientSecret]; ok {
			src.Spec.ProtocolSettings.WebhookAuth.ClientSecret = secret
		}
		if token, ok := dst.Spec.Config[v1alpha2.WebhookAuthTokenURL]; ok {
			src.Spec.ProtocolSettings.WebhookAuth.TokenURL = token
		}
		if scope, ok := dst.Spec.Config[v1alpha2.WebhookAuthScope]; ok {
			src.Spec.ProtocolSettings.WebhookAuth.Scope = strings.Split(scope, ",")
		}
	}
}

// setV2SpecTypes sets event types in the Subscription Spec in the v1alpha2 way.
func (src *Subscription) setV2SpecTypes(dst *v1alpha2.Subscription) error {
	for _, filter := range src.Spec.Filter.Filters {
		if dst.Spec.Source == "" {
			dst.Spec.Source = filter.EventSource.Value
		}
		if dst.Spec.Source != "" && filter.EventSource.Value != dst.Spec.Source {
			return errors.New(ErrorMultipleSourceMsg)
		}
		dst.Spec.Types = append(dst.Spec.Types, filter.EventType.Value)
	}
	return nil
}

// natsSpecConfigToV2 converts the v1alpha2 Spec config to v1alpha1
func (src *Subscription) natsSpecConfigToV1(dst *v1alpha2.Subscription) error {
	if maxInFlightMessages, ok := dst.Spec.Config[v1alpha2.MaxInFlightMessages]; ok {
		intVal, err := strconv.Atoi(maxInFlightMessages)
		if err != nil {
			return err
		}
		src.Spec.Config = &SubscriptionConfig{
			MaxInFlightMessages: intVal,
		}
	}
	return nil
}

// natsSpecConfigToV2 converts the hardcoded v1alpha1 Spec config to v1alpha2 generic config version.
func (src *Subscription) natsSpecConfigToV2(dst *v1alpha2.Subscription) {
	if src.Spec.Config != nil {
		dst.Spec.Config = map[string]string{
			v1alpha2.MaxInFlightMessages: fmt.Sprint(src.Spec.Config.MaxInFlightMessages),
		}
	}
}

// setV2StatusTypes sets the original/clean event types mapping to the Subscription's Status v1alpha2 version.
func (src *Subscription) setV2StatusTypes(dst *v1alpha2.Subscription) {
	for i, cleanEventType := range dst.Spec.Types {
		originalType := cleanEventType
		eventType := v1alpha2.EventType{
			OriginalType: originalType,
			CleanType:    src.Status.CleanEventTypes[i],
		}
		dst.Status.Types = append(dst.Status.Types, eventType)
	}
}

// bebBackendStatusToV2 moves the BEB-related to Backend fields of the Status in the v1alpha2.
func (src *Subscription) bebBackendStatusToV2(dst *v1alpha2.Subscription) {
	dst.Status.Backend.Ev2hash = src.Status.Ev2hash
	dst.Status.Backend.Emshash = src.Status.Emshash
	dst.Status.Backend.ExternalSink = src.Status.ExternalSink
	dst.Status.Backend.FailedActivation = src.Status.FailedActivation
	dst.Status.Backend.APIRuleName = src.Status.APIRuleName
	if src.Status.EmsSubscriptionStatus != nil {
		dst.Status.Backend.EmsSubscriptionStatus = &v1alpha2.EmsSubscriptionStatus{
			Status:                   src.Status.EmsSubscriptionStatus.SubscriptionStatus,
			StatusReason:             src.Status.EmsSubscriptionStatus.SubscriptionStatusReason,
			LastSuccessfulDelivery:   src.Status.EmsSubscriptionStatus.LastSuccessfulDelivery,
			LastFailedDelivery:       src.Status.EmsSubscriptionStatus.LastFailedDelivery,
			LastFailedDeliveryReason: src.Status.EmsSubscriptionStatus.LastFailedDeliveryReason,
		}
	}
}

// setBEBBackendStatus moves the BEB-related to Backend fields of the Status in the v1alpha2.
func (src *Subscription) bebBackendStatusToV1(dst *v1alpha2.Subscription) {
	src.Status.Ev2hash = dst.Status.Backend.Ev2hash
	src.Status.Emshash = dst.Status.Backend.Emshash
	src.Status.ExternalSink = dst.Status.Backend.ExternalSink
	src.Status.FailedActivation = dst.Status.Backend.FailedActivation
	src.Status.APIRuleName = dst.Status.Backend.APIRuleName
	if dst.Status.Backend.EmsSubscriptionStatus != nil {
		src.Status.EmsSubscriptionStatus = &EmsSubscriptionStatus{
			SubscriptionStatus:       dst.Status.Backend.EmsSubscriptionStatus.Status,
			SubscriptionStatusReason: dst.Status.Backend.EmsSubscriptionStatus.StatusReason,
			LastSuccessfulDelivery:   dst.Status.Backend.EmsSubscriptionStatus.LastSuccessfulDelivery,
			LastFailedDelivery:       dst.Status.Backend.EmsSubscriptionStatus.LastFailedDelivery,
			LastFailedDeliveryReason: dst.Status.Backend.EmsSubscriptionStatus.LastFailedDeliveryReason,
		}
	}
}

// natsBackendStatusToV1 moves the NATS-related to Backend fields of the Status in the v1alpha2.
func (src *Subscription) natsBackendStatusToV1(dst *v1alpha2.Subscription) {
	if maxInFlightMessages, ok := dst.Spec.Config[v1alpha2.MaxInFlightMessages]; ok {
		intVal, err := strconv.Atoi(maxInFlightMessages)
		if err == nil {
			src.Status.Config = &SubscriptionConfig{}
			src.Status.Config.MaxInFlightMessages = intVal
		}
	}
}

// setBEBBackendStatus moves the NATS-related to Backend fields of the Status in the v1alpha2.
func (src *Subscription) natsBackendStatusToV2(dst *v1alpha2.Subscription) {
	if src.Status.EmsSubscriptionStatus == nil {
		for _, eventType := range dst.Spec.Types {
			dst.Status.Backend.Types = append(dst.Status.Backend.Types, v1alpha2.JetStreamTypes{OriginalType: eventType})
		}
	}
}

// setV1CleanEvenTypes sets the clean event types to v1alpha1 Subscription Status.
func (src *Subscription) setV1CleanEvenTypes(dst *v1alpha2.Subscription) {
	src.Status.InitializeCleanEventTypes()
	for _, eventType := range dst.Status.Types {
		src.Status.CleanEventTypes = append(src.Status.CleanEventTypes, eventType.CleanType)
	}
}

// ConditionV1ToV2 converts the v1alpha1 Condition to v1alpha2 version.
func ConditionV1ToV2(condition Condition) v1alpha2.Condition {
	return v1alpha2.Condition{
		Type:               v1alpha2.ConditionType(condition.Type),
		Status:             condition.Status,
		LastTransitionTime: condition.LastTransitionTime,
		Reason:             v1alpha2.ConditionReason(condition.Reason),
		Message:            condition.Message,
	}
}

// ConditionV2ToV1 converts the v1alpha2 Condition to v1alpha1 version.
func ConditionV2ToV1(condition v1alpha2.Condition) Condition {
	return Condition{
		Type:               ConditionType(condition.Type),
		Status:             condition.Status,
		LastTransitionTime: condition.LastTransitionTime,
		Reason:             ConditionReason(condition.Reason),
		Message:            condition.Message,
	}
}
