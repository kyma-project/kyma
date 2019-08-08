package testkit

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type TestConfig struct {
	DirectorURL string `envconfig:"DIRECTOR_URL" required:"true"`
	Tenant      string `envconfig:"TENANT" required:"true"`
	RuntimeId   string `envconfig:"RUNTIME_ID" required:"true"`

	// TODO - cleanup unused ENV

	// TODO - dafaults not working?
	Namespace                string `envconfig:"NAMESPACE" default:"compass-system"`
	IntegrationNamespace     string `envconfig:"INTEGRATION_NAMESPACE" default:"kyma-integration"`
	MockServicePort          int32  `envconfig:"MOCK_SERVICE_PORT" default:"8080"`
	MockServiceSelectorKey   string `envconfig:"SELECTOR_KEY" default:"app"`
	MockServiceSelectorValue string `envconfig:"SELECTOR_VALUE" default:"compass-runtime-agent-tests"`
	MockServiceName          string `envconfig:"MOCK_SERVICE_NAME" default:"compass-runtime-agent-tests-mock"`
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
