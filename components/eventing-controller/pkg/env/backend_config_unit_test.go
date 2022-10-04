package env

import (
	"testing"

	. "github.com/onsi/gomega"
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
	backendConfig := GetBackendConfig()
	// Ensure optional variables can be set
	g.Expect(backendConfig.PublisherConfig.RequestsCPU).To(Equal(envs["PUBLISHER_REQUESTS_CPU"]))
	g.Expect(backendConfig.PublisherConfig.RequestsMemory).To(Equal(envs["PUBLISHER_REQUESTS_MEMORY"]))
	g.Expect(backendConfig.PublisherConfig.RequestTimeout).To(Equal(envs["PUBLISHER_REQUEST_TIMEOUT"]))
}
