// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	pager "github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	resource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"
)

// secretSvc is an autogenerated mock type for the secretSvc type
type secretSvc struct {
	mock.Mock
}

// Delete provides a mock function with given fields: name, namespace
func (_m *secretSvc) Delete(name string, namespace string) error {
	ret := _m.Called(name, namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(name, namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Find provides a mock function with given fields: name, namespace
func (_m *secretSvc) Find(name string, namespace string) (*v1.Secret, error) {
	ret := _m.Called(name, namespace)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(string, string) *v1.Secret); ok {
		r0 = rf(name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(name, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: namespace, params
func (_m *secretSvc) List(namespace string, params pager.PagingParams) ([]*v1.Secret, error) {
	ret := _m.Called(namespace, params)

	var r0 []*v1.Secret
	if rf, ok := ret.Get(0).(func(string, pager.PagingParams) []*v1.Secret); ok {
		r0 = rf(namespace, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, pager.PagingParams) error); ok {
		r1 = rf(namespace, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Subscribe provides a mock function with given fields: listener
func (_m *secretSvc) Subscribe(listener resource.Listener) {
	_m.Called(listener)
}

// Unsubscribe provides a mock function with given fields: listener
func (_m *secretSvc) Unsubscribe(listener resource.Listener) {
	_m.Called(listener)
}

// Update provides a mock function with given fields: name, namespace, update
func (_m *secretSvc) Update(name string, namespace string, update v1.Secret) (*v1.Secret, error) {
	ret := _m.Called(name, namespace, update)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(string, string, v1.Secret) *v1.Secret); ok {
		r0 = rf(name, namespace, update)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, v1.Secret) error); ok {
		r1 = rf(name, namespace, update)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
