package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/auth"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
)

// compile time check
var _ MessagingBackend = &Beb{}

type Beb struct {
	Client           *client.Client
	WebhookAuth      *types.WebhookAuth
	ProtocolSettings *eventingv1alpha1.ProtocolSettings
	Namespace        string
	Log              logr.Logger
}

type BebResponse struct {
	StatusCode int
	Error      error
}

func (b *Beb) Initialize(cfg env.Config) error {
	if b.Client == nil {
		authenticator := auth.NewAuthenticator(cfg)
		b.Client = client.NewClient(config.GetDefaultConfig(cfg.BebApiUrl), authenticator)
		b.WebhookAuth = getWebHookAuth(cfg)
		b.ProtocolSettings = &eventingv1alpha1.ProtocolSettings{
			ContentMode:     &cfg.ContentMode,
			ExemptHandshake: &cfg.ExemptHandshake,
			Qos:             &cfg.Qos,
		}
		b.Namespace = cfg.BEBNamespace
	}
	return nil
}

// getWebHookAuth returns the webhook auth config from the given env config
// or returns an error if the env config contains invalid grant type or auth type.
func getWebHookAuth(cfg env.Config) *types.WebhookAuth {
	return &types.WebhookAuth{
		ClientID:     cfg.WebhookClientID,
		ClientSecret: cfg.WebhookClientSecret,
		TokenURL:     cfg.WebhookTokenEndpoint,
		Type:         types.AuthTypeClientCredentials,
		GrantType:    types.GrantTypeClientCredentials,
	}
}

// SyncSubscription synchronize the EV2 subscription with the EMS subscription. It returns true, if the EV2 subscription status was changed
func (b *Beb) SyncSubscription(subscription *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner, params ...interface{}) (bool, error) {
	apiRule, ok := params[0].(*apigatewayv1alpha1.APIRule)
	if !ok {
		err := fmt.Errorf("failed to get apiRule from params[0]: %v", params[0])
		b.Log.Error(err, "wrong parameter for subscription", "name:", subscription.Name)
	}

	// get the internal view for the ev2 subscription
	var statusChanged = false
	sEv2, err := getInternalView4Ev2(subscription, apiRule, b.WebhookAuth, b.ProtocolSettings, b.Namespace)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ev2 subscription", "name:", subscription.Name)
		return false, err
	}
	newEv2Hash, err := getHash(sEv2)
	if err != nil {
		b.Log.Error(err, "failed to get the hash value", "subscription name", sEv2.Name)
		return false, err
	}
	var emsSubscription *types.Subscription
	// check the hash values for ev2 and ems
	if newEv2Hash != subscription.Status.Ev2hash {
		// delete & create a new Ems subscription
		var newEmsHash int64
		emsSubscription, newEmsHash, err = b.deleteCreateAndHashSubscription(sEv2, cleaner)
		if err != nil {
			return false, err
		}
		subscription.Status.Ev2hash = newEv2Hash
		subscription.Status.Emshash = newEmsHash
		statusChanged = true
	} else {
		// check if ems subscription is the same as in the past
		emsSubscription, err = b.getSubscription(sEv2.Name)
		if err != nil {
			b.Log.Error(err, "failed to get ems subscription", "subscription name", sEv2.Name)
			return false, err
		}
		// get the internal view for the ems subscription
		sEms, err := getInternalView4Ems(emsSubscription)
		if err != nil {
			b.Log.Error(err, "failed to get internal view for ems subscription", "subscription name:", emsSubscription.Name)
			return false, err
		}
		newEmsHash, err := getHash(sEms)
		if err != nil {
			b.Log.Error(err, "failed to get the hash value for ems subscription", "subscription", sEms.Name)
			return false, err
		}
		if newEmsHash != subscription.Status.Emshash {
			// delete & create a new Ems subscription
			emsSubscription, newEmsHash, err = b.deleteCreateAndHashSubscription(sEv2, cleaner)
			if err != nil {
				return false, err
			}
			subscription.Status.Emshash = newEmsHash
			statusChanged = true
		}
	}
	// set the status of emsSubscription in ev2Subscription
	statusChanged = b.setEmsSubscriptionStatus(subscription, emsSubscription) || statusChanged

	return statusChanged, nil
}

// DeleteSubscription deletes the corresponding EMS subscription
func (b *Beb) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	return b.deleteSubscription(subscription.Name)
}

