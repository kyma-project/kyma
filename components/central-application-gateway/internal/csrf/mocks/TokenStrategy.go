// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	apperrors "github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"

	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// TokenStrategy is an autogenerated mock type for the TokenStrategy type
type TokenStrategy struct {
	mock.Mock
}

// AddCSRFToken provides a mock function with given fields: apiRequest, skipTLSVerify
func (_m *TokenStrategy) AddCSRFToken(apiRequest *http.Request, skipTLSVerify bool) apperrors.AppError {
	ret := _m.Called(apiRequest, skipTLSVerify)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(*http.Request, bool) apperrors.AppError); ok {
		r0 = rf(apiRequest, skipTLSVerify)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// Invalidate provides a mock function with given fields:
func (_m *TokenStrategy) Invalidate() {
	_m.Called()
}

type mockConstructorTestingTNewTokenStrategy interface {
	mock.TestingT
	Cleanup(func())
}

// NewTokenStrategy creates a new instance of TokenStrategy. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTokenStrategy(t mockConstructorTestingTNewTokenStrategy) *TokenStrategy {
	mock := &TokenStrategy{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
