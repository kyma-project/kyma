// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"

// KubernetesResourceSupervisor is an autogenerated mock type for the KubernetesResourceSupervisor type
type KubernetesResourceSupervisor struct {
	mock.Mock
}

// EnsureLabelsCreated provides a mock function with given fields: resourceNs, resourceName, usageName, usageVer, labels
func (_m *KubernetesResourceSupervisor) EnsureLabelsCreated(resourceNs string, resourceName string, usageName string, labels map[string]string) error {
	ret := _m.Called(resourceNs, resourceName, usageName, labels)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, map[string]string) error); ok {
		r0 = rf(resourceNs, resourceName, usageName, labels)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnsureLabelsDeleted provides a mock function with given fields: resourceNs, resourceName, usageName
func (_m *KubernetesResourceSupervisor) EnsureLabelsDeleted(resourceNs string, resourceName string, usageName string) error {
	ret := _m.Called(resourceNs, resourceName, usageName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(resourceNs, resourceName, usageName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetInjectedLabels provides a mock function with given fields: resourceNs, resourceName, usageName
func (_m *KubernetesResourceSupervisor) GetInjectedLabels(resourceNs string, resourceName string, usageName string) (map[string]string, error) {
	ret := _m.Called(resourceNs, resourceName, usageName)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(string, string, string) map[string]string); ok {
		r0 = rf(resourceNs, resourceName, usageName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(resourceNs, resourceName, usageName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

