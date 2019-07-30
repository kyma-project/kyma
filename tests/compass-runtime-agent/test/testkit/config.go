package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	directorURLEnvName        = "DIRECTOR_URL"
	runtimeIdConfigMapEnvName = "RUNTIME_ID_CONFIG_MAP"
	tenantEnvName             = "TENANT"
)

type TestConfig struct {
	DirectorURL        string
	RuntimeIdConfigMap string
	Tenant             string
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

	config := TestConfig{
		DirectorURL:        directorUrl,
		RuntimeIdConfigMap: runtimeIdConfigMap,
		Tenant:             tenant,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
