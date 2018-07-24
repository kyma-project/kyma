package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterMiddlewareFiltersRequests(t *testing.T) {
	for tn, tc := range map[string]struct {
		givenURLPath string
		expectedCode int
	}{
		"index.yaml": {
			givenURLPath: "index.yaml",
			expectedCode: http.StatusOK,
		},
		"tgz archive": {
			givenURLPath: "archive.tgz",
			expectedCode: http.StatusOK,
		},
		"Not allowed extension": {
			givenURLPath: "archive.html",
			expectedCode: http.StatusNotFound,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost/%s", tc.givenURLPath), nil)
			recorder := httptest.NewRecorder()
			called := false
			next := func(rw http.ResponseWriter, r *http.Request) {
				called = true
			}

			// when
			filteringMiddleware(recorder, req, next)

			// then
			assert.Equal(t, tc.expectedCode, recorder.Code)
			// next must be called only if code is 200
			assert.Equal(t, http.StatusOK == recorder.Code, called)
		})
	}
}

func TestContentTypeSetting(t *testing.T) {
	for tn, tc := range map[string]struct {
		givenURLPath        string
		expectedContentType string
	}{
		"index.yaml": {
			givenURLPath:        "index.yaml",
			expectedContentType: "text/plain",
		},
		"tgz archive": {
			givenURLPath:        "archive.tgz",
			expectedContentType: "application/gzip",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost/%s", tc.givenURLPath), nil)
			recorder := httptest.NewRecorder()
			next := func(rw http.ResponseWriter, r *http.Request) {} // handler stub

			// when
			contentTypeMiddleware(recorder, req, next)

			// then
			assert.Equal(t, tc.expectedContentType, recorder.Header().Get("Content-Type"))
		})
	}
}
