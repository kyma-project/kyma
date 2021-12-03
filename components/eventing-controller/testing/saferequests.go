package testing

import (
	"net/http"
	"sync"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

// SafeRequests encapsulates Requests to provide mutual exclusion.
type SafeRequests struct {
	sync.RWMutex
	requests map[*http.Request]interface{}
}

// NewSafeRequests returns a new instance of SafeRequests.
func NewSafeRequests() *SafeRequests {
	return &SafeRequests{
		sync.RWMutex{},
		make(map[*http.Request]interface{}),
	}
}

// StoreRequest adds a request to requests and sets it's corresponding subscription to nil.
func (r *SafeRequests) StoreRequest(request *http.Request) {
	r.Lock()
	defer r.Unlock()
	r.requests[request] = nil
}

// PutSubscription sets a the subscription of a request.
func (r *SafeRequests) PutSubscription(request *http.Request, subscription types.Subscription) {
	r.Lock()
	defer r.Unlock()
	r.requests[request] = subscription
}

// Len returns the length of requests.
func (r *SafeRequests) Len() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.requests)
}

// CheckIfAny iterates over requests and checks if a given func f is true for any iteration's request and payload.
// CheckIfAny is only read-safe; f must be a read-only operation.
func (r *SafeRequests) CheckIfAny(f func(request *http.Request, payload interface{}) bool) bool {
	r.RLock()
	defer r.RUnlock()
	for req, payload := range r.requests {
		if f(req, payload) {
			return true
		}
	}
	return false
}

// ReadEach iterates over requests and executes a given func f with each iteration's request and payload.
func (r *SafeRequests) ReadEach(f func(request *http.Request, payload interface{})) {
	r.RLock()
	defer r.RUnlock()
	for req, payload := range r.requests {
		f(req, payload)
	}
}
