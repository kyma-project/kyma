// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// NameResolver is an autogenerated mock type for the NameResolver type
type NameResolver struct {
	mock.Mock
}

// GetCredentialsSecretName provides a mock function with given fields: application, bundleID
func (_m *NameResolver) GetCredentialsSecretName(application string, bundleID string) string {
	ret := _m.Called(application, bundleID)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(application, bundleID)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetRequestParametersSecretName provides a mock function with given fields: application, bundleID
func (_m *NameResolver) GetRequestParametersSecretName(application string, bundleID string) string {
	ret := _m.Called(application, bundleID)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(application, bundleID)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
