package jetstream

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrMissingSubscription = errors.New("failed to find a NATS subscription for a subject")
	ErrAddConsumer         = errors.New("failed to add a consumer")
	ErrGetConsumer         = errors.New("failed to get consumer info")
	ErrUpdateConsumer      = errors.New("failed to update consumer")
	ErrDeleteConsumer      = errors.New("failed to delete consumer")
	ErrFailedSubscribe     = errors.New("failed to create NATS JetStream subscription")
	ErrFailedUnsubscribe   = errors.New("failed to unsubscribe from NATS JetStream")

	ErrConnect           = errors.New("failed to connect to NATS JetStream")
	ErrEmptyStreamName   = errors.New("stream name cannot be empty")
	ErrStreamNameTooLong = fmt.Errorf("stream name should be max %d characters long", jsMaxStreamNameLength)
)
