package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// TODO - Envs:
// Connector Internal API url
//
const (
	connectorInternalAPIURLEnv = "CONNECTOR_INTERNAL_API_URL"
)

type TestConfig struct {
	ConnectorInternalAPIURL string
}

func ReadConfig() (TestConfig, error) {

	connectorInternalURL, found := os.LookupEnv(connectorInternalAPIURLEnv)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", connectorInternalAPIURLEnv))
	}

	config := TestConfig{
		ConnectorInternalAPIURL: connectorInternalURL,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
