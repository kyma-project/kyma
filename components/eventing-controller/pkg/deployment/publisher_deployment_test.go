package deployment

import (
	"os"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
)

func Test_GetBEBEnvVars(t *testing.T) {
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
	backendConfig := env.GetBackendConfig()
	envVars := getBEBEnvVars(backendConfig.PublisherConfig)

	// Ensure request timeout is set as Env variable
	requestTimeoutEnv := findEnvVar(envVars, "REQUEST_TIMEOUT")
	g.Expect(requestTimeoutEnv).ShouldNot(BeNil())
	g.Expect(requestTimeoutEnv.Value).Should(Equal(envs["PUBLISHER_REQUEST_TIMEOUT"]))
}

func Test_GetNATSEnvVars(t *testing.T) {
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
	backendConfig := env.GetBackendConfig()
	envVars := getNATSEnvVars(backendConfig.PublisherConfig)

	// Ensure request timeout is set as Env variable
	requestTimeoutEnv := findEnvVar(envVars, "REQUEST_TIMEOUT")
	g.Expect(requestTimeoutEnv).ShouldNot(BeNil())
	g.Expect(requestTimeoutEnv.Value).Should(Equal(envs["PUBLISHER_REQUEST_TIMEOUT"]))
}

// findEnvVar returns the env variable which has name == envVar.Name,
// or nil if there is no such env variable.
func findEnvVar(envVars []v1.EnvVar, name string) *v1.EnvVar {
	for _, n := range envVars {
		if name == n.Name {
			return &n
		}
	}
	return nil
}
