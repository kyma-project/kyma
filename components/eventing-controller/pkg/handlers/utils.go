package handlers

import (
	"fmt"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"strings"

	types2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/types"
	"github.com/mitchellh/hashstructure"
)

func getHash(subscription *types2.Subscription) (uint64, error) {
	if hash, err := hashstructure.Hash(subscription, nil); err != nil {
		return 0, err
	} else {
		return hash, nil
	}
}

func getInternalView4Ev2(subscription *eventingv1alpha1.Subscription) (*types2.Subscription, error) {
	emsSubscription := &types2.Subscription{} //false, too many default fields

	// Name
	emsSubscription.Name = subscription.Name
	emsSubscription.ContentMode = subscription.Spec.ProtocolSettings.ContentMode
	emsSubscription.ExemptHandshake = subscription.Spec.ProtocolSettings.ExemptHandshake

	// Qos
	qos := strings.ReplaceAll(subscription.Spec.ProtocolSettings.Qos, "-", "_")
	if qos == string(types2.QosAtLeastOnce) {
		emsSubscription.Qos = types2.QosAtLeastOnce
	} else if qos == string(types2.QosAtMostOnce) {
		emsSubscription.Qos = types2.QosAtMostOnce
	} else {
		return nil, fmt.Errorf("invalid Qos: %v", subscription.Spec.ProtocolSettings.Qos)
	}

	// WebhookUrl
	emsSubscription.WebhookUrl = subscription.Spec.Sink

	// Events
	for _, e := range subscription.Spec.Filter.Filters {
		s := e.EventSource.Value
		t := e.EventType.Value
		emsSubscription.Events = append(emsSubscription.Events, types2.Event{Source: s, Type: t})
	}

	// WebhookAuth
	auth := &types2.WebhookAuth{}
	auth.ClientID = subscription.Spec.ProtocolSettings.WebhookAuth.ClientId
	auth.ClientSecret = subscription.Spec.ProtocolSettings.WebhookAuth.ClientSecret
	if subscription.Spec.ProtocolSettings.WebhookAuth.GrantType == string(types2.GrantTypeClientCredentials) {
		auth.GrantType = types2.GrantTypeClientCredentials
	} else {
		return nil, fmt.Errorf("invalid GrantType: %v", subscription.Spec.ProtocolSettings.WebhookAuth.GrantType)
	}
	if subscription.Spec.ProtocolSettings.WebhookAuth.Type == string(types2.AuthTypeClientCredentials) {
		auth.Type = types2.AuthTypeClientCredentials
	} else {
		return nil, fmt.Errorf("invalid Type: %v", subscription.Spec.ProtocolSettings.WebhookAuth.Type)
	}
	auth.TokenURL = subscription.Spec.ProtocolSettings.WebhookAuth.TokenUrl
	emsSubscription.WebhookAuth = auth

	return emsSubscription, nil
}

func getInternalView4Ems(subscription *types2.Subscription) (*types2.Subscription, error) {
	emsSubscription := &types2.Subscription{}

	// Name
	emsSubscription.Name = subscription.Name
	emsSubscription.ContentMode = subscription.ContentMode
	emsSubscription.ExemptHandshake = subscription.ExemptHandshake

	// Qos
	emsSubscription.Qos = subscription.Qos

	// WebhookUrl
	emsSubscription.WebhookUrl = subscription.WebhookUrl

	// Events
	for _, e := range subscription.Events {
		s := e.Source
		t := e.Type
		emsSubscription.Events = append(emsSubscription.Events, types2.Event{Source: s, Type: t})
	}

	return emsSubscription, nil
}

func getHash4WebhookAuth(subscription *types2.Subscription) (uint64, error) {
	hash, err := hashstructure.Hash(subscription.WebhookAuth, nil)
	if err != nil {
		return 0, err
	}
	return hash, nil
}
