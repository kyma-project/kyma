package executor

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	applicationEnvName       = "APPLICATION"
	namespaceEnvName         = "NAMESPACE"
	mockSelectorKeyEnvName   = "SELECTOR_KEY"
	mockSelectorValueEnvName = "SELECTOR_VALUE"
	mockServicePortEnvName   = "MOCK_SERVICE_PORT"
)

type TestConfig struct {
	Application       string
	Namespace         string
	MockSelectorKey   string
	MockSelectorValue string
	MockServicePort   int32
}

func ReadConfig() (TestConfig, error) {
	namespace, found := os.LookupEnv(namespaceEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnvName))
	}

	application, found := os.LookupEnv(applicationEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", applicationEnvName))
	}

	selectorKey, found := os.LookupEnv(mockSelectorKeyEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", mockSelectorKeyEnvName))
	}

	selectorValue, found := os.LookupEnv(mockSelectorValueEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", mockSelectorValueEnvName))
	}

	mockServicePortStr, found := os.LookupEnv(mockServicePortEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", mockServicePortEnvName))
	}

	mockServicePort, err := strconv.Atoi(mockServicePortStr)
	if err != nil {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to parse %s env value to int", mockServicePortEnvName))
	}

	config := TestConfig{
		Application:       application,
		Namespace:         namespace,
		MockSelectorKey:   selectorKey,
		MockSelectorValue: selectorValue,
		MockServicePort:   int32(mockServicePort),
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