func (b *Beb) deleteCreateAndHashSubscription(subscription *types.Subscription, cleaner eventtype.Cleaner) (*types.Subscription, int64, error) {
	// delete Ems subscription
	if err := b.deleteSubscription(subscription.Name); err != nil {
		b.Log.Error(err, "delete ems subscription failed", "subscription name:", subscription.Name)
		return nil, 0, err
	}

	// clean the application name segment in the subscription event-types from none-alphanumeric characters
	if err := cleanEventTypes(subscription, cleaner); err != nil {
		b.Log.Error(err, "clean application name in the subscription event-types failed", "subscription name:", subscription.Name)
		return nil, 0, err
	}

	// create a new Ems subscription
	if err := b.createSubscription(subscription); err != nil {
		b.Log.Error(err, "create ems subscription failed", "subscription name:", subscription.Name)
		return nil, 0, err
	}

	// get the new Ems subscription
	emsSubscription, err := b.getSubscription(subscription.Name)
	if err != nil {
		b.Log.Error(err, "failed to get ems subscription", "subscription name", subscription.Name)
		return nil, 0, err
	}

	// get the new hash
	sEms, err := getInternalView4Ems(emsSubscription)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ems subscription", "subscription name:", emsSubscription.Name)
	}
	newEmsHash, err := getHash(sEms)
	if err != nil {
		b.Log.Error(err, "failed to get the hash value for ems subscription", "subscription", sEms.Name)
		return nil, 0, err
	}

	return emsSubscription, newEmsHash, nil
}

// cleanEventTypes cleans the application name segment in the subscription event-types from none-alphanumeric characters
// note: the given subscription instance will be updated with the cleaned event-types
func cleanEventTypes(subscription *types.Subscription, cleaner eventtype.Cleaner) error {
	events := make(types.Events, 0, len(subscription.Events))
	for _, event := range subscription.Events {
		eventType, err := cleaner.Clean(event.Type)
		if err != nil {
			return err
		}
		events = append(events, types.Event{Source: event.Source, Type: eventType})
	}
	subscription.Events = events
	return nil
}

// setEmsSubscriptionStatus sets the status of emsSubscription in ev2Subscription
func (b *Beb) setEmsSubscriptionStatus(subscription *eventingv1alpha1.Subscription, emsSubscription *types.Subscription) bool {
	var statusChanged = false
	if subscription.Status.EmsSubscriptionStatus.SubscriptionStatus != string(emsSubscription.SubscriptionStatus) {
		subscription.Status.EmsSubscriptionStatus.SubscriptionStatus = string(emsSubscription.SubscriptionStatus)
		statusChanged = true
	}
	if subscription.Status.EmsSubscriptionStatus.SubscriptionStatusReason != emsSubscription.SubscriptionStatusReason {
		subscription.Status.EmsSubscriptionStatus.SubscriptionStatusReason = emsSubscription.SubscriptionStatusReason
		statusChanged = true
	}
	if subscription.Status.EmsSubscriptionStatus.LastSuccessfulDelivery != emsSubscription.LastSuccessfulDelivery {
		subscription.Status.EmsSubscriptionStatus.LastSuccessfulDelivery = emsSubscription.LastSuccessfulDelivery
		statusChanged = true
	}
	if subscription.Status.EmsSubscriptionStatus.LastFailedDelivery != emsSubscription.LastFailedDelivery {
		subscription.Status.EmsSubscriptionStatus.LastFailedDelivery = emsSubscription.LastFailedDelivery
		statusChanged = true
	}
	if subscription.Status.EmsSubscriptionStatus.LastFailedDeliveryReason != emsSubscription.LastFailedDeliveryReason {
		subscription.Status.EmsSubscriptionStatus.LastFailedDeliveryReason = emsSubscription.LastFailedDeliveryReason
		statusChanged = true
	}
	return statusChanged
}

func (b *Beb) getSubscription(name string) (*types.Subscription, error) {
	b.Log.Info("BEB getSubscription()", "subscription name:", name)
	emsSubscription, resp, err := b.Client.Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription with error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get subscription with error: %v; %v", resp.StatusCode, resp.Message)
	}
	return emsSubscription, nil
}

func (b *Beb) deleteSubscription(name string) error {
	b.Log.Info("BEB deleteSubscription()", "subscription name:", name)
	resp, err := b.Client.Delete(name)
	if err != nil {
		return fmt.Errorf("failed to delete subscription with error: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to delete subscription with error: %v; %v", resp.StatusCode, resp.Message)
	}
	return nil
}

func (b *Beb) createSubscription(subscription *types.Subscription) error {
	b.Log.Info("BEB createSubscription()", "subscription name:", subscription.Name)
	createResponse, err := b.Client.Create(subscription)
	if err != nil {
		return fmt.Errorf("failed to create subscription with error: %v", err)
	}
	if createResponse.StatusCode > http.StatusAccepted && createResponse.StatusCode != http.StatusConflict {
		return fmt.Errorf("failed to create subscription with error: %v", createResponse)
	}
	return nil
}
