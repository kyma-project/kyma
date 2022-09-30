package proxy

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLRewriter(t *testing.T) {
	type redirectTest struct {
		name      string
		gwURL     string
		targetURL string
		location  string
		expected  string
		empty     bool
	}

	tests := []redirectTest{
		{
			name:      "Simple redirect",
			gwURL:     "http://central-gateway.cluser.local/app/service",
			targetURL: "https://httpbin.org/api/v1",
			location:  "https://httpbin.org/api/v1/some/sub/path",
			expected:  "http://central-gateway.cluser.local/app/service/some/sub/path",
		},
		{
			name:      "Relative Path",
			gwURL:     "http://central-gateway.cluser.local/app/service",
			targetURL: "https://httpbin.org/api/v1",
			location:  "/some/sub/path",
			empty:     true,
		},
		{
			name:      "Changed Host",
			gwURL:     "http://central-gateway.cluser.local/app/service",
			targetURL: "https://httpbin.org/api/v1",
			location:  "https://otherService.org/api/v1/some/sub/path",
			empty:     true,
		},
		{
			name:      "Changed Subpath",
			gwURL:     "http://central-gateway.cluser.local/app/service",
			targetURL: "https://httpbin.org/api/v1",
			location:  "https://httpbin.org/api/v2/some/sub/path",
			empty:     true,
		},
		{
			name:      "Changed Protocol",
			gwURL:     "http://central-gateway.cluser.local/app/service",
			targetURL: "https://httpbin.org/api/v1",
			location:  "ftp://httpbin.org/api/v1/some/sub/path",
			empty:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gwURL, err := url.Parse(tc.gwURL)
			require.Nil(t, err)

			targetURL, err := url.Parse(tc.targetURL)
			require.Nil(t, err)

			location, err := url.Parse(tc.location)
			require.Nil(t, err)

			newURL := urlRewriter(gwURL, targetURL, location)

			if tc.empty {
				assert.Nil(t, newURL)
			} else {
				require.NotNil(t, newURL)

				actual := newURL.String()
				t.Log("Got", actual)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestResponseModifier(t *testing.T) {
	type testcase struct {
		name        string
		request     func() *http.Request
		response    *http.Response
		urlRewriter func(t *testing.T) func(gatewayURL, target, loc *url.URL) *url.URL
		validate    func(t *testing.T, r *http.Response, called bool)
	}

	const locationHeader = "Location"
	const target = "https://httpbin.org/api/v1"
	const gateway = "http://central-gw.cluster.local"

	tests := []testcase{
		{
			name: "Temporary redirect",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, gateway+"/some/url", nil)
				return req
			},
			response: &http.Response{
				StatusCode: http.StatusTemporaryRedirect,
				Header:     http.Header{locationHeader: []string{target + "/some/endpoint"}},
			},
			urlRewriter: func(t *testing.T) func(gatewayURL, target, loc *url.URL) *url.URL {
				return func(gatewayURL, trgt, loc *url.URL) *url.URL {
					assert.Equal(t, target+"/some/endpoint", loc.String())

					u, err := url.Parse("https://other.addr/with/path")
					require.Nil(t, err)
					return u
				}
			},
			validate: func(t *testing.T, r *http.Response, called bool) {
				assert.True(t, called, "url rewrite should have been called")
				newLoc := r.Header.Get(locationHeader)
				assert.Equal(t, "https://other.addr/with/path", newLoc)
			},
		},
		{
			name: "Temporary redirect without location",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, gateway+"/some/url", nil)
				return req
			},
			response: &http.Response{
				StatusCode: http.StatusTemporaryRedirect,
			},
			validate: func(t *testing.T, r *http.Response, called bool) {
				assert.False(t, called, "url rewrite shouldn't have been called")

				_, exists := r.Header[locationHeader]
				assert.False(t, exists, "location header shouldn't have been added")
			},
		},
		{
			name: "200 Ok",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, gateway+"/some/url", nil)
				return req
			},
			response: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{locationHeader: []string{target + "/some/endpoint"}},
			},
			validate: func(t *testing.T, r *http.Response, called bool) {
				assert.False(t, called, "url rewrite shouldn't have been called")
				newLoc := r.Header.Get(locationHeader)
				assert.Equal(t, target+"/some/endpoint", newLoc)
			},
		},
		{
			name: "201 Created",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, gateway+"/some/url", nil)
				return req
			},
			response: &http.Response{
				StatusCode: http.StatusCreated,
				Header:     http.Header{locationHeader: []string{target + "/some/endpoint"}},
			},
			urlRewriter: func(t *testing.T) func(gatewayURL, target, loc *url.URL) *url.URL {
				return func(gatewayURL, trgt, loc *url.URL) *url.URL {
					assert.Equal(t, target+"/some/endpoint", loc.String())

					u, err := url.Parse("https://other.addr/with/path")
					require.Nil(t, err)
					return u
				}
			},
			validate: func(t *testing.T, r *http.Response, called bool) {
				assert.True(t, called, "url rewrite should have been called")

				newLoc := r.Header.Get(locationHeader)
				assert.Equal(t, "https://other.addr/with/path", newLoc)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			called := false
			res := tc.response
			res.Request = tc.request()

			rewriter := func(gatewayURL, target, loc *url.URL) *url.URL {
				called = true

				if tc.urlRewriter == nil {
					return nil
				}

				return tc.urlRewriter(t)(gatewayURL, target, loc)
			}

			gw, err := url.Parse(gateway) // this could be out of the loop, but here it felt more readable
			require.Nil(t, err)

			rm := responseModifier(gw, target, rewriter)

			// when
			err = rm(res)
			require.Nil(t, err)

			// then
			tc.validate(t, res, called)
		})
	}
}
