package errors

import (
	"fmt"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
)

type MissingSubscriptionError struct {
	Subject eventingv1alpha2.EventType
}

func NewMissingSubscriptionError(subject eventingv1alpha2.EventType) error {
	return fmt.Errorf("%w", &MissingSubscriptionError{Subject: subject})
}

func (e *MissingSubscriptionError) Error() string {
	return fmt.Sprintf("failed to find a NATS subscription for subject: %v", e.Subject)
}

func (e *MissingSubscriptionError) Is(target error) bool {
	if _, ok := target.(*MissingSubscriptionError); !ok {
		return false
	}
	return true
}
