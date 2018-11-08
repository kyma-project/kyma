// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"

// serviceNameProvider is an autogenerated mock type for the serviceNameProvider type
type ServiceNameProvider struct {
	mock.Mock
}

// GetServiceNameForNsBroker provides a mock function with given fields: ns
func (_m *ServiceNameProvider) GetServiceNameForNsBroker(ns string) string {
	ret := _m.Called(ns)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(ns)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
