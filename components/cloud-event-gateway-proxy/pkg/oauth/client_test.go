package oauth

import (
	"context"
	"fmt"
	testingutils "github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/testing"
	"net/http"
	"testing"
	"time"

	"go.opencensus.io/plugin/ochttp"
	"golang.org/x/oauth2"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	const (
		maxIdleConns        = 100
		maxIdleConnsPerHost = 200
	)

	client := NewClient(context.Background(), &gateway.EnvConfig{MaxIdleConns: maxIdleConns, MaxIdleConnsPerHost: maxIdleConnsPerHost})
	defer client.CloseIdleConnections()

	ocTransport, ok := client.Transport.(*ochttp.Transport)
	if !ok {
		t.Errorf("Failed to convert to OpenCensus transport")
	}

	secTransport, ok := ocTransport.Base.(*oauth2.Transport)
	if !ok {
		t.Errorf("Failed to convert to oauth2 transport")
	}

	httpTransport, ok := secTransport.Base.(*http.Transport)
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
		tokenEndpoint  = "/token"
		eventsEndpoint = "/events"
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
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockServer := testingutils.NewMockServer()
			mockServer.Start(t, tokenEndpoint, eventsEndpoint, test.givenExpiresInSec)
			defer mockServer.Close()

			emsCEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint)
			authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)
			env := testingutils.NewEnvConfig(emsCEURL, authURL, 1, 1)
			client := NewClient(context.Background(), env)
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
