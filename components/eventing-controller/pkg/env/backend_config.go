package env

import (
	"github.com/kelseyhightower/envconfig"
	"log"
	"time"
)

// BackendConfig represents the environment config for the Backend Controller.
type BackendConfig struct {
	PublisherConfig PublisherConfig

	BackendCRNamespace string `envconfig:"BACKEND_CR_NAMESPACE" default:"kyma-system"`
	BackendCRName      string `envconfig:"BACKEND_CR_NAME" default:"eventing-backend"`

	DefaultSubscriptionConfig DefaultSubscriptionConfig
}

type PublisherConfig struct {
	Image           string `envconfig:"PUBLISHER_IMAGE" default:"eu.gcr.io/kyma-project/event-publisher-proxy:c06eb4fc"`
	ImagePullPolicy string `envconfig:"PUBLISHER_IMAGE_PULL_POLICY" default:"IfNotPresent"`
	PortNum         int    `envconfig:"PUBLISHER_PORT_NUM" default:"8080"`
	MetricsPortNum  int    `envconfig:"PUBLISHER_METRICS_PORT_NUM" default:"8080"`
	ServiceAccount  string `envconfig:"PUBLISHER_SERVICE_ACCOUNT" default:"eventing-publisher-proxy"`
	Replicas        int32  `envconfig:"PUBLISHER_REPLICAS" default:"1"`
	RequestsCPU     string `envconfig:"PUBLISHER_REQUESTS_CPU" default:"32m"`
	RequestsMemory  string `envconfig:"PUBLISHER_REQUESTS_MEMORY" default:"64Mi"`
	RequestTimeout  string `envconfig:"PUBLISHER_REQUEST_TIMEOUT" default:"5s"`
	LimitsCPU       string `envconfig:"PUBLISHER_LIMITS_CPU" default:"100m"`
	LimitsMemory    string `envconfig:"PUBLISHER_LIMITS_MEMORY" default:"128Mi"`
}

type DefaultSubscriptionConfig struct {
	MaxInFlightMessages int `envconfig:"DEFAULT_MAX_IN_FLIGHT_MESSAGES" default:"10"`
	DispatchRetryPeriod time.Duration `envconfig:"DEFAULT_DISPATCHER_RETRY_PERIOD" default:"5m"`
	DispatcherMaxRetries int `envconfig:"DEFAULT_DISPATCHER_MAX_RETRIES" default:"10"`
}

func GetBackendConfig() BackendConfig {
	cfg := BackendConfig{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
