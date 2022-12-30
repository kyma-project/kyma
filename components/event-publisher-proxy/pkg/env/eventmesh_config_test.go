package env

import (
	"net/http"
	"testing"
)

func TestConfigureTransport(t *testing.T) {
	t.Parallel()

	const (
		maxIdleConnections        = 100
		maxIdleConnectionsPerHost = 200
	)

	transport := &http.Transport{}
	cfg := BEBConfig{MaxIdleConns: maxIdleConnections, MaxIdleConnsPerHost: maxIdleConnectionsPerHost}
	cfg.ConfigureTransport(transport)

	if transport.MaxIdleConns != maxIdleConnections {
		t.Errorf("HTTP Transport MaxIdleConns is misconfigured want: %d but got: %d", maxIdleConnections, transport.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != maxIdleConnectionsPerHost {
		t.Errorf("HTTP Transport MaxIdleConnsPerHost is misconfigured want: %d but got: %d", maxIdleConnectionsPerHost, transport.MaxIdleConnsPerHost)
	}
}
