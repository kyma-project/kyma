package testing

import (
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
)

func NewEnvConfig(emsCEURL, authURL string, opts ...EnvConfigOption) *env.BEBConfig {
	envConfig := &env.BEBConfig{Port: 8080, EmsPublishURL: emsCEURL, TokenEndpoint: authURL, RequestTimeout: time.Minute}
	for _, opt := range opts {
		opt(envConfig)
	}
	return envConfig
}

type EnvConfigOption func(e *env.BEBConfig)

func WithPort(port int) EnvConfigOption {
	return func(e *env.BEBConfig) {
		e.Port = port
	}
}

func WithMaxIdleConns(maxIdleConns int) EnvConfigOption {
	return func(e *env.BEBConfig) {
		e.MaxIdleConns = maxIdleConns
	}
}

func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) EnvConfigOption {
	return func(e *env.BEBConfig) {
		e.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func WithRequestTimeout(requestTimeout time.Duration) EnvConfigOption {
	return func(e *env.BEBConfig) {
		e.RequestTimeout = requestTimeout
	}
}

func WithBEBNamespace(bebNs string) EnvConfigOption {
	return func(e *env.BEBConfig) {
		e.BEBNamespace = bebNs
	}
}

func WithEventTypePrefix(eventTypePrefix string) EnvConfigOption {
	return func(e *env.BEBConfig) {
		e.EventTypePrefix = eventTypePrefix
	}
}
