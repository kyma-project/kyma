package proxy

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyFactory(t *testing.T) {

	proxyTimeout := 10
	//apiIdentifier := model.APIIdentifier{
	//	Application: "app",
	//	Service:     "service",
	//	Entry:       "entry",
	//}

	t.Run("should proxy using application and service name", func(t *testing.T) {
		// TODO
	})

	t.Run("should proxy using application, service and entry name", func(t *testing.T) {
		// TODO
	})

	type testcase struct {
		name       string
		url        string
		createFunc func() http.Handler
	}

	proxyConfig := Config{
		SkipVerify:    true,
		ProxyTimeout:  proxyTimeout,
		Application:   "test",
		ProxyCacheTTL: proxyTimeout,
	}

	createHandler := func() http.Handler {
		return New(nil, nil, nil, proxyConfig)
	}

	createHandlerForCompass := func() http.Handler {
		return NewForCompass(nil, nil, nil, proxyConfig)
	}

	testCases := []testcase{
		{
			name:       "Should return Internal error when failed to extract data from empty path",
			url:        "",
			createFunc: createHandler,
		},
		{
			name:       "Should return Internal error when failed to extract data from path containing application name only",
			url:        "/appName",
			createFunc: createHandler,
		},
		{
			name:       "Should return Internal error when failed to extract data from empty path (Compass)",
			url:        "",
			createFunc: createHandlerForCompass,
		},
		{
			name:       "Should return Internal error when failed to extract data from path containing application name only (Compass)",
			url:        "/appName",
			createFunc: createHandlerForCompass,
		},
		{
			name:       "Should return Internal error when failed to extract data from path containing application and service name only (Compass)",
			url:        "/appName/serviceName",
			createFunc: createHandlerForCompass,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// given
			handler := testCase.createFunc()

			req, err := http.NewRequest(http.MethodGet, testCase.url, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()

			// when
			handler.ServeHTTP(rr, req)

			// then
			assert.Equal(t, http.StatusInternalServerError, rr.Code)
		})
	}
}
