package testing

import (
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
)

func NewEnvConfig(emsCEURL, authURL string, opts ...EnvConfigOption) *env.BebConfig {
	envConfig := &env.BebConfig{Port: 8080, EmsPublishURL: emsCEURL, TokenEndpoint: authURL, RequestTimeout: time.Minute}
	for _, opt := range opts {
		opt(envConfig)
	}
	return envConfig
}

type EnvConfigOption func(e *env.BebConfig)

func WithPort(port int) EnvConfigOption {
	return func(e *env.BebConfig) {
		e.Port = port
	}
}

func WithMaxIdleConns(maxIdleConns int) EnvConfigOption {
	return func(e *env.BebConfig) {
		e.MaxIdleConns = maxIdleConns
	}
}

func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) EnvConfigOption {
	return func(e *env.BebConfig) {
		e.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func WithRequestTimeout(requestTimeout time.Duration) EnvConfigOption {
	return func(e *env.BebConfig) {
		e.RequestTimeout = requestTimeout
	}
}

func WithBEBNamespace(bebNs string) EnvConfigOption {
	return func(e *env.BebConfig) {
		e.BEBNamespace = bebNs
	}
}

func WithEventTypePrefix(eventTypePrefix string) EnvConfigOption {
	return func(e *env.BebConfig) {
		e.EventTypePrefix = eventTypePrefix
	}
}
