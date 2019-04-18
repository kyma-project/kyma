// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import clientcontext "github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

import mock "github.com/stretchr/testify/mock"

// LookupService is an autogenerated mock type for the LookupService type
type LookupService struct {
	mock.Mock
}

// Fetch provides a mock function with given fields: context, configFilePath
func (_m LookupService) Fetch(context clientcontext.ApplicationContext, configFilePath string) (string, error) {
	ret := _m.Called(context, configFilePath)

	var r0 string
	if rf, ok := ret.Get(0).(func(clientcontext.ApplicationContext, string) string); ok {
		r0 = rf(context, configFilePath)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(clientcontext.ApplicationContext, string) error); ok {
		r1 = rf(context, configFilePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
