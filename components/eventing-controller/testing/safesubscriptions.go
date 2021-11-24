package testing

import (
	"strings"
	"sync"

	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

// SafeSubscription encapsulates Subscriptions to provide mutual exclusion.
type SafeSubscription struct {
	mutex         *sync.RWMutex
	subscriptions map[string]*bebtypes.Subscription
}

// NewSafeSubscription returns a new SafeSubscription.
func NewSafeSubscription() *SafeSubscription {
	return &SafeSubscription{
		&sync.RWMutex{}, make(map[string]*bebtypes.Subscription),
	}
}

// GetSubscription returns a Subscription via the corresponding the key.
func (s *SafeSubscription) GetSubscription(key string) *bebtypes.Subscription {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.subscriptions[key]
}

// DeleteSubscription deletes a Subscription via the corresponding key.
func (s *SafeSubscription) DeleteSubscription(key string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	delete(s.subscriptions, key)
}

// DeleteSubscriptionsByName deletes all Subscriptions that contain the substring 'name' in their own name.
func (s *SafeSubscription) DeleteSubscriptionsByName(name string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for k := range s.subscriptions {
		if strings.Contains(k, name) {
			delete(s.subscriptions, k)
		}
	}
}

// PutSubscription sets a Subscription of SafeSubscription via the corresponding key.
func (s *SafeSubscription) PutSubscription(key string, subscription *bebtypes.Subscription) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.subscriptions[key] = subscription
}
