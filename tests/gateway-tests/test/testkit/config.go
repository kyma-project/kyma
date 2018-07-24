package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	gatewayUrlEnvName        = "GATEWAY_URL"
	namespaceEnvName         = "NAMESPACE"
	remoteEnvironmentEnvName = "REMOTE_ENVIRONMENT"
)

type TestConfig struct {
	GatewayUrl        string
	Namespace         string
	RemoteEnvironment string
}

func ReadConfig() (TestConfig, error) {
	gatewayUrl, found := os.LookupEnv(gatewayUrlEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", gatewayUrlEnvName))
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
		GatewayUrl:        gatewayUrl,
		Namespace:         namespace,
		RemoteEnvironment: remoteEnvironment,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
