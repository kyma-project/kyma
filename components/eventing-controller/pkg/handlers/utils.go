package handlers

import (
	"fmt"
	"strings"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/mitchellh/hashstructure"
)

func getHash(subscription *types.Subscription) (int64, error) {
	if hash, err := hashstructure.Hash(subscription, nil); err != nil {
		return 0, err
	} else {
		return int64(hash), nil
	}
}

func getInternalView4Ev2(subscription *eventingv1alpha1.Subscription) (*types.Subscription, error) {
	emsSubscription := &types.Subscription{}

	// Name
	emsSubscription.Name = subscription.Name
	emsSubscription.ContentMode = subscription.Spec.ProtocolSettings.ContentMode
	emsSubscription.ExemptHandshake = subscription.Spec.ProtocolSettings.ExemptHandshake

	// Qos
	qos := strings.ReplaceAll(subscription.Spec.ProtocolSettings.Qos, "-", "_")
	if qos == string(types.QosAtLeastOnce) {
		emsSubscription.Qos = types.QosAtLeastOnce
	} else if qos == string(types.QosAtMostOnce) {
		emsSubscription.Qos = types.QosAtMostOnce
	} else {
		return nil, fmt.Errorf("invalid Qos: %v", subscription.Spec.ProtocolSettings.Qos)
	}

	// WebhookUrl
	emsSubscription.WebhookUrl = subscription.Spec.Sink

	// Events
	for _, e := range subscription.Spec.Filter.Filters {
		s := e.EventSource.Value
		t := e.EventType.Value
		emsSubscription.Events = append(emsSubscription.Events, types.Event{Source: s, Type: t})
	}

	// WebhookAuth
	auth := &types.WebhookAuth{}
	auth.ClientID = subscription.Spec.ProtocolSettings.WebhookAuth.ClientId
	auth.ClientSecret = subscription.Spec.ProtocolSettings.WebhookAuth.ClientSecret
	if subscription.Spec.ProtocolSettings.WebhookAuth.GrantType == string(types.GrantTypeClientCredentials) {
		auth.GrantType = types.GrantTypeClientCredentials
	} else {
		return nil, fmt.Errorf("invalid GrantType: %v", subscription.Spec.ProtocolSettings.WebhookAuth.GrantType)
	}
	if subscription.Spec.ProtocolSettings.WebhookAuth.Type == string(types.AuthTypeClientCredentials) {
		auth.Type = types.AuthTypeClientCredentials
	} else {
		return nil, fmt.Errorf("invalid Type: %v", subscription.Spec.ProtocolSettings.WebhookAuth.Type)
	}
	auth.TokenURL = subscription.Spec.ProtocolSettings.WebhookAuth.TokenUrl
	emsSubscription.WebhookAuth = auth

	return emsSubscription, nil
}

func getInternalView4Ems(subscription *types.Subscription) (*types.Subscription, error) {
	emsSubscription := &types.Subscription{}

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
		emsSubscription.Events = append(emsSubscription.Events, types.Event{Source: s, Type: t})
	}

	return emsSubscription, nil
}
