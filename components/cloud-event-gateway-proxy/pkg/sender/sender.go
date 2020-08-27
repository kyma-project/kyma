package sender

import (
	"context"
	"net/http"

	"go.opencensus.io/plugin/ochttp"
)

// ConnectionArgs allow to configure connection parameters to the underlying
// HTTP Client transport.
type ConnectionArgs struct {
	// MaxIdleConns refers to the max idle connections, as in net/http/transport.
	MaxIdleConns int
	// MaxIdleConnsPerHost refers to the max idle connections per host, as in net/http/transport.
	MaxIdleConnsPerHost int
}

func (ca *ConnectionArgs) ConfigureTransport(transport *http.Transport) {
	if ca == nil {
		return
	}

	transport.MaxIdleConns = ca.MaxIdleConns
	transport.MaxIdleConnsPerHost = ca.MaxIdleConnsPerHost
}

type HttpMessageSender struct {
	Client *http.Client
	Target string
}

func NewHttpMessageSender(connectionArgs *ConnectionArgs, target string, httpClient *http.Client) (*HttpMessageSender, error) {
	// Add connection options to the default transport.
	var base = http.DefaultTransport.(*http.Transport).Clone()
	connectionArgs.ConfigureTransport(base)
	httpClient.Transport = &ochttp.Transport{Base: base}
	return &HttpMessageSender{Client: httpClient, Target: target}, nil
}

func (s *HttpMessageSender) NewCloudEventRequest(ctx context.Context) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodPost, s.Target, nil)
}

func (s *HttpMessageSender) NewCloudEventRequestWithTarget(ctx context.Context, target string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
}

func (s *HttpMessageSender) Send(req *http.Request) (*http.Response, error) {
	return s.Client.Do(req)
}
