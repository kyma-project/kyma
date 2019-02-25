// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"

// Accessor is an autogenerated mock type for the Accessor type
type Accessor struct {
	mock.Mock
}

// GetName provides a mock function with given fields:
func (_m *Accessor) GetName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetNamespace provides a mock function with given fields:
func (_m *Accessor) GetNamespace() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
