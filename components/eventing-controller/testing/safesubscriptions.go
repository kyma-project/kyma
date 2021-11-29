package testing

import (
	"strings"
	"sync"

	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

// SafeSubscriptions encapsulates Subscriptions to provide mutual exclusion.
type SafeSubscriptions struct {
	mutex         *sync.RWMutex
	subscriptions map[string]*bebtypes.Subscription
}

// NewSafeSubscriptions returns a new SafeSubscriptions.
func NewSafeSubscriptions() *SafeSubscriptions {
	return &SafeSubscriptions{
		&sync.RWMutex{}, make(map[string]*bebtypes.Subscription),
	}
}

// GetSubscriptions returns a Subscription via the corresponding the key.
func (s *SafeSubscriptions) GetSubscriptions(key string) *bebtypes.Subscription {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.subscriptions[key]
}

// DeleteSubscriptions deletes a Subscription via the corresponding key.
func (s *SafeSubscriptions) DeleteSubscriptions(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.subscriptions, key)
}

// DeleteSubscriptionsByName deletes all Subscriptions that contain the substring 'name' in their own name.
func (s *SafeSubscriptions) DeleteSubscriptionsByName(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for k := range s.subscriptions {
		if strings.Contains(k, name) {
			delete(s.subscriptions, k)
		}
	}
}

// PutSubscriptions sets a Subscription of SafeSubscriptions via the corresponding key.
func (s *SafeSubscriptions) PutSubscriptions(key string, subscription *bebtypes.Subscription) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.subscriptions[key] = subscription
}
