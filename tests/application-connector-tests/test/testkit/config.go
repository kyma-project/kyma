package testkit

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const (
	skipSSLVerifyEnv = "SKIP_SSL_VERIFY"
	namespaceEnv     = "NAMESPACE"
)

type TestConfig struct {
	SkipSSLVerify bool
	Namespace     string
}

func ReadConfig() (TestConfig, error) {
	skipVerify := false
	sv, found := os.LookupEnv(skipSSLVerifyEnv)
	if found {
		skipVerify, _ = strconv.ParseBool(sv)
	}

	namespace, found := os.LookupEnv(namespaceEnv)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnv))
	}

	config := TestConfig{
		SkipSSLVerify: skipVerify,
		Namespace:     namespace,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
