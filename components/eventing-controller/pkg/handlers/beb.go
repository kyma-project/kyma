package handlers

import (
	"fmt"
	"net/http"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/auth"

	"github.com/go-logr/logr"
)

// compile time check
var _ Interface = &Beb{}

type Interface interface {
	Initialize()
	SyncBebSubscription(subscription *eventingv1alpha1.Subscription) (bool, error)
	DeleteBebSubscription(subscription *eventingv1alpha1.Subscription) error
}

type Beb struct {
	Client *client.Client
	Log    logr.Logger
}

type BebResponse struct {
	StatusCode int
	Error      error
}

func (b *Beb) Initialize() {
	if b.Client == nil {
		authenticator := auth.NewAuthenticator()
		b.Client = client.NewClient(config.GetDefaultConfig(), authenticator)
	}
}

// SyncBebSubscription synchronize the EV@ subscription with the EMS subscription. It returns true, if the EV2 susbcription status was changed
func (b *Beb) SyncBebSubscription(subscription *eventingv1alpha1.Subscription) (bool, error) {
	// get the internal view for the ev2 subscription
	var statusChanged = false
	sEv2, err := getInternalView4Ev2(subscription)
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
		emsSubscription, newEmsHash, err = b.deleteCreateAndHashSubscription(sEv2)
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
			emsSubscription, newEmsHash, err = b.deleteCreateAndHashSubscription(sEv2)
			if err != nil {
				return false, err
			}
			subscription.Status.Emshash = newEmsHash
			statusChanged = true
		}
	}
	// set the status of emsSubscription in ev2Subscription
	statusChanged = b.setEmsSubscritionStatus(subscription, emsSubscription) || statusChanged

	return statusChanged, nil
}

// DeleteBebSubscription deletes the corresponding EMS subscription
func (b *Beb) DeleteBebSubscription(subscription *eventingv1alpha1.Subscription) error {
	sEv2, err := getInternalView4Ev2(subscription)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ev2 subscription", "name:", subscription.Name)
		return err
	}
	return b.deleteSubscription(sEv2.Name)
}

func (b *Beb) deleteCreateAndHashSubscription(subscription *types.Subscription) (*types.Subscription, int64, error) {
	// delete Ems susbcription
	if err := b.deleteSubscription(subscription.Name); err != nil {
		b.Log.Error(err, "delete ems subscription failed", "subscription name:", subscription.Name)
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

// Set the status of emsSubscription in ev2Subscription
func (b *Beb) setEmsSubscritionStatus(subscription *eventingv1alpha1.Subscription, emsSubscription *types.Subscription) bool {
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
