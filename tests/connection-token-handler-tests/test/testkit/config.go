package testkit

import (
	"log"
	"os"
	"strconv"
)

const centralEnvName = "CENTRAL"

type TestConfig struct {
	Central bool
}

func ReadConfig() (TestConfig, error) {
	central := false
	c, found := os.LookupEnv(centralEnvName)
	if found {
		central, _ = strconv.ParseBool(c)
	}

	config := TestConfig{
		Central: central,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
