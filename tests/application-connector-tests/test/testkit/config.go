package testkit

import (
	"log"
	"os"
	"strconv"
)

// TODO - Envs:
// Namespace
//
const (
	isCentralEnv = "CENTRAL"
)

type TestConfig struct {
	ConnectorInternalAPIURL string
	IsCentral               bool
}

func ReadConfig() (TestConfig, error) {

	central := false
	c, found := os.LookupEnv(isCentralEnv)
	if found {
		central, _ = strconv.ParseBool(c)
	}

	config := TestConfig{
		IsCentral: central,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
