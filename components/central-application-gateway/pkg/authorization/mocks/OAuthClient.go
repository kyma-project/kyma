// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	apperrors "github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"

	mock "github.com/stretchr/testify/mock"

	tls "crypto/tls"
)

// OAuthClient is an autogenerated mock type for the OAuthClient type
type OAuthClient struct {
	mock.Mock
}

// GetToken provides a mock function with given fields: clientID, clientSecret, authURL, headers, queryParameters, skipTLSVerification
func (_m *OAuthClient) GetToken(clientID string, clientSecret string, authURL string, headers *map[string][]string, queryParameters *map[string][]string, skipTLSVerification bool) (string, apperrors.AppError) {
	ret := _m.Called(clientID, clientSecret, authURL, headers, queryParameters, skipTLSVerification)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, string, *map[string][]string, *map[string][]string, bool) string); ok {
		r0 = rf(clientID, clientSecret, authURL, headers, queryParameters, skipTLSVerification)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, string, string, *map[string][]string, *map[string][]string, bool) apperrors.AppError); ok {
		r1 = rf(clientID, clientSecret, authURL, headers, queryParameters, skipTLSVerification)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// GetTokenMTLS provides a mock function with given fields: clientID, authURL, cert, headers, queryParameters, skipTLSVerification
func (_m *OAuthClient) GetTokenMTLS(clientID string, authURL string, cert tls.Certificate, headers *map[string][]string, queryParameters *map[string][]string, skipTLSVerification bool) (string, apperrors.AppError) {
	ret := _m.Called(clientID, authURL, cert, headers, queryParameters, skipTLSVerification)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, tls.Certificate, *map[string][]string, *map[string][]string, bool) string); ok {
		r0 = rf(clientID, authURL, cert, headers, queryParameters, skipTLSVerification)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, string, tls.Certificate, *map[string][]string, *map[string][]string, bool) apperrors.AppError); ok {
		r1 = rf(clientID, authURL, cert, headers, queryParameters, skipTLSVerification)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// InvalidateTokenCache provides a mock function with given fields: clientID, authURL
func (_m *OAuthClient) InvalidateTokenCache(clientID string, authURL string) {
	_m.Called(clientID, authURL)
}

type mockConstructorTestingTNewOAuthClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewOAuthClient creates a new instance of OAuthClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewOAuthClient(t mockConstructorTestingTNewOAuthClient) *OAuthClient {
	mock := &OAuthClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
