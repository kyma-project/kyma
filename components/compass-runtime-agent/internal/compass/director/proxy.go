package director

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type ProxyConfig struct {
	Port int
	InsecureSkipVerify bool `envconfig:"default=false"`
}

// serviceUnavailableHandler returns a simple request handler that replies to each request with a 503 status code.
func serviceUnavailableHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Proxy to Director is not initialized. Try again.", http.StatusServiceUnavailable)
	})
}

// Proxy provides functionality for proxy'ing upcomming traffic to Director server
// and attaches required client certificates
type Proxy struct {
	mux                sync.RWMutex
	port               int
	proxy              http.Handler
	insecureSkipVerify bool
}

func NewProxy(cfg ProxyConfig) *Proxy {
	return &Proxy{
		port:               cfg.Port,
		proxy:              serviceUnavailableHandler(),
		insecureSkipVerify: cfg.InsecureSkipVerify,
	}
}

// UpdateCertAndURL updates the underlying proxy for Director server.
func (p *Proxy) UpdateCertAndURL(cert *tls.Certificate, directorURL string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	parsed, err := url.Parse(directorURL)
	if err != nil {
		return err
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig.Certificates = []tls.Certificate{*cert}
	transport.TLSClientConfig.InsecureSkipVerify = p.insecureSkipVerify

	rproxy := httputil.NewSingleHostReverseProxy(parsed)
	rproxy.Transport = transport

	// update proxy instance
	p.proxy = rproxy

	return nil
}

// ServeHTTP fulfills the http.Handler interface.
// It handles traffic to external Director server.
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.mux.RLock()
	defer p.mux.RUnlock()

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
