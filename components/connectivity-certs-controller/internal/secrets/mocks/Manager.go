// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import (
	mock "github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// Create provides a mock function with given fields: secret
func (_m *Manager) Create(secret *v1.Secret) (*v1.Secret, error) {
	ret := _m.Called(secret)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(*v1.Secret) *v1.Secret); ok {
		r0 = rf(secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.Secret) error); ok {
		r1 = rf(secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: name, options
func (_m *Manager) Delete(name string, options *metav1.DeleteOptions) error {
	ret := _m.Called(name, options)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *metav1.DeleteOptions) error); ok {
		r0 = rf(name, options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: name, options
func (_m *Manager) Get(name string, options metav1.GetOptions) (*v1.Secret, error) {
	ret := _m.Called(name, options)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(string, metav1.GetOptions) *v1.Secret); ok {
		r0 = rf(name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, metav1.GetOptions) error); ok {
		r1 = rf(name, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: secret
func (_m *Manager) Update(secret *v1.Secret) (*v1.Secret, error) {
	ret := _m.Called(secret)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(*v1.Secret) *v1.Secret); ok {
		r0 = rf(secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.Secret) error); ok {
		r1 = rf(secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
