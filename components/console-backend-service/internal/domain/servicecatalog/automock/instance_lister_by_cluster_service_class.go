// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"

import v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

// instanceListerByClusterServiceClass is an autogenerated mock type for the instanceListerByClusterServiceClass type
type instanceListerByClusterServiceClass struct {
	mock.Mock
}

// ListForClusterServiceClass provides a mock function with given fields: className, externalClassName, namespace
func (_m *instanceListerByClusterServiceClass) ListForClusterServiceClass(className string, externalClassName string, namespace *string) ([]*v1beta1.ServiceInstance, error) {
	ret := _m.Called(className, externalClassName, namespace)

	var r0 []*v1beta1.ServiceInstance
	if rf, ok := ret.Get(0).(func(string, string, *string) []*v1beta1.ServiceInstance); ok {
		r0 = rf(className, externalClassName, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1beta1.ServiceInstance)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *string) error); ok {
		r1 = rf(className, externalClassName, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
