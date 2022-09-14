//go:build integration
// +build integration

package oauth_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"go.opencensus.io/plugin/ochttp"
	"golang.org/x/oauth2"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"

	sut "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	// given
	const (
		maxIdleConns        = 100
		maxIdleConnsPerHost = 200
	)

	client := NewClient(context.Background(), &env.BEBConfig{MaxIdleConns: maxIdleConns, MaxIdleConnsPerHost: maxIdleConnsPerHost})
	defer client.CloseIdleConnections()

	// when
	ocTransport, ok := client.Transport.(*ochttp.Transport)
	// then
	if !ok {
		t.Errorf("Failed to convert to OpenCensus transport")
	}

	// when
	secTransport, ok := ocTransport.Base.(*oauth2.Transport)
	// then
	if !ok {
		t.Errorf("Failed to convert to oauth2 transport")
	}

	// when
	httpTransport, ok := secTransport.Base.(*http.Transport)
	// then
	if !ok {
		t.Errorf("Failed to convert to HTTP transport")
	}
	if httpTransport.MaxIdleConns != maxIdleConns {
		t.Errorf("HTTP Client Transport MaxIdleConns is misconfigured want: %d but got: %d", maxIdleConns, httpTransport.MaxIdleConns)
	}
	if httpTransport.MaxIdleConnsPerHost != maxIdleConnsPerHost {
		t.Errorf("HTTP Client Transport MaxIdleConnsPerHost is misconfigured want: %d but got: %d", maxIdleConnsPerHost, httpTransport.MaxIdleConnsPerHost)
	}
}

func TestGetToken(t *testing.T) {
	t.Parallel()

	const (
		tokenEndpoint         = "/token"
		eventsEndpoint        = "/events"
		eventsHTTP400Endpoint = "/events400"
	)

	testCases := []struct {
		name                     string
		delay                    time.Duration
		requestsCount            int
		givenExpiresInSec        int
		wantGeneratedTokensCount int
	}{
		{
			name:                     "Token expires every 60 seconds",
			delay:                    time.Millisecond,
			requestsCount:            50,
			givenExpiresInSec:        60,
			wantGeneratedTokensCount: 1,
		},
		{
			name:                     "Token expires every second",
			delay:                    time.Second + time.Millisecond,
			requestsCount:            5,
			givenExpiresInSec:        1,
			wantGeneratedTokensCount: 5,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockServer := testingutils.NewMockServer(testingutils.WithExpiresIn(test.givenExpiresInSec))
			mockServer.Start(t, tokenEndpoint, eventsEndpoint, eventsHTTP400Endpoint)
			defer mockServer.Close()

			emsCEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint)
			authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)
			cfg := testingutils.NewEnvConfig(emsCEURL, authURL)
			client := sut.NewClient(context.Background(), cfg)
			defer client.CloseIdleConnections()

			for i := 0; i < test.requestsCount; i++ {
				req, err := http.NewRequest(http.MethodPost, emsCEURL, nil)
				if err != nil {
					t.Errorf("Failed to create HTTP request with error: %v", err)
				}

				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("Failed to post HTTP request with error: %v", err)
				}
				_ = resp.Body.Close()

				time.Sleep(test.delay)
			}

			if got := mockServer.GeneratedTokensCount(); got != test.wantGeneratedTokensCount {
				t.Fatalf("Tokens count does not match, want: %d but got: %d", test.wantGeneratedTokensCount, got)
			}
		})
	}
}
