//go:build unit
// +build unit

package env_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

func Test_GetNatsConfig(t *testing.T) {
	maxIdleConns := 10
	maxConnsPerHost := 20
	maxIdleConnsPerHost := 30
	idleConnTimeout := time.Second * 40
	jsStreamReplicas := 3

	envs := map[string]string{
		"NATS_URL":                "NATS_URL",
		"JS_STREAM_NAME":          "kyma",
		"JS_STREAM_REPLICAS":      fmt.Sprintf("%d", jsStreamReplicas),
		"EVENT_TYPE_PREFIX":       "EVENT_TYPE_PREFIX",
		"MAX_IDLE_CONNS":          fmt.Sprintf("%d", maxIdleConns),
		"MAX_CONNS_PER_HOST":      fmt.Sprintf("%d", maxConnsPerHost),
		"MAX_IDLE_CONNS_PER_HOST": fmt.Sprintf("%d", maxIdleConnsPerHost),
		"IDLE_CONN_TIMEOUT":       fmt.Sprintf("%v", idleConnTimeout),
	}

	for k, v := range envs {
		t.Setenv(k, v)
	}

	maxReconnects, reconnectWait := 1, time.Second
	config := env.GetNatsConfig(maxReconnects, reconnectWait)

	require.Equal(t, config.MaxReconnects, maxReconnects)
	require.Equal(t, config.ReconnectWait, reconnectWait)

	require.Equal(t, config.URL, envs["NATS_URL"])
	require.Equal(t, config.EventTypePrefix, envs["EVENT_TYPE_PREFIX"])

	require.Equal(t, config.MaxIdleConns, maxIdleConns)
	require.Equal(t, config.MaxConnsPerHost, maxConnsPerHost)
	require.Equal(t, config.MaxIdleConnsPerHost, maxIdleConnsPerHost)
	require.Equal(t, config.IdleConnTimeout, idleConnTimeout)

	require.Equal(t, config.JSStreamName, envs["JS_STREAM_NAME"])
	require.Equal(t, config.JSStreamReplicas, jsStreamReplicas)
}
