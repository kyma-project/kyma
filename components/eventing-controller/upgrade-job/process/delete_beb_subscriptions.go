package process

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	pkgerrors "github.com/pkg/errors"
)

var _ Step = &DeleteBebSubscriptions{}

// DeleteBebSubscriptions struct implements the interface Step
type DeleteBebSubscriptions struct {
	name    string
	process *Process
}

// NewDeleteBebSubscriptions returns new instance of NewDeleteBebSubscriptions struct
func NewDeleteBebSubscriptions(p *Process) DeleteBebSubscriptions {
	return DeleteBebSubscriptions{
		name:    "Delete BEB subscriptions from Event-Mesh",
		process: p,
	}
}

// ToString returns step name
func (s DeleteBebSubscriptions) ToString() string {
	return s.name
}

// Do deletes the subscriptions from Event Mesh (BEB)
func (s DeleteBebSubscriptions) Do() error {
	// First check if BEB client is initialized or not
	if !s.process.Clients.EventMesh.IsInitialised() {
		return errors.New("event mesh (BEB) client is not initialised")
	}

	// Traverse through the subscriptions and migrate
	for _, subscription := range s.process.State.FilteredSubscriptions.Items {
		s.process.Logger.WithContext().Info("Deleting Event Mesh (BEB) Subscription: ", subscription.Name)
		err := s.DeleteBEBSubscription(subscription.Name)
		if err != nil {
			s.process.Logger.WithContext().Error(err)
			continue
		}
	}

	return nil
}

// DeleteBEBSubscription deletes the subscription with provided name from Event Mesh (BEB)
func (s DeleteBebSubscriptions) DeleteBEBSubscription(name string) error {
	maxAttempts := uint(10)
	delay := 15 * time.Second

	err := retry.Do(
		func() error {
			result, err := s.process.Clients.EventMesh.Delete(name)
			if err != nil {
				return pkgerrors.Wrapf(err, "failed to delete BEB subscription")
			}
			// If 404, then we don't retry as subscription is not on BEB
			if result.StatusCode == http.StatusNotFound {
				s.process.Logger.WithContext().Error(fmt.Sprintf("subscription: %s not found on Event Mesh (BEB)", name))
				return nil
			}

			// If status code is other than 404 and 2xx, then we return error and retry
			if !is2XXStatusCode(result.StatusCode) {
				return fmt.Errorf("response code is not 2xx, received response code is: %d for subscription: %s", result.StatusCode, name)
			}

			return nil
		},
		retry.Delay(delay),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(maxAttempts),
		retry.OnRetry(func(n uint, err error) {
			s.process.Logger.WithContext().Error("BEB subscription delete retry failed", err)
		}),
	)

	if err != nil {
		return pkgerrors.Wrapf(err, "failed to delete BEB subscription after retries")
	}

	return nil
}
