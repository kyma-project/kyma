//go:build unit
// +build unit

package env_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

func Test_GetBackendConfig(t *testing.T) {
	g := NewGomegaWithT(t)
	envs := map[string]string{
		// optional
		"PUBLISHER_REQUESTS_CPU":    "64m",
		"PUBLISHER_REQUESTS_MEMORY": "128Mi",
		"PUBLISHER_REQUEST_TIMEOUT": "10s",
	}

	for k, v := range envs {
		t.Setenv(k, v)
	}
	backendConfig := env.GetBackendConfig()
	// Ensure optional variables can be set
	g.Expect(backendConfig.PublisherConfig.RequestsCPU).To(Equal(envs["PUBLISHER_REQUESTS_CPU"]))
	g.Expect(backendConfig.PublisherConfig.RequestsMemory).To(Equal(envs["PUBLISHER_REQUESTS_MEMORY"]))
	g.Expect(backendConfig.PublisherConfig.RequestTimeout).To(Equal(envs["PUBLISHER_REQUEST_TIMEOUT"]))
}
