package jetstreamv2

import "github.com/pkg/errors"

var (
	ErrMissingSubscription = errors.New("failed to find a NATS subscription for a subject")
	ErrAddConsumer         = errors.New("failed to add a consumer")
	ErrGetConsumer         = errors.New("failed to get consumer info")
	ErrUpdateConsumer      = errors.New("failed to update consumer")
	ErrFailedSubscribe     = errors.New("failed to create NATS JetStream subscription")
	ErrFailedUnsubscribe   = errors.New("failed to unsubscribe from NATS JetStream")
)
