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

import (
	"errors"
	"testing"
)

func TestSkippableError(t *testing.T) {
	nilErr := (error)(nil)
	if NewSkippable(nilErr).Error() != "" {
		t.Error("Expected empty string")
	}

	const errStr = "some error"

	nonNilErr := errors.New(errStr)
	if NewSkippable(nonNilErr).Error() != errStr {
		t.Errorf("Expected error string to be %q", errStr)
	}
}
