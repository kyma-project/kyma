package env

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// BackendConfig represents the environment config for the Backend Controller.
type BackendConfig struct {
	PublisherImage          string `envconfig:"PUBLISHER_IMAGE" default:"eu.gcr.io/kyma-project/event-publisher-proxy:c06eb4fc"`
	PublisherPortNum        int    `envconfig:"PUBLISHER_PORT_NUM" default:"8080"`
	PublisherMetricsPortNum int    `envconfig:"PUBLISHER_METRICS_PORT_NUM" default:"8080"`
	PublisherServiceAccount string `envconfig:"PUBLISHER_SERVICE_ACCOUNT" default:"eventing-publisher-proxy"`
	PublisherReplicas       int32  `envconfig:"PUBLISHER_REPLICAS" default:"1"`

	BackendCRNamespace string `envconfig:"BACKEND_CR_NAMESPACE" default:"kyma-system"`
	BackendCRName      string `envconfig:"BACKEND_CR_NAME" default:"eventing-backend"`
}

func GetBackendConfig() BackendConfig {
	cfg := BackendConfig{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
