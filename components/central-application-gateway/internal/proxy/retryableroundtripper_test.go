package proxy

import (
	"crypto/tls"
	csrfMock "github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	authMock "github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRetryableRoundTripper(t *testing.T) {

	defaultAuthStrategyMock := func() *authMock.Strategy {
		return &authMock.Strategy{}
	}
	retryAuthStrategyMock := func() *authMock.Strategy {
		result := &authMock.Strategy{}
		result.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("SetClientCertificateFunc")).
			Return(nil).
			Once()
		result.On("Invalidate").Return().Once()
		return result
	}

	defaultCsrfTokenStrategyMock := func() *csrfMock.TokenStrategy {
		return &csrfMock.TokenStrategy{}
	}
	retryCsrfTokenStrategyMock := func() *csrfMock.TokenStrategy {
		result := &csrfMock.TokenStrategy{}
		result.On("AddCSRFToken", mock.AnythingOfType("*http.Request")).Return(nil)
		result.On("Invalidate").Return().Once()
		return result
	}

	type serverResponse struct {
		statusCode int
		body       string
	}

	tests := []struct {
		name                  string
		authStrategyFunc      func() *authMock.Strategy
		csrfTokenStrategyFunc func() *csrfMock.TokenStrategy
		requestBody           string
		serverResponses       []serverResponse
		expectedStatusCode    int
		expectedBody          string
		expectedClientCert    *tls.Certificate
	}{
		{
			name:                  "Success",
			expectedStatusCode:    http.StatusOK,
			authStrategyFunc:      defaultAuthStrategyMock,
			csrfTokenStrategyFunc: defaultCsrfTokenStrategyMock,
		},
		{
			name:                  "Internal error",
			expectedStatusCode:    http.StatusInternalServerError,
			expectedBody:          "internal error",
			authStrategyFunc:      defaultAuthStrategyMock,
			csrfTokenStrategyFunc: defaultCsrfTokenStrategyMock,
			serverResponses: []serverResponse{
				{
					statusCode: http.StatusInternalServerError,
					body:       "internal error",
				},
			},
		},
		{
			name:                  "Retry on 403 and success",
			expectedStatusCode:    http.StatusOK,
			expectedBody:          "success",
			authStrategyFunc:      retryAuthStrategyMock,
			csrfTokenStrategyFunc: retryCsrfTokenStrategyMock,
			serverResponses: []serverResponse{
				{
					statusCode: http.StatusForbidden,
					body:       "error",
				},
				{
					statusCode: http.StatusOK,
					body:       "success",
				},
			},
		},
		{
			name:                  "Retry on 403 and failure",
			expectedStatusCode:    http.StatusForbidden,
			expectedBody:          "error 2",
			authStrategyFunc:      retryAuthStrategyMock,
			csrfTokenStrategyFunc: retryCsrfTokenStrategyMock,
			serverResponses: []serverResponse{
				{
					statusCode: http.StatusForbidden,
					body:       "error 1",
				},
				{
					statusCode: http.StatusForbidden,
					body:       "error 2",
				},
			},
		},
		{
			name:                  "Retry on 401 and success",
			expectedStatusCode:    http.StatusOK,
			expectedBody:          "success",
			authStrategyFunc:      retryAuthStrategyMock,
			csrfTokenStrategyFunc: retryCsrfTokenStrategyMock,
			serverResponses: []serverResponse{
				{
					statusCode: http.StatusUnauthorized,
					body:       "error",
				},
				{
					statusCode: http.StatusOK,
					body:       "success",
				},
			},
		},
		{
			name:                  "Retry on 401 and failure",
			expectedStatusCode:    http.StatusInternalServerError,
			expectedBody:          "error 2",
			authStrategyFunc:      retryAuthStrategyMock,
			csrfTokenStrategyFunc: retryCsrfTokenStrategyMock,
			serverResponses: []serverResponse{
				{
					statusCode: http.StatusUnauthorized,
					body:       "error 1",
				},
				{
					statusCode: http.StatusInternalServerError,
					body:       "error 2",
				},
			},
		},
	}
	for _, tc := range tests {
		var requestCount int
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.serverResponses == nil || len(tc.serverResponses) <= requestCount {
					w.WriteHeader(http.StatusOK)
				} else {
					serverResponse := tc.serverResponses[requestCount]
					w.WriteHeader(serverResponse.statusCode)
					w.Write([]byte(serverResponse.body))
				}
				requestCount++
			}))
			defer ts.Close()

			authStrategyMock := tc.authStrategyFunc()
			csrfTokenStrategyMock := tc.csrfTokenStrategyFunc()
			clientCertificate := clientcert.NewClientCertificate(nil)

			transport := NewRetryableRoundTripper(http.DefaultTransport, authStrategyMock, csrfTokenStrategyMock, clientCertificate, 10)
			httpClient := &http.Client{
				Transport: transport,
			}
			req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(tc.requestBody))
			require.NoError(t, err)

			res, err := httpClient.Do(req)
			require.NoError(t, err)

			resBody, err := io.ReadAll(res.Body)
			_ = res.Body.Close()
			require.NoError(t, err)
			require.Equal(t, res.StatusCode, tc.expectedStatusCode)
			require.Equal(t, string(resBody), tc.expectedBody)
			require.Equal(t, clientCertificate.GetCertificate(), tc.expectedClientCert)

			authStrategyMock.AssertExpectations(t)
			csrfTokenStrategyMock.AssertExpectations(t)
		})
	}
}
