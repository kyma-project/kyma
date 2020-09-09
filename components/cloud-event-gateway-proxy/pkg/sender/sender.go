package sender

import (
	"context"
	"net/http"
)

type HttpMessageSender struct {
	Client *http.Client
	Target string
}

func NewHttpMessageSender(target string, client *http.Client) (*HttpMessageSender, error) {
	return &HttpMessageSender{Client: client, Target: target}, nil
}

func (s *HttpMessageSender) NewRequestWithTarget(ctx context.Context, target string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
}

func (s *HttpMessageSender) Send(req *http.Request) (*http.Response, error) {
	return s.Client.Do(req)
}
