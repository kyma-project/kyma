package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	namespaceEnvName          = "NAMESPACE"
	tillerHostEnvName	= "TILLER_HOST"
)

type TestConfig struct {
	Namespace          string
	TillerHost	string
}

func ReadConfig() (TestConfig, error) {
	namespace, found := os.LookupEnv(namespaceEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnvName))
	}

	tillerHost, found := os.LookupEnv(tillerHostEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", tillerHostEnvName))
	}

	config := TestConfig{
		Namespace:          namespace,
		TillerHost: tillerHost,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
