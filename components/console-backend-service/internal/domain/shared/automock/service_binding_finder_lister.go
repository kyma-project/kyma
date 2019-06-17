// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"

import v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

// ServiceBindingFinderLister is an autogenerated mock type for the ServiceBindingFinderLister type
type ServiceBindingFinderLister struct {
	mock.Mock
}

// Find provides a mock function with given fields: ns, name
func (_m *ServiceBindingFinderLister) Find(ns string, name string) (*v1beta1.ServiceBinding, error) {
	ret := _m.Called(ns, name)

	var r0 *v1beta1.ServiceBinding
	if rf, ok := ret.Get(0).(func(string, string) *v1beta1.ServiceBinding); ok {
		r0 = rf(ns, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1beta1.ServiceBinding)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(ns, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListForServiceInstance provides a mock function with given fields: ns, instanceName
func (_m *ServiceBindingFinderLister) ListForServiceInstance(ns string, instanceName string) ([]*v1beta1.ServiceBinding, error) {
	ret := _m.Called(ns, instanceName)

	var r0 []*v1beta1.ServiceBinding
	if rf, ok := ret.Get(0).(func(string, string) []*v1beta1.ServiceBinding); ok {
		r0 = rf(ns, instanceName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1beta1.ServiceBinding)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(ns, instanceName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
