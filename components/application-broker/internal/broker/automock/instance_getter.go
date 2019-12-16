// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import internal "github.com/kyma-project/kyma/components/application-broker/internal"
import mock "github.com/stretchr/testify/mock"

// instanceGetter is an autogenerated mock type for the instanceGetter type
type instanceGetter struct {
	mock.Mock
}

// Get provides a mock function with given fields: id
func (_m *instanceGetter) Get(id internal.InstanceID) (*internal.Instance, error) {
	ret := _m.Called(id)

	var r0 *internal.Instance
	if rf, ok := ret.Get(0).(func(internal.InstanceID) *internal.Instance); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internal.Instance)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(internal.InstanceID) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
