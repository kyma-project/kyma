package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	namespaceEnvName           = "NAMESPACE"
	tillerHostEnvName          = "TILLER_HOST"
	installationTimeoutEnvName = "INSTALLATION_TIMEOUT"

	defaultInstallationTimeout = 180
)

type TestConfig struct {
	Namespace           string
	TillerHost          string
	ProvisioningTimeout int
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

	var timeoutValue int
	var err error
	installationTimeout, found := os.LookupEnv(installationTimeoutEnvName)
	if !found {
		timeoutValue = defaultInstallationTimeout
	} else {
		timeoutValue, err = strconv.Atoi(installationTimeout)
		if err != nil {
			return TestConfig{}, errors.New(fmt.Sprintf("failed to parse %s env value to int", installationTimeoutEnvName))
		}
	}

	config := TestConfig{
		Namespace:           namespace,
		TillerHost:          tillerHost,
		ProvisioningTimeout: timeoutValue,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
