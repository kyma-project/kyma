package testkit

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type TestConfig struct {
	DirectorURL string `envconfig:"DIRECTOR_URL" required:"true"`
	Tenant      string `envconfig:"TENANT" required:"true"`
	RuntimeId   string `envconfig:"RUNTIME_ID" required:"true"`

	Namespace                 string `envconfig:"NAMESPACE" default:"compass-system"`
	IntegrationNamespace      string `envconfig:"INTEGRATION_NAMESPACE" default:"kyma-integration"`
	TestPodAppLabel           string `envconfig:"TEST_POD_APP_LABEL" default:"compass-runtime-agent-tests"`
	MockServicePort           int32  `envconfig:"MOCK_SERVICE_PORT" default:"8080"`
	MockServiceName           string `envconfig:"MOCK_SERVICE_NAME" default:"compass-runtime-agent-tests-mock"`
	ConfigApplicationWaitTime int64  `envconfig:"CONFIG_APPLICATION_WAIT_TIME" default:"40"`
	ProxyInvalidationWaitTime int64  `envconfig:"PROXY_INVALIDATION_WAIT_TIME" default:"150"`
	GraphQLLog                bool   `envconfig:"GRAPHQL_LOG" default:"false"`
	ScenarioLabel             string `envconfig:"SCENARIO_LABEL" default:"COMPASS_RUNTIME_AGENT_TESTS"`
	HyadraURL                 string `envconfig:"HYDRA_URL" default:"https://oauth2.kyma.local"`
}

func ReadConfig() (TestConfig, error) {
	var config TestConfig
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
