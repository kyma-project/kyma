package errors

import (
	"fmt"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"golang.org/x/xerrors"
)

type ErrMissingNATSSubscription = MissingNATSSubscriptionError

type MissingNATSSubscriptionError struct {
	Subject eventingv1alpha2.EventType
}

func NewMissingNatsSubscriptionError(subject eventingv1alpha2.EventType) error {
	return xerrors.Errorf("%w", &ErrMissingNATSSubscription{Subject: subject})
}

func (e *MissingNATSSubscriptionError) Error() string {
	return fmt.Sprintf("failed to create NATS JetStream subscription for subject: %v", e.Subject)
}

func (e *MissingNATSSubscriptionError) Is(target error) bool {
	if _, ok := target.(*MissingNATSSubscriptionError); !ok {
		return false
	}
	return true
}
