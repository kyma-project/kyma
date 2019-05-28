package testkit

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const (
	isCentralEnv     = "CENTRAL"
	skipSSLVerifyEnv = "SKIP_SSL_VERIFY"
	namespaceEnv     = "NAMESPACE"
)

type TestConfig struct {
	IsCentral     bool
	SkipSSLVerify bool
	Namespace     string
}

func ReadConfig() (TestConfig, error) {

	central := false
	c, found := os.LookupEnv(isCentralEnv)
	if found {
		central, _ = strconv.ParseBool(c)
	}

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
		IsCentral:     central,
		SkipSSLVerify: skipVerify,
		Namespace:     namespace,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
