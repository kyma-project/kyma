/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
