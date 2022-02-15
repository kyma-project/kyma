package testing

import (
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
)

func NewEnvConfig(emsCEURL, authURL string, opts ...EnvConfigOption) *env.Config {
	envConfig := &env.Config{Port: 8080, EmsPublishURL: emsCEURL, TokenEndpoint: authURL, RequestTimeout: time.Minute}
	for _, opt := range opts {
		opt(envConfig)
	}
	return envConfig
}

type EnvConfigOption func(e *env.Config)

func WithPort(port int) EnvConfigOption {
	return func(e *env.Config) {
		e.Port = port
	}
}

func WithMaxIdleConns(maxIdleConns int) EnvConfigOption {
	return func(e *env.Config) {
		e.MaxIdleConns = maxIdleConns
	}
}

func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) EnvConfigOption {
	return func(e *env.Config) {
		e.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func WithRequestTimeout(requestTimeout time.Duration) EnvConfigOption {
	return func(e *env.Config) {
		e.RequestTimeout = requestTimeout
	}
}
