package testkit

import (
	"log"
	"time"

	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	Tenant    string
	RuntimeId string

	Runtime struct {
		EventsURL  string `envconfig:"default=https://gateway.kyma.local"`
		ConsoleURL string `envconfig:"default=https://console.kyma.local"`
	}

	DirectorURL                    string        `envconfig:"default=https://compass-gateway-auth-oauth.kyma.local"`
	Namespace                      string        `envconfig:"default=compass-system"`
	IntegrationNamespace           string        `envconfig:"default=kyma-integration"`
	TestPodAppLabel                string        `envconfig:"default=compass-runtime-agent-tests"`
	MockServicePort                int32         `envconfig:"default=8080"`
	MockServiceName                string        `envconfig:"default=compass-runtime-agent-tests-mock"`
	ConfigApplicationWaitTime      time.Duration `envconfig:"default=40s"`
	ProxyInvalidationWaitTime      time.Duration `envconfig:"default=150s"`
	GraphQLLog                     bool          `envconfig:"default=false"`
	ScenarioLabel                  string        `envconfig:"default=COMPASS_RUNTIME_AGENT_TESTS"`
	HydraPublicURL                 string        `envconfig:"default=http://ory-hydra-public.kyma-system:4444"`
	HydraAdminURL                  string        `envconfig:"default=http://ory-hydra-admin.kyma-system:4445"`
	ApplicationInstallationTimeout time.Duration `envconfig:"default=180s"`
}

func ReadConfig() (TestConfig, error) {
	var config TestConfig
	err := envconfig.Init(&config)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
