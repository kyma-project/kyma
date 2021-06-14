// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	apperrors "github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"

	mock "github.com/stretchr/testify/mock"
)

// OAuthClient is an autogenerated mock type for the OAuthClient type
type OAuthClient struct {
	mock.Mock
}

// GetToken provides a mock function with given fields: clientID, clientSecret, authURL, headers, queryParameters
func (_m *OAuthClient) GetToken(clientID string, clientSecret string, authURL string, headers *map[string][]string, queryParameters *map[string][]string) (string, apperrors.AppError) {
	ret := _m.Called(clientID, clientSecret, authURL, headers, queryParameters)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, string, *map[string][]string, *map[string][]string) string); ok {
		r0 = rf(clientID, clientSecret, authURL, headers, queryParameters)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, string, string, *map[string][]string, *map[string][]string) apperrors.AppError); ok {
		r1 = rf(clientID, clientSecret, authURL, headers, queryParameters)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// InvalidateTokenCache provides a mock function with given fields: clientID
func (_m *OAuthClient) InvalidateTokenCache(clientID string) {
	_m.Called(clientID)
}
