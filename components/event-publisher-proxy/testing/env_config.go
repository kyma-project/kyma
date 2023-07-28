package testing

import (
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
)

const (
	defaultPort = 8080
)

func NewEnvConfig(emsCEURL, authURL string, opts ...EnvConfigOption) *env.EventMeshConfig {
	envConfig := &env.EventMeshConfig{
		Port:                defaultPort,
		EventMeshPublishURL: emsCEURL,
		TokenEndpoint:       authURL,
		RequestTimeout:      time.Minute,
	}
	for _, opt := range opts {
		opt(envConfig)
	}
	return envConfig
}

type EnvConfigOption func(e *env.EventMeshConfig)

func WithPort(port int) EnvConfigOption {
	return func(e *env.EventMeshConfig) {
		e.Port = port
	}
}

func WithMaxIdleConns(maxIdleConns int) EnvConfigOption {
	return func(e *env.EventMeshConfig) {
		e.MaxIdleConns = maxIdleConns
	}
}

func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) EnvConfigOption {
	return func(e *env.EventMeshConfig) {
		e.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func WithRequestTimeout(requestTimeout time.Duration) EnvConfigOption {
	return func(e *env.EventMeshConfig) {
		e.RequestTimeout = requestTimeout
	}
}

func WithNamespace(namespace string) EnvConfigOption {
	return func(e *env.EventMeshConfig) {
		e.EventMeshNamespace = namespace
	}
}

func WithEventTypePrefix(eventTypePrefix string) EnvConfigOption {
	return func(e *env.EventMeshConfig) {
		e.EventTypePrefix = eventTypePrefix
	}
}
