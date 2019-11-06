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

package errors

// skippable is a type of error that can skip reconciliation retries.
type skippable interface {
	Skip()
}

// IsSkippable returns whether the given error is skippable.
func IsSkippable(e error) bool {
	_, ok := e.(skippable)
	return ok
}

// NewSkippable wraps an error into a skippable type so it gets ignored by the
// reconciler.
func NewSkippable(e error) error {
	return &skipRetry{err: e}
}

type skipRetry struct{ err error }

// type skipRetry implements skippable.
var _ skippable = &skipRetry{}

// Skip implements skippable.
func (*skipRetry) Skip() {}

// Error implements the error interface.
func (e *skipRetry) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}
