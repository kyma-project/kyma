package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
)

const (
	ErrorHubVersionMsg     = "hub version is not the expected v1alpha2 version"
	ErrorMultipleSourceMsg = "subscription contains more than 1 eventSource"
)

var v1alpha1TypeCleaner eventtype.Cleaner //nolint:gochecknoglobals // using global var because there is no runtime
// object to hold this instance.

func InitializeEventTypeCleaner(cleaner eventtype.Cleaner) {
	v1alpha1TypeCleaner = cleaner
}

// ConvertTo converts this Subscription in version v1 to the Hub version v2.
func (src *Subscription) ConvertTo(dstRaw conversion.Hub) error {
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
		Filters: []*EventMeshFilter{},
	}

	for _, eventType := range src.Spec.Types {
		filter := &EventMeshFilter{
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
		if src.Spec.ProtocolSettings.ExemptHandshake != nil {
			dst.Spec.Config[v1alpha2.ProtocolSettingsExemptHandshake] = fmt.Sprint(*src.Spec.ProtocolSettings.ExemptHandshake)
		}
		if src.Spec.ProtocolSettings.Qos != nil {
			dst.Spec.Config[v1alpha2.ProtocolSettingsQos] = *src.Spec.ProtocolSettings.Qos
		}
		// webhookAuth fields
		if src.Spec.ProtocolSettings.WebhookAuth != nil {
			if src.Spec.ProtocolSettings.WebhookAuth.Type != "" {
				dst.Spec.Config[v1alpha2.WebhookAuthType] = src.Spec.ProtocolSettings.WebhookAuth.Type
			}
			dst.Spec.Config[v1alpha2.WebhookAuthGrantType] = src.Spec.ProtocolSettings.WebhookAuth.GrantType
			dst.Spec.Config[v1alpha2.WebhookAuthClientID] = src.Spec.ProtocolSettings.WebhookAuth.ClientID
			dst.Spec.Config[v1alpha2.WebhookAuthClientSecret] = src.Spec.ProtocolSettings.WebhookAuth.ClientSecret
			dst.Spec.Config[v1alpha2.WebhookAuthTokenURL] = src.Spec.ProtocolSettings.WebhookAuth.TokenURL
			if src.Spec.ProtocolSettings.WebhookAuth.Scope != nil {
				dst.Spec.Config[v1alpha2.WebhookAuthScope] = strings.Join(src.Spec.ProtocolSettings.WebhookAuth.Scope, ",")
			}
		}
	}
}

func (src *Subscription) initializeProtocolSettingsIfNil() {
	if src.Spec.ProtocolSettings == nil {
		src.Spec.ProtocolSettings = &ProtocolSettings{}
	}
}

func (src *Subscription) initializeWebhookAuthIfNil() {
	src.initializeProtocolSettingsIfNil()
	if src.Spec.ProtocolSettings.WebhookAuth == nil {
		src.Spec.ProtocolSettings.WebhookAuth = &WebhookAuth{}
	}
}

// setV1ProtocolFields converts the protocol-related fields from v1alpha1 to v1alpha2 Subscription version.
func (src *Subscription) setV1ProtocolFields(dst *v1alpha2.Subscription) {
	if protocol, ok := dst.Spec.Config[v1alpha2.Protocol]; ok {
		src.Spec.Protocol = protocol
	}

	if currentMode, ok := dst.Spec.Config[v1alpha2.ProtocolSettingsContentMode]; ok {
		src.initializeProtocolSettingsIfNil()
		src.Spec.ProtocolSettings.ContentMode = &currentMode
	}
	if qos, ok := dst.Spec.Config[v1alpha2.ProtocolSettingsQos]; ok {
		src.initializeProtocolSettingsIfNil()
		src.Spec.ProtocolSettings.Qos = &qos
	}
	if exemptHandshake, ok := dst.Spec.Config[v1alpha2.ProtocolSettingsExemptHandshake]; ok {
		handshake, err := strconv.ParseBool(exemptHandshake)
		if err != nil {
			handshake = true
		}
		src.initializeProtocolSettingsIfNil()
		src.Spec.ProtocolSettings.ExemptHandshake = &handshake
	}

	if authType, ok := dst.Spec.Config[v1alpha2.WebhookAuthType]; ok {
		src.initializeWebhookAuthIfNil()
		src.Spec.ProtocolSettings.WebhookAuth.Type = authType
	}
	if grantType, ok := dst.Spec.Config[v1alpha2.WebhookAuthGrantType]; ok {
		src.initializeWebhookAuthIfNil()
		src.Spec.ProtocolSettings.WebhookAuth.GrantType = grantType
	}
	if clientID, ok := dst.Spec.Config[v1alpha2.WebhookAuthClientID]; ok {
		src.initializeWebhookAuthIfNil()
		src.Spec.ProtocolSettings.WebhookAuth.ClientID = clientID
	}
	if secret, ok := dst.Spec.Config[v1alpha2.WebhookAuthClientSecret]; ok {
		src.initializeWebhookAuthIfNil()
		src.Spec.ProtocolSettings.WebhookAuth.ClientSecret = secret
	}
	if token, ok := dst.Spec.Config[v1alpha2.WebhookAuthTokenURL]; ok {
		src.initializeWebhookAuthIfNil()
		src.Spec.ProtocolSettings.WebhookAuth.TokenURL = token
	}
	if scope, ok := dst.Spec.Config[v1alpha2.WebhookAuthScope]; ok {
		src.initializeWebhookAuthIfNil()
		src.Spec.ProtocolSettings.WebhookAuth.Scope = strings.Split(scope, ",")
	}
}

