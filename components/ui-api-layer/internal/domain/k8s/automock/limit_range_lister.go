// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"
import v1 "k8s.io/api/core/v1"

// limitRangeLister is an autogenerated mock type for the limitRangeLister type
type limitRangeLister struct {
	mock.Mock
}

// List provides a mock function with given fields: env
func (_m *limitRangeLister) List(env string) ([]*v1.LimitRange, error) {
	ret := _m.Called(env)

	var r0 []*v1.LimitRange
	if rf, ok := ret.Get(0).(func(string) []*v1.LimitRange); ok {
		r0 = rf(env)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1.LimitRange)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(env)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
