package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	internalAPIUrlEnvName = "INTERNAL_API_URL"
	externalAPIUrlEnvName = "EXTERNAL_API_URL"
	gatewayUrlEnvName     = "GATEWAY_URL"
)

type TestConfig struct {
	InternalAPIUrl string
	ExternalAPIUrl string
	GatewayUrl     string
}

func ReadConfig() (TestConfig, error) {
	internalAPIUrl, found := os.LookupEnv(internalAPIUrlEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", internalAPIUrlEnvName))
	}

	externalAPIUrl, found := os.LookupEnv(externalAPIUrlEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", externalAPIUrlEnvName))
	}

	gatewayUrl, found := os.LookupEnv(gatewayUrlEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", gatewayUrlEnvName))
	}

	config := TestConfig{
		InternalAPIUrl: internalAPIUrl,
		ExternalAPIUrl: externalAPIUrl,
		GatewayUrl:     gatewayUrl,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
