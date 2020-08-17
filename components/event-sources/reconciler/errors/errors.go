// Package errors contains custom error types and error utilities for reconcilers.
package errors

import (
	"context"

	"go.uber.org/zap"
	"knative.dev/pkg/logging"
)

// Handle inspects the received error and returns either nil or the original
// error depending on the nature of that error. The context and message are
// used to produce log entries. Used to ensure a queued item is either requeued
// or dropped.
func Handle(err error, ctx context.Context, msg string) error {
	switch {
	case err == nil:
		return nil
	case IsSkippable(err):
		logging.FromContext(ctx).Debugw("[skipped] "+msg, zap.Error(err))
		return nil
	default:
		logging.FromContext(ctx).Errorw(msg, zap.Error(err))
		return err
	}
}
