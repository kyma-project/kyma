// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// Specification is an autogenerated mock type for the Specification type
type Specification struct {
	mock.Mock
}

// Decode provides a mock function with given fields: data
func (_m *Specification) Decode(data []byte) error {
	ret := _m.Called(data)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
