package gateway

import "time"

type Config struct {
	StatusRefreshPeriod  time.Duration `envconfig:"default=15s"`
	StatusCallTimeout    time.Duration `envconfig:"default=500ms"`
	IntegrationNamespace string        `envconfig:"default=kyma-integration"`
}
