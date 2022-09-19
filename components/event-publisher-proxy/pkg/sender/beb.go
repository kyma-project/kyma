package sender

import (
	"context"
	"net/http"
)

// BEBMessageSender is responsible for sending messages over HTTP.
type BEBMessageSender struct {
	Client *http.Client
	Target string
}

// NewBEBMessageSender returns a new BEBMessageSender instance with the given target and client.
func NewBEBMessageSender(target string, client *http.Client) *BEBMessageSender {
	return &BEBMessageSender{Client: client, Target: target}
}

// NewRequestWithTarget returns a new HTTP POST request with the given context and target.
func (s *BEBMessageSender) NewRequestWithTarget(ctx context.Context, target string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
}

// Send sends the given HTTP request and returns the HTTP response back.
func (s *BEBMessageSender) Send(req *http.Request) (*http.Response, error) {
	return s.Client.Do(req)
}
