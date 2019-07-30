package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	directorURLEnvName        = "DIRECTOR_URL"
	runtimeIdConfigMapEnvName = "RUNTIME_ID_CONFIG_MAP"
	tenantEnvName             = "TENANT"
	namespaceEnvName          = "NAMESPACE"
	mockSelectorKeyEnvName    = "SELECTOR_KEY"
	mockSelectorValueEnvName  = "SELECTOR_VALUE"
	mockServicePortEnvName    = "MOCK_SERVICE_PORT"
)

type TestConfig struct {
	DirectorURL              string
	RuntimeIdConfigMap       string
	Tenant                   string
	Namespace                string
	MockServicePort          int32
	MockServiceSelectorKey   string
	MockServiceSelectorValue string
}

func ReadConfig() (TestConfig, error) {
	directorUrl, found := os.LookupEnv(directorURLEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", directorURLEnvName))
	}

	runtimeIdConfigMap, found := os.LookupEnv(runtimeIdConfigMapEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", runtimeIdConfigMapEnvName))
	}

	tenant, found := os.LookupEnv(tenantEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", tenantEnvName))
	}

	namespace, found := os.LookupEnv(namespaceEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnvName))
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
		DirectorURL:              directorUrl,
		RuntimeIdConfigMap:       runtimeIdConfigMap,
		Tenant:                   tenant,
		Namespace:                namespace,
		MockServicePort:          int32(mockServicePort),
		MockServiceSelectorKey:   selectorKey,
		MockServiceSelectorValue: selectorValue,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
