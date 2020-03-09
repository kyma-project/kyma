package director

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

// ProxyConfig holds configuration for Director proxy
type ProxyConfig struct {
	Port               int
	InsecureSkipVerify bool `envconfig:"default=false"`
}

// unavailableHandler returns a simple request handler that replies to each request with a 503 status code.
func unavailableHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Proxy to Director is not initialized. Try again later.", http.StatusServiceUnavailable)
	})
}

// Proxy provides functionality for proxy'ing upcomming traffic to Director server
// and attaches required client certificates
type Proxy struct {
	port               int
	insecureSkipVerify bool
	mux                sync.RWMutex
	proxy              http.Handler
	targetURL          *url.URL
	transport          *http.Transport
}

// NewProxy returns new instance of Proxy
func NewProxy(cfg ProxyConfig) *Proxy {
	return &Proxy{
		port:               cfg.Port,
		insecureSkipVerify: cfg.InsecureSkipVerify,
		transport:          http.DefaultTransport.(*http.Transport).Clone(),
	}
}

// SetURLAndCerts updates the underlying proxy for Director server.
func (p *Proxy) SetURLAndCerts(directorURL string, cert *tls.Certificate) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	targetURL, err := url.Parse(directorURL)
	if err != nil {
		return errors.Wrapf(err, "while parsing given URL %q", directorURL)
	}

	p.targetURL = targetURL
	p.transport.TLSClientConfig.Certificates = []tls.Certificate{*cert}
	p.transport.TLSClientConfig.InsecureSkipVerify = p.insecureSkipVerify

	// proxy instance "lazy init"
	if p.proxy == nil {
		p.proxy = &httputil.ReverseProxy{
			Director:  p.director,
			Transport: p.transport,
		}
	}

	return nil
}

// ServeHTTP fulfills the http.Handler interface.
// It handles traffic to external Director server.
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	if p.proxy == nil {
		unavailableHandler().ServeHTTP(rw, req)
		return
	}

	p.proxy.ServeHTTP(rw, req)
}

// Start fulfills the controller manager Runnable interface.
// It starts the HTTP server on given port and support graceful shutdown.
func (p *Proxy) Start(stop <-chan struct{}) error {
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", p.port),
		Handler: p,
	}

	idleConnsClosed := make(chan struct{})
	var shutdownErr error
	go func() {
		<-stop
		// we received an stop signal, shut down.
		shutdownErr = srv.Shutdown(context.Background())
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	// when Shutdown is called then ListenAndServe immediately returns ErrServerClosed,
	// so we need to wait instead for Shutdown to return.
	<-idleConnsClosed

	return shutdownErr
}

// director is copied from the Go 1.13 implementation
// github.com/golang/go/blob/release-branch.go1.13/src/net/http/httputil/reverseproxy.go#L107 with two changes:
//
// * behaviour for `singleJoiningSlash` function
// * setting `req.Host` field
//
func (p *Proxy) director(req *http.Request) {
	targetQuery := p.targetURL.RawQuery

	req.URL.Scheme = p.targetURL.Scheme
	req.URL.Host = p.targetURL.Host
	req.URL.Path = singleJoiningSlash(p.targetURL.Path, req.URL.Path)

	// see: https://github.com/golang/go/issues/28168
	req.Host = p.targetURL.Host

	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
}

// singleJoiningSlash was copied from the Go 1.13 implementation
// https://github.com/golang/go/blob/release-branch.go1.13/src/net/http/httputil/reverseproxy.go#L88
//
// Problem with the core implementation is that it adds additional slash at the end of path causing 404 when calling
// Director URL because /director/graphql != /director/graphql/
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case b == "", b == "/":
		return a
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
