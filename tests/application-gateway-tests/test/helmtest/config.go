package helmtest

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
)

type TestConfig struct {
	Application        string
	Namespace          string
	ServiceAccountName string
	MockServicePort    int
	TestExecutorImage  string
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

	serviceAccountName, found := os.LookupEnv(serviceAccountEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", serviceAccountEnvName))
	}

	mockServicePortStr, found := os.LookupEnv(mockServicePortEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", mockServicePortEnvName))
	}

	mockServicePort, err := strconv.Atoi(mockServicePortStr)
	if err != nil {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to parse %s env value to int", mockServicePortEnvName))
	}

	testExecutorImage, found := os.LookupEnv(testExecutorImageEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to parse %s env value to int", testExecutorImageEnvName))
	}

	config := TestConfig{
		Application:        application,
		Namespace:          namespace,
		ServiceAccountName: serviceAccountName,
		MockServicePort:    mockServicePort,
		TestExecutorImage:  testExecutorImage,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
