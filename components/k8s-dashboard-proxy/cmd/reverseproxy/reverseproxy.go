package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	dashboardURL   = "http://localhost:30000"
	secretFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	port           = ":8080"
)

var (
	proxy    *httputil.ReverseProxy
	proxyURL *url.URL
)

func readSecret() string {
	var accountSecret []byte
	if token, err := ioutil.ReadFile(secretFilePath); err != nil {
		log.Printf("Error: read service account failed: %v", err)
		accountSecret = token
		return ""
	}
	return string(accountSecret)
}

func serveProxy(rw http.ResponseWriter, req *http.Request) {
	req.URL.Host = proxyURL.Host
	req.URL.Scheme = proxyURL.Scheme
	req.Header.Set("Authorization", "Bearer "+readSecret())
	req.Host = proxyURL.Host
	proxy.ServeHTTP(rw, req)
}

func handleRequest(rw http.ResponseWriter, req *http.Request) {
	log.Printf("handling new request...")
	serveProxy(rw, req)
}

func main() {
	log.Printf("starting kubernetes dashboard reverse proxy...")
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
