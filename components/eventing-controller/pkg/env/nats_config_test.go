package env

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_GetNatsConfig(t *testing.T) {
	maxIdleConns := 10
	maxConnsPerHost := 20
	maxIdleConnsPerHost := 30
	idleConnTimeout := time.Second * 40

	envs := map[string]string{
		"NATS_URL":                 "NATS_URL",
		"JS_STREAM_NAME":           "kyma",
		"JS_STREAM_SUBJECT_PREFIX": "kyma",
		"EVENT_TYPE_PREFIX":        "EVENT_TYPE_PREFIX",
		"MAX_IDLE_CONNS":           fmt.Sprintf("%d", maxIdleConns),
		"MAX_CONNS_PER_HOST":       fmt.Sprintf("%d", maxConnsPerHost),
		"MAX_IDLE_CONNS_PER_HOST":  fmt.Sprintf("%d", maxIdleConnsPerHost),
		"IDLE_CONN_TIMEOUT":        fmt.Sprintf("%v", idleConnTimeout),
	}

	defer func() {
		for k := range envs {
			require.NoError(t, os.Unsetenv(k))
		}
	}()

	for k, v := range envs {
		require.NoError(t, os.Setenv(k, v))
	}

	maxReconnects, reconnectWait := 1, time.Second
	config := GetNatsConfig(maxReconnects, reconnectWait)

	require.Equal(t, config.MaxReconnects, maxReconnects)
	require.Equal(t, config.ReconnectWait, reconnectWait)

	require.Equal(t, config.URL, envs["NATS_URL"])
	require.Equal(t, config.EventTypePrefix, envs["EVENT_TYPE_PREFIX"])

	require.Equal(t, config.MaxIdleConns, maxIdleConns)
	require.Equal(t, config.MaxConnsPerHost, maxConnsPerHost)
	require.Equal(t, config.MaxIdleConnsPerHost, maxIdleConnsPerHost)
	require.Equal(t, config.IdleConnTimeout, idleConnTimeout)

	require.Equal(t, config.JSStreamName, envs["JS_STREAM_NAME"])
	require.Equal(t, config.JSStreamSubjectPrefix, "EVENTTYPEPREFIX")
}

func Test_JSStreamSubjectPrefix(t *testing.T) {
	maxIdleConns := 10
	maxConnsPerHost := 20
	maxIdleConnsPerHost := 30
	idleConnTimeout := time.Second * 40
	maxReconnects, reconnectWait := 1, time.Second

	envs := map[string]string{
		"NATS_URL":                 "NATS_URL",
		"JS_STREAM_NAME":           "kyma",
		"JS_STREAM_SUBJECT_PREFIX": "kyma",
		"MAX_IDLE_CONNS":           fmt.Sprintf("%d", maxIdleConns),
		"MAX_CONNS_PER_HOST":       fmt.Sprintf("%d", maxConnsPerHost),
		"MAX_IDLE_CONNS_PER_HOST":  fmt.Sprintf("%d", maxIdleConnsPerHost),
		"IDLE_CONN_TIMEOUT":        fmt.Sprintf("%v", idleConnTimeout),
	}

	// set initial env variables
	for k, v := range envs {
		require.NoError(t, os.Setenv(k, v))
	}

	// define defer to unset env variables
	defer func() {
		for k := range envs {
			require.NoError(t, os.Unsetenv(k))
		}
	}()

	// define test cases
	testCases := []struct {
		name                      string
		givenEventTypePrefix      string
		wantJSStreamSubjectPrefix string
	}{
		{
			name:                      "When eventTypePrefix is empty",
			givenEventTypePrefix:      "",
			wantJSStreamSubjectPrefix: "kyma",
		},
		{
			name:                      "When eventTypePrefix is non-empty",
			givenEventTypePrefix:      "one.two.three",
			wantJSStreamSubjectPrefix: "one.two.three",
		},
	}
	// run test cases
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// given
			require.NoError(t, os.Setenv("EVENT_TYPE_PREFIX", tc.givenEventTypePrefix))

			// when
			config := GetNatsConfig(maxReconnects, reconnectWait)

			// then
			require.Equal(t, config.JSStreamSubjectPrefix, tc.wantJSStreamSubjectPrefix)
		})
	}
}
