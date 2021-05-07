// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	apperrors "github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	mock "github.com/stretchr/testify/mock"

	proxyconfig "github.com/kyma-project/kyma/components/central-application-gateway/pkg/proxyconfig"
)

// TargetConfigProvider is an autogenerated mock type for the TargetConfigProvider type
type TargetConfigProvider struct {
	mock.Mock
}

// GetDestinationConfig provides a mock function with given fields: secretName, apiName
func (_m *TargetConfigProvider) GetDestinationConfig(secretName string, apiName string) (proxyconfig.ProxyDestinationConfig, apperrors.AppError) {
	ret := _m.Called(secretName, apiName)

	var r0 proxyconfig.ProxyDestinationConfig
	if rf, ok := ret.Get(0).(func(string, string) proxyconfig.ProxyDestinationConfig); ok {
		r0 = rf(secretName, apiName)
	} else {
		r0 = ret.Get(0).(proxyconfig.ProxyDestinationConfig)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, string) apperrors.AppError); ok {
		r1 = rf(secretName, apiName)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
