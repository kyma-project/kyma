// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"
import v1beta2 "k8s.io/api/apps/v1beta2"

// deploymentGetter is an autogenerated mock type for the deploymentGetter type
type deploymentGetter struct {
	mock.Mock
}

// Find provides a mock function with given fields: name, environment
func (_m *deploymentGetter) Find(name string, environment string) (*v1beta2.Deployment, error) {
	ret := _m.Called(name, environment)

	var r0 *v1beta2.Deployment
	if rf, ok := ret.Get(0).(func(string, string) *v1beta2.Deployment); ok {
		r0 = rf(name, environment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1beta2.Deployment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(name, environment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
