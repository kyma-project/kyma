package env

import (
	"os"
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
	defer func() {
		for k := range envs {
			err := os.Unsetenv(k)
			g.Expect(err).ShouldNot(HaveOccurred())
		}
	}()

	for k, v := range envs {
		err := os.Setenv(k, v)
		g.Expect(err).ShouldNot(HaveOccurred())
	}
	backendConfig := GetBackendConfig()
	// Ensure optional variables can be set
	g.Expect(backendConfig.PublisherConfig.RequestsCPU).To(Equal(envs["PUBLISHER_REQUESTS_CPU"]))
	g.Expect(backendConfig.PublisherConfig.RequestsMemory).To(Equal(envs["PUBLISHER_REQUESTS_MEMORY"]))
	g.Expect(backendConfig.PublisherConfig.RequestTimeout).To(Equal(envs["PUBLISHER_REQUEST_TIMEOUT"]))
}
