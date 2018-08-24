// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"

// BrokerFlavorProvider is an autogenerated mock type for the brokerFlavorProvider type
type BrokerFlavorProvider struct {
	mock.Mock
}

// GetNsFromBrokerURL provides a mock function with given fields: url
func (_m *BrokerFlavorProvider) GetNsFromBrokerURL(url string) (string, error) {
	ret := _m.Called(url)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(url)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(url)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsClusterScoped provides a mock function with given fields:
func (_m *BrokerFlavorProvider) IsClusterScoped() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
