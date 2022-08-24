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
	dst.Spec.ProtocolSettings = src.Spec.ProtocolSettings
	dst.Spec.Sink = src.Spec.Sink
	dst.Spec.TypeMatching = v1alpha2.EXACT

	for _, filter := range src.Spec.Filter.Filters {
		if dst.Spec.Source == "" {
			dst.Spec.Source = filter.EventSource.Value
		}
		dst.Spec.Types = append(dst.Spec.Types, filter.EventType.Value)
	}

	if src.Spec.Config != nil {
		dst.Spec.Config = map[string]interface{}{
			"maxInFlightMessages": src.Spec.Config.MaxInFlightMessages,
		}
	}

	// Status

	dst.Status.Conditions = src.Status.Conditions
	dst.Status.Ready = src.Status.Ready
	dst.Status.Types = src.Status.CleanEventTypes
	dst.Status.Ev2hash = src.Status.Ev2hash
	dst.Status.Emshash = src.Status.Emshash
	dst.Status.ExternalSink = src.Status.ExternalSink
	dst.Status.FailedActivation = src.Status.FailedActivation
	dst.Status.APIRuleName = src.Status.APIRuleName
	dst.Status.EmsSubscriptionStatus = src.Status.EmsSubscriptionStatus

	return nil
}
