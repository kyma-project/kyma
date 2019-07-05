package main

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const (
	allowOrigin = "domain.com"
	allowMethod = "GET"
	allowHeader = "Content-Type"
)

func TestDeleteUstreamCORSHeaders(t *testing.T) {

	//GIVEN

	//Setup backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(corsAllowOriginHeader, "upstream-cors-allow-origin")
		w.Header().Set(corsAllowMethodsHeader, "upstream-cors-allow-method")
		w.Header().Set(corsAllowHeadersHeader, "upstream-cors-allow-header")
	}))

	defer backend.Close()
	backendURL, _ := url.Parse(backend.URL)

	//Setup proxy
	rp, _ := newReverseProxy(backendURL, &rest.Config{}, true)

	//Setup CORS config
	cfg := corsConfig{
		allowOrigin:  []string{allowOrigin},
		allowMethods: []string{allowMethod},
		allowHeaders: []string{allowHeader},
	}

	//Setup CORS wrapper
	corsHandler := getCORSHandler(rp, cfg)

	//Prepare request
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", allowOrigin)

	//Prepare ResponseWriter
	res := httptest.NewRecorder()

	//WHEN
	corsHandler.ServeHTTP(res, req)

	//THEN
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, 3, len(res.Header())) //expected headers: corsAllowOriginHeader, Content-Length, Date
	assert.Equal(t, allowOrigin, res.Header().Get(corsAllowOriginHeader))
}
