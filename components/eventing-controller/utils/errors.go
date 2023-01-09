package utils

import (
	"fmt"
)

// MakeSubscriptionError creates a new error and includes the underlyingError in the message
// for subscription-related errors.
func MakeSubscriptionError(actualError, underlyingError error, subscription any) error {
	return fmt.Errorf("%w: %v, subscription: %v", actualError, underlyingError, subscription)
}

// MakeConsumerError creates a new error and includes the underlyingError in the message
// for consumer-related errors.
func MakeConsumerError(actualError, underlyingError error, consumer any) error {
	return fmt.Errorf("%w: %v, consumer: %v", actualError, underlyingError, consumer)
}
