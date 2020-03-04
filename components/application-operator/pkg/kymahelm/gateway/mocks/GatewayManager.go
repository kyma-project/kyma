// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import release "k8s.io/helm/pkg/proto/hapi/release"

// GatewayManager is an autogenerated mock type for the GatewayManager type
type GatewayManager struct {
	mock.Mock
}

// DeleteGateway provides a mock function with given fields: namespace
func (_m *GatewayManager) DeleteGateway(namespace string) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GatewayExists provides a mock function with given fields: namespace
func (_m *GatewayManager) GatewayExists(namespace string) (bool, release.Status_Code, error) {
	ret := _m.Called(namespace)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 release.Status_Code
	if rf, ok := ret.Get(1).(func(string) release.Status_Code); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Get(1).(release.Status_Code)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(namespace)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// InstallGateway provides a mock function with given fields: namespace
func (_m *GatewayManager) InstallGateway(namespace string) error {
	ret := _m.Called(namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpgradeGateways provides a mock function with given fields:
func (_m *GatewayManager) UpgradeGateways() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
