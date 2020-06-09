// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	v1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	mock "github.com/stretchr/testify/mock"
)

// eventActivationLister is an autogenerated mock type for the eventActivationLister type
type eventActivationLister struct {
	mock.Mock
}

// List provides a mock function with given fields: namespace
func (_m *eventActivationLister) List(namespace string) ([]*v1alpha1.EventActivation, error) {
	ret := _m.Called(namespace)

	var r0 []*v1alpha1.EventActivation
	if rf, ok := ret.Get(0).(func(string) []*v1alpha1.EventActivation); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1alpha1.EventActivation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
