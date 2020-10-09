package handlers

import (
	"fmt"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"net/http"
	"time"

	client2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/client"
	config2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/config"
	types2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/types"
	auth2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/auth"

	"github.com/go-logr/logr"
)

const (
	activateTimeout = time.Second * 10 // timeout for the subscription to be active
)

type Beb struct {
	Token  *auth2.AccessToken
	Client *client2.Client
	Log    logr.Logger
}

func (b *Beb) ChekAndUpdateEmsSubscription(subscription *eventingv1alpha1.Subscription, oldEv2Hash uint64, oldEmsHash uint64) (uint64, uint64, error) {
	// get the internal view of ev2 subscription
	sEv2, err := GetInternalView4Ev2(subscription)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ev2 subscription", "name:", subscription.Name)
		return 0, 0, err
	}
	newEv2Hash, err := GetHash(sEv2)
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
	sEms, err := GetInternalView4Ems(emsSubscription)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ems subscription", "subscription name:", emsSubscription.Name)
	}
	newEmsHash, err := GetHash(sEms)
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
	sEms, err := GetInternalView4Ems(emsSubscription)
	if err != nil {
		b.Log.Error(err, "failed to get internal view for ems subscription", "subscription name:", emsSubscription.Name)
	}
	newEmsHash, err := GetHash(sEms)
	if err != nil {
		b.Log.Error(err, "failed to get the hash value for ems subscription", "subscription", sEms.Name)
		return 0, err
	}
	return newEmsHash, nil
}

// TODO:
// wrap each function with Authenticate wrapper which should react to a returned 401 and should authenticate
// again and create a new valid token in Beb.Token

func (b *Beb) authenticate() (*auth2.AccessToken, error) {
	authenticator := auth2.NewAuthenticator(auth2.GetDefaultConfig())
	token, err := authenticator.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with error: %v", err)
	}
	b.Token = token
	b.Client = client2.NewClient(config2.GetDefaultConfig())
	return token, nil
}

func (b *Beb) getSubscription(name string) (*types2.Subscription, error) {
	emsSubscription, err := b.Client.Get(b.Token, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription with error: %v", err)
	}
	return emsSubscription, nil
}

func (b *Beb) deleteSubscription(name string) error {
	// delete subscription
	if _, err := b.Client.Delete(b.Token, name); err != nil {
		return fmt.Errorf("failed to delete subscription with error: %v", err)
	}
	return nil
}

func (b *Beb) createSubscription(subscription *types2.Subscription) error {
	createResponse, err := b.Client.Create(b.Token, subscription)
	if err != nil {
		return fmt.Errorf("failed to create subscription with error: %v", err)
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
	timeout := time.After(activateTimeout)
	tick := time.Tick(time.Millisecond * 500)
	for {
		select {
		case <-timeout:
			{
				return false
			}
		case <-tick:
			{
				sub, err := b.Client.Get(b.Token, name)
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
