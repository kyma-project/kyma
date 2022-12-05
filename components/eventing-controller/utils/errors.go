package utils

import "fmt"

// MakeError creates a new error and includes the underlyingError in the message.
// However, it does not expose/wrap the underlyingError.
//
// Following the recommendation of https://go.dev/blog/go1.13-errors,
// the actualError is encapsulated into a new error and not returned directly.
// This forces callers to use
// errors.Is(err, pkg.ErrPermission) instead of
// err == pkg.ErrPermission { … }.
func MakeError(actualError, underlyingError error) error {
	return fmt.Errorf("%w: %v", actualError, underlyingError)
}

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
