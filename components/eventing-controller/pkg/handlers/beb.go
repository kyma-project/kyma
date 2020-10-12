package handlers

import (
	"fmt"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"net/http"
	"time"

	client2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/client"
	config2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/config"
	types2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/types"
	auth2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/auth"

	"github.com/go-logr/logr"
)

// compile time check
var _ Interface = &Beb{}

type Interface interface {
	Initialize()
	SyncBebSubscription(subscription *eventingv1alpha1.Subscription, oldEv2Hash uint64, oldEmsHash uint64) (uint64, uint64, error)
	DeleteBebSubscription(subscription *eventingv1alpha1.Subscription) error
}

type Beb struct {
	Token  *auth2.AccessToken
	Client *client2.Client
	Log    logr.Logger
}

type BebResponse struct {
	StatusCode int
	Error    error
}

func (b *Beb) Initialize() {
	if b.Token == nil {
		if err := b.authenticate(); err != nil {
			b.Log.Error(err, "failed to authenticate")
		}
	}
}

func (b *Beb) SyncBebSubscription(subscription *eventingv1alpha1.Subscription, oldEv2Hash uint64, oldEmsHash uint64) (uint64, uint64, error) {
	// get the internal view for the ev2 subscription
	sEv2, err := getInternalView4Ev2(subscription)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ev2 subscription", "name:", subscription.Name)
		return 0, 0, err
	}
	newEv2Hash, err := getHash(sEv2)
	if err != nil {
		b.Log.Error(err, "failed to get the hash value", "subscription name", sEv2.Name)
		return 0, 0, err
	}
	if newEv2Hash != oldEv2Hash {
		// delete & create a new Ems subscription
		newEmsHash, err := b.deleteCreateAndHashSubscription(sEv2)
		if err != nil {
			return 0, 0, err
		}
		return newEv2Hash, newEmsHash, nil
	}
	// no change on ev2 side
	// check if ems subscription is the same as in the past
	emsSubscription, err := b.getSubscription(sEv2.Name)
	if err != nil {
		b.Log.Error(err, "failed to get ems subscription", "subscription name", sEv2.Name)
		return 0, 0, err
	}
	// get the internal view for the ems subscription
	sEms, err := getInternalView4Ems(emsSubscription)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ems subscription", "subscription name:", emsSubscription.Name)
	}
	newEmsHash, err := getHash(sEms)
	if err != nil {
		b.Log.Error(err, "failed to get the hash value for ems subscription", "subscription", sEms.Name)
		return 0, 0, err
	}
	if newEmsHash != oldEmsHash {
		// delete & create a new Ems subscription
		newEmsHash, err = b.deleteCreateAndHashSubscription(sEv2)
		if err != nil {
			return 0, 0, err
		}
		return oldEv2Hash, newEmsHash, err
	}
	// returns, no changes
	return oldEv2Hash, oldEmsHash, nil
}

func (b *Beb) DeleteBebSubscription(subscription *eventingv1alpha1.Subscription) error {
	sEv2, err := getInternalView4Ev2(subscription)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ev2 subscription", "name:", subscription.Name)
		return err
	}
	return b.deleteSubscription(sEv2.Name)
}

func (b *Beb) deleteCreateAndHashSubscription(subscription *types2.Subscription) (uint64, error) {
	// delete Ems susbcription
	if err := b.deleteSubscription(subscription.Name); err != nil {
		b.Log.Error(err, "delete ems subscription failed", "subscription name:", subscription.Name)
		return 0, err
	}
	// create a new Ems subscription
	if err := b.createSubscription(subscription); err != nil {
		b.Log.Error(err, "create ems subscription failed", "subscription name:", subscription.Name)
		return 0, err
	}
	// get the new Ems subscription
	emsSubscription, err := b.getSubscription(subscription.Name)
	if err != nil {
		b.Log.Error(err, "failed to get ems subscription", "subscription name", subscription.Name)
		return 0, err
	}
	// get the new hash
	sEms, err := getInternalView4Ems(emsSubscription)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ems subscription", "subscription name:", emsSubscription.Name)
	}
	newEmsHash, err := getHash(sEms)
	if err != nil {
		b.Log.Error(err, "failed to get the hash value for ems subscription", "subscription", sEms.Name)
		return 0, err
	}
	return newEmsHash, nil
}

func (b *Beb) authenticate() error {
	b.Log.Info("BEB authenticate()")
	authenticator := auth2.NewAuthenticator(auth2.GetDefaultConfig())
	token, err := authenticator.Authenticate()
	if err != nil {
		return fmt.Errorf("failed to authenticate with error: %v", err)
	}
	b.Token = token
	b.Client = client2.NewClient(config2.GetDefaultConfig())
	return nil
}

func (b *Beb) getSubscription(name string) (*types2.Subscription, error) {
	b.Log.Info("BEB getSubscription()","subscription name:", name)
	emsSubscription, resp, err := b.Client.Get(b.Token, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription with error: %v", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		b.Token = nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get subscription with error: %v; %v", resp.StatusCode, resp.Message)
	}
	return emsSubscription, nil
}

func (b *Beb) deleteSubscription(name string) error {
	b.Log.Info("BEB deleteSubscription()","subscription name:", name)
	resp, err := b.Client.Delete(b.Token, name)
	if err != nil {
		return fmt.Errorf("failed to delete subscription with error: %v", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		b.Token = nil
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to delete subscription with error: %v", resp.StatusCode)
	}
	return nil
}

func (b *Beb) createSubscription(subscription *types2.Subscription) error {
	b.Log.Info("BEB createSubscription()","subscription name:", subscription.Name)
	createResponse, err := b.Client.Create(b.Token, subscription)
	if err != nil {
		return fmt.Errorf("failed to create subscription with error: %v", err)
	}
	if createResponse.StatusCode == http.StatusUnauthorized {
		b.Token = nil
	}
	if createResponse.StatusCode > http.StatusAccepted && createResponse.StatusCode != http.StatusConflict {
		return fmt.Errorf("failed to create subscription with error: %v", createResponse)
	}
	if !b.waitForSubscriptionActive(subscription.Name) {
		return fmt.Errorf("timeout waiting for the subscription to be active: %v", subscription.Name)
	}
	return nil
}

func (b *Beb) waitForSubscriptionActive(name string) bool {
	timeout := time.After(env.GetConfig().WebhookActivationTimeout)
	tick := time.Tick(time.Millisecond * 500)
	for {
		select {
		case <-timeout:
			{
				return false
			}
		case <-tick:
			{
				sub, _, err := b.Client.Get(b.Token, name)
				if err != nil {
					return false
				}
				if sub != nil && sub.SubscriptionStatus == types2.SubscriptionStatusActive {
					return true
				}
			}
		}
	}
}
