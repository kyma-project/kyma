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
