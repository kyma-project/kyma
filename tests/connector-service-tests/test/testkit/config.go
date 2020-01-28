package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	internalAPIUrlEnvName = "INTERNAL_API_URL"
	externalAPIUrlEnvName = "EXTERNAL_API_URL"
	gatewayUrlEnvName     = "GATEWAY_URL"
	skipVerifyEnvName     = "SKIP_SSL_VERIFY"
	centralEnvName        = "CENTRAL"
	compassEnvName        = "COMPASS"
	eventsBaseURLEnvName  = "EVENTS_BASE_URL"
)

type TestConfig struct {
	InternalAPIUrl string
	ExternalAPIUrl string
	GatewayUrl     string
	EventBaseURL   string
	SkipSslVerify  bool
	Central        bool
	Compass        bool
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

	skipVerify := false
	sv, found := os.LookupEnv(skipVerifyEnvName)
	if found {
		skipVerify, _ = strconv.ParseBool(sv)
	}

	central := getBoolEnv(centralEnvName)
	compass := getBoolEnv(compassEnvName)

	eventsBaseURL := gatewayUrl
	ebu, found := os.LookupEnv(eventsBaseURLEnvName)
	if found {
		eventsBaseURL = ebu
	}

	config := TestConfig{
		InternalAPIUrl: internalAPIUrl,
		ExternalAPIUrl: externalAPIUrl,
		GatewayUrl:     gatewayUrl,
		EventBaseURL:   eventsBaseURL,
		SkipSslVerify:  skipVerify,
		Central:        central,
		Compass:        compass,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}

func getBoolEnv(envName string) bool {
	value := false
	v, found := os.LookupEnv(envName)
	if found {
		value, _ = strconv.ParseBool(v)
	}

	return value
}
