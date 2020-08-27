package sender

import (
	"context"
	"go.opencensus.io/plugin/ochttp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHttpMessageSender(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{}
	connectionArgs := &ConnectionArgs{MaxIdleConns: maxIdleConns, MaxIdleConnsPerHost: maxIdleConnsPerHost}
	msgSender, err := NewHttpMessageSender(connectionArgs, target, httpClient)

	if err != nil {
		t.Errorf("Failed to create a new message sender with error: %v", err)
	}
	if msgSender.Target != target {
		t.Errorf("Message sender target misconfigured want: %s but got: %s", target, msgSender.Target)
	}

	ocTransport, ok := httpClient.Transport.(*ochttp.Transport)
	if !ok {
		t.Errorf("Failed to convert HTTP transport")
	}

	httpTransport, ok := ocTransport.Base.(*http.Transport)
	if !ok {
		t.Errorf("Failed to convert OpenCensus transport")
	}

	if httpTransport.MaxIdleConns != maxIdleConns {
		t.Errorf("HTTP Client Transport MaxIdleConns is misconfigured want: %d but got: %d", maxIdleConns, httpTransport.MaxIdleConns)
	}
	if httpTransport.MaxIdleConnsPerHost != maxIdleConnsPerHost {
		t.Errorf("HTTP Client Transport MaxIdleConnsPerHost is misconfigured want: %d but got: %d", maxIdleConnsPerHost, httpTransport.MaxIdleConnsPerHost)
	}
}

func TestNewCloudEventRequestWithTarget(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{}
	connectionArgs := &ConnectionArgs{MaxIdleConns: maxIdleConns, MaxIdleConnsPerHost: maxIdleConnsPerHost}

	msgSender, err := NewHttpMessageSender(connectionArgs, target, httpClient)
	if err != nil {
		t.Errorf("Failed to create a new message sender with error: %v", err)
	}

	const ctxKey, ctxValue = "testKey", "testValue"
	ctx := context.WithValue(context.Background(), ctxKey, ctxValue)
	req, err := msgSender.NewCloudEventRequestWithTarget(ctx, target)
	if err != nil {
		t.Errorf("Failed to create a CloudEvent HTTP request with error: %v", err)
	}
	if req == nil {
		t.Error("Failed to create a CloudEvent HTTP request want new request but got nil")
	}
	if req.Method != http.MethodPost {
		t.Errorf("HTTP request has invalid method want: %s but got: %s", http.MethodPost, req.Method)
	}
	if req.URL.Path != target {
		t.Errorf("HTTP request has invalid target want: %s but got: %s", target, req.URL.Path)
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
	mockServer := startMockServer()
	defer mockServer.Close()

	httpClient := &http.Client{}
	connectionArgs := &ConnectionArgs{MaxIdleConns: maxIdleConns, MaxIdleConnsPerHost: maxIdleConnsPerHost}

	msgSender, err := NewHttpMessageSender(connectionArgs, mockServer.URL, httpClient)
	if err != nil {
		t.Errorf("Failed to create a new message sender with error: %v", err)
	}

	ctx := context.Background()
	request, err := msgSender.NewCloudEventRequestWithTarget(ctx, msgSender.Target)
	if err != nil {
		t.Errorf("Failed to create a CloudEvent HTTP request with error: %v", err)
	}

	resp, err := msgSender.Send(request)
	if err != nil {
		t.Errorf("Failed to send request with error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("HTTP response has invalid HTTP status code want: %d but got: %d", http.StatusOK, resp.StatusCode)
	}
}

func startMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

const (
	target              = "/target"
	maxIdleConns        = 100000000
	maxIdleConnsPerHost = 200000000
)
