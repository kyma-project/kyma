package testkit

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

// TODO - Envs:
// Namespace
//
const (
	isCentralEnv = "CENTRAL"
	namespaceEnv = "NAMESPACE"
)

type TestConfig struct {
	IsCentral bool
	Namespace string
}

func ReadConfig() (TestConfig, error) {

	central := false
	c, found := os.LookupEnv(isCentralEnv)
	if found {
		central, _ = strconv.ParseBool(c)
	}

	namespace, found := os.LookupEnv(namespaceEnv)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnv))
	}

	config := TestConfig{
		IsCentral: central,
		Namespace: namespace,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
