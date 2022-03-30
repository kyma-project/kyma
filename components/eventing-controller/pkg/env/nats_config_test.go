package env

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/onsi/gomega"
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

	g := NewGomegaWithT(t)
	defer func() {
		for k := range envs {
			err := os.Unsetenv(k)
			g.Expect(err).ShouldNot(HaveOccurred())
		}
	}()

	for k, v := range envs {
		err := os.Setenv(k, v)
		g.Expect(err).ShouldNot(HaveOccurred())
	}

	maxReconnects, reconnectWait := 1, time.Second
	config := GetNatsConfig(maxReconnects, reconnectWait)

	g.Expect(config.MaxReconnects).To(Equal(maxReconnects))
	g.Expect(config.ReconnectWait).To(Equal(reconnectWait))

	g.Expect(config.URL).To(Equal(envs["NATS_URL"]))
	g.Expect(config.EventTypePrefix).To(Equal(envs["EVENT_TYPE_PREFIX"]))

	g.Expect(config.MaxIdleConns).To(Equal(maxIdleConns))
	g.Expect(config.MaxConnsPerHost).To(Equal(maxConnsPerHost))
	g.Expect(config.MaxIdleConnsPerHost).To(Equal(maxIdleConnsPerHost))
	g.Expect(config.IdleConnTimeout).To(Equal(idleConnTimeout))
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
