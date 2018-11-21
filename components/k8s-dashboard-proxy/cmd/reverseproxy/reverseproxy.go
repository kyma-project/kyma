package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/k8s-dashboard-proxy/util"
)

var (
	dashboardURL   string
	secretFilePath string
	port           string
)

var (
	proxy     *httputil.ReverseProxy
	proxyURL  *url.URL
	sfp       *secret
	proxyOpts *util.Options
)

type secret struct {
	token string
	mux   sync.Mutex
}

func (sfp *secret) updateSecret() {
	sfp.mux.Lock()
	defer sfp.mux.Unlock()
	if token, err := ioutil.ReadFile(secretFilePath); err != nil {
		log.Printf("Error: read service account failed: %v", err)
	} else {
		sfp.token = string(token)
	}
}

func (sfp *secret) readSecret() string {
	sfp.mux.Lock()
	defer sfp.mux.Unlock()
	return sfp.token
}

func serveProxy(rw http.ResponseWriter, req *http.Request) {
	req.URL.Host = proxyURL.Host
	req.URL.Scheme = proxyURL.Scheme
	req.Header.Set("Authorization", "Bearer "+sfp.readSecret())
	req.Host = proxyURL.Host
	proxy.ServeHTTP(rw, req)
}

func handleRequest(rw http.ResponseWriter, req *http.Request) {
	log.Printf("handling new request...")
	serveProxy(rw, req)
}

func main() {
	log.Printf("starting kubernetes dashboard reverse proxy...")
	proxyOpts = util.ParseFlags()
	dashboardURL = proxyOpts.DashboardURL
	secretFilePath = proxyOpts.SecretFilePath
	port = proxyOpts.Port
	sfp = &secret{token: "INVALID TOKEN"}
	sfp.updateSecret()
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			sfp.updateSecret()
		}
	}()
	var err error
	proxyURL, err = url.Parse(dashboardURL)
	log.Printf("kubernetes dashboard reverse proxy started and listening on port%v", port)
	log.Printf("proxied URL: %v", proxyURL)
	if err != nil {
		log.Fatalf("failed in parsing the kubernetes kubernetes dashboard URL: %v", err)
	}
	proxy = httputil.NewSingleHostReverseProxy(proxyURL)

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(port, nil))
}
