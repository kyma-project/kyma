package sender

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

const (
	// mock server endpoints
	tokenEndpoint         = "/token"
	eventsEndpoint        = "/events"
	eventsHTTP400Endpoint = "/events400"

	// connection settings
	maxIdleConns        = 100
	maxIdleConnsPerHost = 200
)

func TestNewHttpMessageSender(t *testing.T) {
	t.Parallel()

	client := oauth.NewClient(context.Background(), &env.BebConfig{})
	defer client.CloseIdleConnections()

	msgSender := NewBebMessageSender(eventsEndpoint, client)
	if msgSender.Target != eventsEndpoint {
		t.Errorf("Message sender target is misconfigured want: %s but got: %s", eventsEndpoint, msgSender.Target)
	}
	if msgSender.Client != client {
		t.Errorf("Message sender client is misconfigured want: %#v but got: %#v", client, msgSender.Client)
	}
}

func TestNewRequestWithTarget(t *testing.T) {
	t.Parallel()

	client := oauth.NewClient(context.Background(), &env.BebConfig{MaxIdleConns: maxIdleConns, MaxIdleConnsPerHost: maxIdleConnsPerHost})
	defer client.CloseIdleConnections()

	msgSender := NewBebMessageSender(eventsEndpoint, client)

	const ctxKey, ctxValue = "testKey", "testValue"
	ctx := context.WithValue(context.Background(), ctxKey, ctxValue)
	req, err := msgSender.NewRequestWithTarget(ctx, eventsEndpoint)
	if err != nil {
		t.Errorf("Failed to create a CloudEvent HTTP request with error: %v", err)
	}
	if req == nil {
		t.Error("Failed to create a CloudEvent HTTP request want new request but got nil")
	}
	if req.Method != http.MethodPost {
		t.Errorf("HTTP request has invalid method want: %s but got: %s", http.MethodPost, req.Method)
	}
	if req.URL.Path != eventsEndpoint {
		t.Errorf("HTTP request has invalid target want: %s but got: %s", eventsEndpoint, req.URL.Path)
	}
	if len(req.Header) > 0 {
		t.Error("HTTP request should be created with empty headers")
	}
	if req.Close != false {
		t.Errorf("HTTP request close is invalid want: %v but got: %v", false, req.Close)
	}
	if req.Body != nil {
		t.Error("HTTP request should be created with empty body")
	}
	if req.Context() != ctx {
		t.Errorf("HTTP request context does not match original context want: %#v, but got %#v", ctx, req.Context())
	}
	if got := req.Context().Value(ctxKey); got != ctxValue {
		t.Errorf("HTTP request context key:value do not match mant: %v:%v but got %v:%v", ctxKey, ctxValue, ctxKey, got)
	}
}

func TestSend(t *testing.T) {
	mockServer := testingutils.NewMockServer()
	mockServer.Start(t, tokenEndpoint, eventsEndpoint, eventsHTTP400Endpoint)
	defer mockServer.Close()

	ctx := context.Background()
	emsCEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint)
	authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)
	cfg := testingutils.NewEnvConfig(emsCEURL, authURL, testingutils.WithMaxIdleConns(maxIdleConns), testingutils.WithMaxIdleConnsPerHost(maxIdleConnsPerHost))
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	msgSender := NewBebMessageSender(emsCEURL, client)

	request, err := msgSender.NewRequestWithTarget(ctx, msgSender.Target)
	if err != nil {
		t.Errorf("Failed to create a CloudEvent HTTP request with error: %v", err)
	}

	resp, err := msgSender.Send(request)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		t.Errorf("Failed to send request with error: %v", err)
	}
	if resp.StatusCode > http.StatusNoContent {
		t.Errorf("HTTP response has invalid HTTP status code want: %d but got: %d", http.StatusNoContent, resp.StatusCode)
	}
}
