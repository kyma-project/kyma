package spdy

import (
	"io"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/httpstream/spdy"
	"k8s.io/client-go/rest"
	client_spdy "k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
)

type Proxy struct {
	kubeconfig  *rest.Config
	upstreamUrl *url.URL
}

func New(kubeconfig *rest.Config, upstreamUrl *url.URL) *Proxy {
	return &Proxy{kubeconfig: kubeconfig, upstreamUrl: upstreamUrl}
}

func (p *Proxy) IsSpdyRequest(req *http.Request) bool {
	return req.Header.Get("upgrade") == "SPDY/3.1"
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	clientTransport, upgrader, err := client_spdy.RoundTripperFor(p.kubeconfig)
	if err != nil {
		panic(err)
	}

	client := &http.Client{Transport: clientTransport}

	protocols := req.Header.Get("X-Stream-Protocol-Version")
	clientUrl, _ := url.Parse(p.upstreamUrl.String())
	clientUrl.Path = req.URL.Path
	clientUrl.RawQuery = req.URL.RawQuery
	clientReq, err := http.NewRequest(req.Method, clientUrl.String(), req.Body)
	if err != nil {
		panic(err)
	}

	clientReq.Header.Set("upgrade", "SPDY/3.1")
	clientReq.Header.Set("connection", "upgrade")

	serverConnection, s, err := client_spdy.Negotiate(upgrader, client, clientReq, protocols)
	if err != nil {
		panic(err)
	}
	w.Header().Set(httpstream.HeaderProtocolVersion, s)
	clientConnection := spdy.NewResponseUpgrader().UpgradeResponse(w, req, func(clientStream httpstream.Stream, replySent <-chan struct{}) error {
		serverStream, err := serverConnection.CreateStream(clientStream.Headers())
		if err != nil {
			return err
		}

		go io.Copy(clientStream, serverStream)
		go io.Copy(serverStream, clientStream)
		return nil
	})

	go func() {
		<-serverConnection.CloseChan()
		clientConnection.Close()
	}()
}
