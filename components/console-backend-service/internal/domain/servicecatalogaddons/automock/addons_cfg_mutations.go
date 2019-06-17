// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"

import v1 "k8s.io/api/core/v1"

// addonsCfgMutations is an autogenerated mock type for the addonsCfgMutations type
type addonsCfgMutations struct {
	mock.Mock
}

// Create provides a mock function with given fields: name, urls, labels
func (_m *addonsCfgMutations) Create(name string, urls []string, labels *gqlschema.Labels) (*v1.ConfigMap, error) {
	ret := _m.Called(name, urls, labels)

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func(string, []string, *gqlschema.Labels) *v1.ConfigMap); ok {
		r0 = rf(name, urls, labels)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []string, *gqlschema.Labels) error); ok {
		r1 = rf(name, urls, labels)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: name
func (_m *addonsCfgMutations) Delete(name string) (*v1.ConfigMap, error) {
	ret := _m.Called(name)

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func(string) *v1.ConfigMap); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: name, urls, labels
func (_m *addonsCfgMutations) Update(name string, urls []string, labels *gqlschema.Labels) (*v1.ConfigMap, error) {
	ret := _m.Called(name, urls, labels)

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func(string, []string, *gqlschema.Labels) *v1.ConfigMap); ok {
		r0 = rf(name, urls, labels)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []string, *gqlschema.Labels) error); ok {
		r1 = rf(name, urls, labels)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
