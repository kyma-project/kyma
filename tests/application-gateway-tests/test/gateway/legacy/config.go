package legacy

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
	serviceAccountEnvName    = "SERVICE_ACCOUNT"
	mockServicePortEnvName   = "MOCK_SERVICE_PORT"
	testExecutorImageEnvName = "TEST_EXECUTOR_IMAGE"
	GatewayLegacyModeEnvName = "GatewayLegacyModeEnvName"
	mockSelectorKeyEnvName   = "SELECTOR_KEY"
	mockSelectorValueEnvName = "SELECTOR_VALUE"
)

type LegacyModeTestConfig struct {
	Application        string
	Namespace          string
	ServiceAccountName string
	MockServicePort    int32
	MockSelectorKey    string
	MockSelectorValue  string
	TestExecutorImage  string
	GatewayLegacyMode  bool
}

func ReadConfig() (LegacyModeTestConfig, error) {
	namespace, found := os.LookupEnv(namespaceEnvName)
	if !found {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnvName))
	}

	application, found := os.LookupEnv(applicationEnvName)
	if !found {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", applicationEnvName))
	}

	serviceAccountName, found := os.LookupEnv(serviceAccountEnvName)
	if !found {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", serviceAccountEnvName))
	}

	mockServicePortStr, found := os.LookupEnv(mockServicePortEnvName)
	if !found {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", mockServicePortEnvName))
	}

	mockSelectorKey, found := os.LookupEnv(mockSelectorKeyEnvName)
	if !found {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", mockSelectorKeyEnvName))
	}

	mockSelectorValue, found := os.LookupEnv(mockSelectorValueEnvName)
	if !found {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", mockSelectorValueEnvName))
	}

	mockServicePort, err := strconv.Atoi(mockServicePortStr)
	if err != nil {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to parse %s env value to int", mockServicePortEnvName))
	}

	testExecutorImage, found := os.LookupEnv(testExecutorImageEnvName)
	if !found {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to parse %s env value to int", testExecutorImageEnvName))
	}

	gatewayLegacyModeStr, found := os.LookupEnv(GatewayLegacyModeEnvName)
	if !found {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", GatewayLegacyModeEnvName))
	}

	legacyMode, err := strconv.ParseBool(gatewayLegacyModeStr)

	if err != nil {
		return LegacyModeTestConfig{}, errors.New(fmt.Sprintf("failed to parse %s env value to bool", GatewayLegacyModeEnvName))
	}

	config := LegacyModeTestConfig{
		Application:        application,
		Namespace:          namespace,
		ServiceAccountName: serviceAccountName,
		MockServicePort:    int32(mockServicePort),
		MockSelectorKey:    mockSelectorKey,
		MockSelectorValue:  mockSelectorValue,
		TestExecutorImage:  testExecutorImage,
		GatewayLegacyMode:  legacyMode,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
