package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	namespaceEnvName          = "NAMESPACE"
)

type TestConfig struct {
	MetadataServiceUrl string
	Namespace          string
}

func ReadConfig() (TestConfig, error) {
	namespace, found := os.LookupEnv(namespaceEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnvName))
	}

	config := TestConfig{
		Namespace:          namespace,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
