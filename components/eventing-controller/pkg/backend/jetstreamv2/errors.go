package jetstreamv2

import (
	"errors"
	"fmt"
)

var (
	ErrMissingSubscription = errors.New("failed to find a NATS subscription for a subject")
	ErrAddConsumer         = errors.New("failed to add a consumer")
	ErrGetConsumer         = errors.New("failed to get consumer info")
	ErrUpdateConsumer      = errors.New("failed to update consumer")
	ErrDeleteConsumer      = errors.New("failed to delete consumer")
	ErrFailedSubscribe     = errors.New("failed to create NATS JetStream subscription")
	ErrFailedUnsubscribe   = errors.New("failed to unsubscribe from NATS JetStream")
	ErrAddStream           = errors.New("failed to add stream")
	ErrConfig              = errors.New("failed to parse/validate config")

	ErrEmptyStreamName    = errors.New("stream name cannot be empty")
	ErrStreamNameTooLong  = fmt.Errorf("stream name should be max %d characters long", jsMaxStreamNameLength)
	ErrConnect            = errors.New("failed to connect to NATS JetStream")
	ErrContext            = errors.New("failed to build JetStream context")
	ErrUpdateStreamConfig = errors.New("failed to updateConsumer stream config")
	ErrCEClient           = errors.New("failed to build a cloud event client")

	// ErrUnknown is a wildcard error which is returned for all cases where the other errors don't match.
	ErrUnknown = errors.New("failed to talk to NATS")
)
