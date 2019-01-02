// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import corev1 "k8s.io/api/core/v1"
import mock "github.com/stretchr/testify/mock"

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// Get provides a mock function with given fields: name, options
func (_m *Manager) Get(name string, options v1.GetOptions) (*corev1.Secret, error) {
	ret := _m.Called(name, options)

	var r0 *corev1.Secret
	if rf, ok := ret.Get(0).(func(string, v1.GetOptions) *corev1.Secret); ok {
		r0 = rf(name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.Secret)
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
