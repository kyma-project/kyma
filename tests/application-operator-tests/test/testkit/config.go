package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	namespaceEnvName                               = "NAMESPACE"
	helmDriverEnvName                              = "HELM_DRIVER"
	gatewayOncePerNamespaceEnvName                 = "GATEWAY_DEPLOYED_PER_NAMESPACE"
	installationTimeoutEnvName                     = "INSTALLATION_TIMEOUT_SECONDS"
	defaultInstallationTimeout                     = 180
	centralApplicationConnectivityValidatorEnvName = "CENTRAL_APPLICATION_CONNECTIVITY_VALIDATOR"
)

type TestConfig struct {
	Namespace                               string
	HelmDriver                              string
	InstallationTimeoutSeconds              int
	GatewayOncePerNamespace                 bool
	CentralApplicationConnectivityValidator bool
}

func ReadConfig() (TestConfig, error) {
	namespace, found := os.LookupEnv(namespaceEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnvName))
	}

	helmDriver, found := os.LookupEnv(helmDriverEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", helmDriverEnvName))
	}

	gatewayOncePerNamespace := false
	sv, found := os.LookupEnv(gatewayOncePerNamespaceEnvName)
	if found {
		gatewayOncePerNamespace, _ = strconv.ParseBool(sv)
	}

	centralApplicationConnectivityValidator := true
	sv, found = os.LookupEnv(centralApplicationConnectivityValidatorEnvName)
	if found {
		centralApplicationConnectivityValidator, _ = strconv.ParseBool(sv)
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
		Namespace:                               namespace,
		HelmDriver:                              helmDriver,
		InstallationTimeoutSeconds:              timeoutValue,
		GatewayOncePerNamespace:                 gatewayOncePerNamespace,
		CentralApplicationConnectivityValidator: centralApplicationConnectivityValidator,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
