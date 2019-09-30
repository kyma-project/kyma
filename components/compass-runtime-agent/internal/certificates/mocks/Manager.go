// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	certificates "kyma-project.io/compass-runtime-agent/internal/certificates"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// GetClientCredentials provides a mock function with given fields:
func (_m *Manager) GetClientCredentials() (certificates.ClientCredentials, error) {
	ret := _m.Called()

	var r0 certificates.ClientCredentials
	if rf, ok := ret.Get(0).(func() certificates.ClientCredentials); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(certificates.ClientCredentials)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PreserveCredentials provides a mock function with given fields: _a0
func (_m *Manager) PreserveCredentials(_a0 certificates.Credentials) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(certificates.Credentials) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
