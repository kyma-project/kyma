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
	"context"
	"errors"
	"testing"
)

func TestHandle(t *testing.T) {
	dummyErr := errors.New("error")

	testCases := map[string]struct {
		input, expect error
	}{
		"Nil input": {
			nil, nil,
		},
		"Skippable error gets ignored": {
			NewSkippable(dummyErr), nil,
		},
		"Non-skippable error gets returned": {
			dummyErr, dummyErr,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(*testing.T) {
			ctx := context.Background()
			if expect, got := tc.expect, Handle(tc.input, ctx, ""); got != expect {
				t.Errorf("Expected %v, got %v", expect, got)
			}
		})
	}
}
