/*
 *  © 2018 SAP SE or an SAP affiliate company.
 *  All rights reserved.
 *  Please see http://www.sap.com/corporate-en/legal/copyright/index.epx for additional trademark information and
 *  notices.
 */
package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	metadataServiceUrlEnvName = "METADATA_URL"
	namespaceEnvName          = "NAMESPACE"
)

type TestConfig struct {
	MetadataServiceUrl string
	Namespace          string
}

func ReadConfig() (TestConfig, error) {
	metadataUrl, found := os.LookupEnv(metadataServiceUrlEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", metadataServiceUrlEnvName))
	}

	namespace, found := os.LookupEnv(namespaceEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", namespaceEnvName))
	}

	config := TestConfig{
		MetadataServiceUrl: metadataUrl,
		Namespace:          namespace,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
