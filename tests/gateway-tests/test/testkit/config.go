package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	eventServiceUrlEnvName   = "EVENT_SERVICE_URL"
	namespaceEnvName         = "NAMESPACE"
	remoteEnvironmentEnvName = "REMOTE_ENVIRONMENT"
)

type TestConfig struct {
	EventServiceUrl   string
	Namespace         string
	RemoteEnvironment string
}

func ReadConfig() (TestConfig, error) {
	eventServiceUrl, found := os.LookupEnv(eventServiceUrlEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", eventServiceUrlEnvName))
	}

	namespace, found := os.LookupEnv(namespaceEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnvName))
	}

	remoteEnvironment, found := os.LookupEnv(remoteEnvironmentEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", remoteEnvironmentEnvName))
	}

	config := TestConfig{
		EventServiceUrl:   eventServiceUrl,
		Namespace:         namespace,
		RemoteEnvironment: remoteEnvironment,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
