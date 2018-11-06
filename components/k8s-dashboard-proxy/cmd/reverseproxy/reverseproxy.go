package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

const (
	dashboardURL          = "http://localhost:30000"
	defaultSecretFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	port                  = ":8080"
)

var (
	proxy    *httputil.ReverseProxy
	proxyURL *url.URL
	sfp      *secret
)

type secret struct {
	token string
	mux   sync.Mutex
}

func (sfp *secret) updateSecret() {
	sfp.mux.Lock()
	var accountSecret []byte
	if token, err := ioutil.ReadFile(defaultSecretFilePath); err != nil {
		log.Printf("Error: read service account failed: %v", err)
	} else {
		accountSecret = token
	}
	sfp.token = string(accountSecret)
	sfp.mux.Unlock()
}

func serveProxy(rw http.ResponseWriter, req *http.Request) {
	req.URL.Host = proxyURL.Host
	req.URL.Scheme = proxyURL.Scheme
	req.Header.Set("Authorization", "Bearer "+sfp.token)
	req.Host = proxyURL.Host
	proxy.ServeHTTP(rw, req)
}

func handleRequest(rw http.ResponseWriter, req *http.Request) {
	log.Printf("handling new request...")
	serveProxy(rw, req)
}

func main() {
	log.Printf("starting kubernetes dashboard reverse proxy...")
	sfp = &secret{token: defaultSecretFilePath}
	sfp.updateSecret()
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			sfp.updateSecret()
		}
	}()
	var err error
	proxyURL, err = url.Parse(dashboardURL)
	log.Printf("proxied URL: %v", proxyURL)
	if err != nil {
		log.Fatalf("failed in parsing the kubernetes kubernetes dashboard URL: %v", err)
	}
	proxy = httputil.NewSingleHostReverseProxy(proxyURL)

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(port, nil))
}
