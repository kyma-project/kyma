package httptools

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type RoundTripper struct {
	transport *http.Transport
}

type RoundTripperOption func(*RoundTripper)

func WithTLSConfig(config *tls.Config) RoundTripperOption {
	return func(rt *RoundTripper) {
		rt.transport.TLSClientConfig = config
	}
}

func WithTLSSkipVerify(skipVerify bool) RoundTripperOption {
	return func(rt *RoundTripper) {
		if rt.transport.TLSClientConfig == nil {
			rt.transport.TLSClientConfig = &tls.Config{}
		}
		rt.transport.TLSClientConfig.InsecureSkipVerify = skipVerify
	}
}

func WithGetClientCertificate(f func(*tls.CertificateRequestInfo) (*tls.Certificate, error)) RoundTripperOption {
	return func(rt *RoundTripper) {
		if rt.transport.TLSClientConfig == nil {
			rt.transport.TLSClientConfig = &tls.Config{}
		}
		rt.transport.TLSClientConfig.GetClientCertificate = f
	}
}

func NewRoundTripper(options ...RoundTripperOption) *RoundTripper {
	rt := &RoundTripper{
		transport: newDefaultTransport(),
	}
	for _, option := range options {
		option(rt)
	}
	return rt
}

func (p *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return p.transport.RoundTrip(req)
}

func newDefaultTransport() *http.Transport {
	// http.DefaultTransport
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
