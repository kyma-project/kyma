// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
)

// CRManager is an autogenerated mock type for the CRManager type
type CRManager struct {
	mock.Mock
}

// Create provides a mock function with given fields: _a0
func (_m *CRManager) Create(_a0 *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error) {
	ret := _m.Called(_a0)

	var r0 *v1alpha1.CompassConnection
	if rf, ok := ret.Get(0).(func(*v1alpha1.CompassConnection) *v1alpha1.CompassConnection); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.CompassConnection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha1.CompassConnection) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: name, options
func (_m *CRManager) Delete(name string, options *v1.DeleteOptions) error {
	ret := _m.Called(name, options)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *v1.DeleteOptions) error); ok {
		r0 = rf(name, options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: name, options
func (_m *CRManager) Get(name string, options v1.GetOptions) (*v1alpha1.CompassConnection, error) {
	ret := _m.Called(name, options)

	var r0 *v1alpha1.CompassConnection
	if rf, ok := ret.Get(0).(func(string, v1.GetOptions) *v1alpha1.CompassConnection); ok {
		r0 = rf(name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.CompassConnection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, v1.GetOptions) error); ok {
		r1 = rf(name, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: _a0
func (_m *CRManager) Update(_a0 *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error) {
	ret := _m.Called(_a0)

	var r0 *v1alpha1.CompassConnection
	if rf, ok := ret.Get(0).(func(*v1alpha1.CompassConnection) *v1alpha1.CompassConnection); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.CompassConnection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha1.CompassConnection) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
