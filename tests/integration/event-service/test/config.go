package test

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	eventServiceUrlEnvName = "EVENT_SERVICE_URL"
	namespaceEnvName       = "NAMESPACE"
	applicationEnvName     = "APPLICATION"
)

type TestConfig struct {
	EventServiceUrl string
	Namespace       string
	Application     string
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

	application, found := os.LookupEnv(applicationEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", applicationEnvName))
	}

	config := TestConfig{
		EventServiceUrl: eventServiceUrl,
		Namespace:       namespace,
		Application:     application,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
