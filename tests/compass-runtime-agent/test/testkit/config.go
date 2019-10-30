package testkit

import (
	"log"
	"time"

	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	Tenant    string
	RuntimeId string

	// Using internal Director URL and internal JWT is a temporary solution. See https://github.com/kyma-project/kyma/issues/6106
	InternalDirectorJWT string

	Runtime struct {
		EventsURL  string `envconfig:"default=https://gateway.kyma.local"`
		ConsoleURL string `envconfig:"default=https://console.kyma.local"`
	}

	DirectorURL                    string        `envconfig:"default=https://compass-director.compass-system.svc.cluster.local:3000"`
	Namespace                      string        `envconfig:"default=compass-system"`
	IntegrationNamespace           string        `envconfig:"default=kyma-integration"`
	TestPodAppLabel                string        `envconfig:"default=compass-runtime-agent-tests"`
	MockServicePort                int32         `envconfig:"default=8080"`
	MockServiceName                string        `envconfig:"default=compass-runtime-agent-tests-mock"`
	ConfigApplicationWaitTime      time.Duration `envconfig:"default=40s"`
	ProxyInvalidationWaitTime      time.Duration `envconfig:"default=150s"`
	GraphQLLog                     bool          `envconfig:"default=false"`
	ScenarioLabel                  string        `envconfig:"default=COMPASS_RUNTIME_AGENT_TESTS"`
	ApplicationInstallationTimeout time.Duration `envconfig:"default=180s"`

	IdProvider struct {
		Domain        string        `envconfig:"default=kyma.local"`
		UserEmail     string        `envconfig:"default=admin@kyma.cx"`
		UserPassword  string        `envconfig:"default=nimda"` //TODO: Check the actual default credentials
		ClientTimeout time.Duration `envconfig:"default=10s"`
	}
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
