package gateway

import (
	"net/http"
	"testing"
)

func TestConfigureTransport(t *testing.T) {
	t.Parallel()

	const (
		maxIdleConns        = 100
		maxIdleConnsPerHost = 200
	)

	transport := &http.Transport{}
	env := EnvConfig{MaxIdleConns: maxIdleConns, MaxIdleConnsPerHost: maxIdleConnsPerHost}
	env.ConfigureTransport(transport)

	if transport.MaxIdleConns != maxIdleConns {
		t.Errorf("HTTP Transport MaxIdleConns is misconfigured want: %d but got: %d", maxIdleConns, transport.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != maxIdleConnsPerHost {
		t.Errorf("HTTP Transport MaxIdleConnsPerHost is misconfigured want: %d but got: %d", maxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
	}
}
