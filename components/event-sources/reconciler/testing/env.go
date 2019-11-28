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

package testing

import (
	"os"
	"testing"
)

// SetEnvVar sets the value of an env var and returns a function that can be
// deferred to unset that variable.
func SetEnvVar(t *testing.T, name, val string) (unset func()) {
	t.Helper()

	if err := os.Setenv(name, val); err != nil {
		t.Errorf("Failed to set env var %s: %v", name, err)
	}

	return func() {
		_ = os.Unsetenv(name)
	}
}
