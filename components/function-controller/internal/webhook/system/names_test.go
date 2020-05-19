package system

import (
	"os"
	"testing"
)

func TestGetNamespace(t *testing.T) {
	testcases := []struct {
		envVar       string
		expectEnvVar string
	}{{
		envVar:       "",
		expectEnvVar: DefaultNamespace,
	}, {
		envVar:       "test",
		expectEnvVar: "test",
	}}

	for _, ts := range testcases {
		if err := os.Setenv(SystemNamespaceEnvVar, ts.envVar); err != nil {
			t.Fatalf("Failed to set ENV: %v", err)
		}

		if got, want := GetNamespace(), ts.expectEnvVar; got != want {
			t.Fatalf("Invalid namespace: got: %v, want: %v", got, want)
		}
	}
}
