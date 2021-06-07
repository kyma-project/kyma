package sender

import (
	"context"
	"net/http"
)

// BebMessageSender is responsible for sending messages over HTTP.
type BebMessageSender struct {
	Client *http.Client
	Target string
}

// NewBebMessageSender returns a new BebMessageSender instance with the given target and client.
func NewBebMessageSender(target string, client *http.Client) *BebMessageSender {
	return &BebMessageSender{Client: client, Target: target}
}

// NewRequestWithTarget returns a new HTTP POST request with the given context and target.
func (s *BebMessageSender) NewRequestWithTarget(ctx context.Context, target string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
}

// Send sends the given HTTP request and returns the HTTP response back.
func (s *BebMessageSender) Send(req *http.Request) (*http.Response, error) {
	return s.Client.Do(req)
}
