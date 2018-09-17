package graphql

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

const ingressGatewayControllerServiceURL = "istio-ingressgateway.istio-system.svc.cluster.local"

type authenticatedTransport struct {
	http.Transport
	token string
}

func newAuthorizedClient(token string, dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) *http.Client {
	transport := &authenticatedTransport{
		Transport: http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext:     dialContext,
		},
		token: token,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	return t.Transport.RoundTrip(req)
}
