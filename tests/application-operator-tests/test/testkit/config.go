package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	namespaceEnvName               = "NAMESPACE"
	tillerHostEnvName              = "TILLER_HOST"
	helmTLSKeyFileEnvName          = "HELM_TLS_KEY_FILE"
	helmTLSCertificateFileEnvName  = "HELM_TLS_CERTIFICATE_FILE"
	tillerTLSSkipVerifyEnvName     = "TILLER_TLS_SKIP_VERIFY"
	gatewayOncePerNamespaceEnvName = "GATEWAY_DEPLOYED_PER_NAMESPACE"

	installationTimeoutEnvName = "INSTALLATION_TIMEOUT_SECONDS"

	defaultHelmTLSKeyFile         = "/etc/certs/tls.key"
	defaultHelmTLSCertificateFile = "/etc/certs/tls.crt"
	defaultInstallationTimeout    = 180
)

type TestConfig struct {
	Namespace                  string
	TillerHost                 string
	TillerTLSKeyFile           string
	TillerTLSCertificateFile   string
	TillerTLSSkipVerify        bool
	InstallationTimeoutSeconds int
	GatewayOncePerNamespace    bool
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

	helmTLSKeyFile, found := os.LookupEnv(helmTLSKeyFileEnvName)
	if !found {
		log.Printf("failed to read %s environment variable, using default value %s", helmTLSKeyFileEnvName, defaultHelmTLSKeyFile)
		helmTLSKeyFile = defaultHelmTLSKeyFile
	}

	helmTLSCertificateFile, found := os.LookupEnv(helmTLSCertificateFileEnvName)
	if !found {
		log.Printf("failed to read %s environment variable, using default value %s", helmTLSCertificateFileEnvName, defaultHelmTLSCertificateFile)
		helmTLSCertificateFile = defaultHelmTLSCertificateFile
	}

	tillerTLSSkipVerify := true
	sv, found := os.LookupEnv(tillerTLSSkipVerifyEnvName)
	if found {
		tillerTLSSkipVerify, _ = strconv.ParseBool(sv)
	}

	gatewayOncePerNamespace := false
	sv, found = os.LookupEnv(gatewayOncePerNamespaceEnvName)
	if found {
		gatewayOncePerNamespace, _ = strconv.ParseBool(sv)
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
		Namespace:                  namespace,
		TillerHost:                 tillerHost,
		TillerTLSKeyFile:           helmTLSKeyFile,
		TillerTLSCertificateFile:   helmTLSCertificateFile,
		TillerTLSSkipVerify:        tillerTLSSkipVerify,
		InstallationTimeoutSeconds: timeoutValue,
		GatewayOncePerNamespace:    gatewayOncePerNamespace,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
