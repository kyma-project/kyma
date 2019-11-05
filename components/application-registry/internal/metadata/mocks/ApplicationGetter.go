// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

// ApplicationGetter is an autogenerated mock type for the ApplicationGetter type
type ApplicationGetter struct {
	mock.Mock
}

// Get provides a mock function with given fields: name, options
func (_m *ApplicationGetter) Get(name string, options v1.GetOptions) (*v1alpha1.Application, error) {
	ret := _m.Called(name, options)

	var r0 *v1alpha1.Application
	if rf, ok := ret.Get(0).(func(string, v1.GetOptions) *v1alpha1.Application); ok {
		r0 = rf(name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Application)
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
