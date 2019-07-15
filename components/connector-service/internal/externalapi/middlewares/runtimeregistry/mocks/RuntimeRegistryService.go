// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import runtimeregistry "github.com/kyma-project/kyma/components/connector-service/internal/externalapi/middlewares/runtimeregistry"

// RuntimeRegistryService is an autogenerated mock type for the RuntimeRegistryService type
type RuntimeRegistryService struct {
	mock.Mock
}

// ReportState provides a mock function with given fields: state
func (_m *RuntimeRegistryService) ReportState(state runtimeregistry.RuntimeState) error {
	ret := _m.Called(state)

	var r0 error
	if rf, ok := ret.Get(0).(func(runtimeregistry.RuntimeState) error); ok {
		r0 = rf(state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
