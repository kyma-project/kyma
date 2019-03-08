// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import mock "github.com/stretchr/testify/mock"

import v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

// servicePlanLister is an autogenerated mock type for the servicePlanLister type
type servicePlanLister struct {
	mock.Mock
}

// ListForServiceClass provides a mock function with given fields: name, namespace
func (_m *servicePlanLister) ListForServiceClass(name string, namespace string) ([]*v1beta1.ServicePlan, error) {
	ret := _m.Called(name, namespace)

	var r0 []*v1beta1.ServicePlan
	if rf, ok := ret.Get(0).(func(string, string) []*v1beta1.ServicePlan); ok {
		r0 = rf(name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1beta1.ServicePlan)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(name, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