// setV2SpecTypes sets event types in the Subscription Spec in the v1alpha2 way.
func (src *Subscription) setV2SpecTypes(dst *v1alpha2.Subscription) error {
	if v1alpha1TypeCleaner == nil {
		return errors.New("event type cleaner is not initialized")
	}

	if src.Spec.Filter != nil {
		for _, filter := range src.Spec.Filter.Filters {
			if dst.Spec.Source == "" {
				dst.Spec.Source = filter.EventSource.Value
			}
			if dst.Spec.Source != "" && filter.EventSource.Value != dst.Spec.Source {
				return errors.New(ErrorMultipleSourceMsg)
			}
			// clean the type and merge segments if needed
			cleanedType, err := v1alpha1TypeCleaner.Clean(filter.EventType.Value)
			if err != nil {
				return err
			}

			// add the type to spec
			dst.Spec.Types = append(dst.Spec.Types, cleanedType)
		}
	}
	return nil
}

// natsSpecConfigToV2 converts the v1alpha2 Spec config to v1alpha1.
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
		if dst.Spec.Config == nil {
			dst.Spec.Config = map[string]string{}
		}
		dst.Spec.Config[v1alpha2.MaxInFlightMessages] = fmt.Sprint(src.Spec.Config.MaxInFlightMessages)
	}
}

// setBEBBackendStatus moves the BEB-related to Backend fields of the Status in the v1alpha2.
func (src *Subscription) bebBackendStatusToV1(dst *v1alpha2.Subscription) {
	src.Status.Ev2hash = dst.Status.Backend.Ev2hash
	src.Status.Emshash = dst.Status.Backend.EventMeshHash
	src.Status.ExternalSink = dst.Status.Backend.ExternalSink
	src.Status.FailedActivation = dst.Status.Backend.FailedActivation
	src.Status.APIRuleName = dst.Status.Backend.APIRuleName
	if dst.Status.Backend.EventMeshSubscriptionStatus != nil {
		src.Status.EmsSubscriptionStatus = &EmsSubscriptionStatus{
			SubscriptionStatus:       dst.Status.Backend.EventMeshSubscriptionStatus.Status,
			SubscriptionStatusReason: dst.Status.Backend.EventMeshSubscriptionStatus.StatusReason,
			LastSuccessfulDelivery:   dst.Status.Backend.EventMeshSubscriptionStatus.LastSuccessfulDelivery,
			LastFailedDelivery:       dst.Status.Backend.EventMeshSubscriptionStatus.LastFailedDelivery,
			LastFailedDeliveryReason: dst.Status.Backend.EventMeshSubscriptionStatus.LastFailedDeliveryReason,
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

// setV1CleanEvenTypes sets the clean event types to v1alpha1 Subscription Status.
func (src *Subscription) setV1CleanEvenTypes(dst *v1alpha2.Subscription) {
	src.Status.InitializeCleanEventTypes()
	for _, eventType := range dst.Status.Types {
		src.Status.CleanEventTypes = append(src.Status.CleanEventTypes, eventType.CleanType)
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
