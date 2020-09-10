package sender

import (
	"context"
	"net/http"
)

// HttpMessageSender is responsible for sending messages over HTTP.
type HttpMessageSender struct {
	Client *http.Client
	Target string
}

// NewHttpMessageSender returns a new HttpMessageSender instance with the given target and client.
func NewHttpMessageSender(target string, client *http.Client) *HttpMessageSender {
	return &HttpMessageSender{Client: client, Target: target}
}

// NewRequestWithTarget returns a new HTTP POST request with the given context and target.
func (s *HttpMessageSender) NewRequestWithTarget(ctx context.Context, target string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
}

// Send sends the given HTTP request and returns the HTTP response back.
func (s *HttpMessageSender) Send(req *http.Request) (*http.Response, error) {
	return s.Client.Do(req)
}
