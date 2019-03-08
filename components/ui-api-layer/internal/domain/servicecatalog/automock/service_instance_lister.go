// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import mock "github.com/stretchr/testify/mock"
import pager "github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"

import status "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
import v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

// serviceInstanceLister is an autogenerated mock type for the serviceInstanceLister type
type serviceInstanceLister struct {
	mock.Mock
}

// Find provides a mock function with given fields: name, namespace
func (_m *serviceInstanceLister) Find(name string, namespace string) (*v1beta1.ServiceInstance, error) {
	ret := _m.Called(name, namespace)

	var r0 *v1beta1.ServiceInstance
	if rf, ok := ret.Get(0).(func(string, string) *v1beta1.ServiceInstance); ok {
		r0 = rf(name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1beta1.ServiceInstance)
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

// List provides a mock function with given fields: namespace, pagingParams
func (_m *serviceInstanceLister) List(namespace string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceInstance, error) {
	ret := _m.Called(namespace, pagingParams)

	var r0 []*v1beta1.ServiceInstance
	if rf, ok := ret.Get(0).(func(string, pager.PagingParams) []*v1beta1.ServiceInstance); ok {
		r0 = rf(namespace, pagingParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1beta1.ServiceInstance)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, pager.PagingParams) error); ok {
		r1 = rf(namespace, pagingParams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListForStatus provides a mock function with given fields: namespace, pagingParams, _a2
func (_m *serviceInstanceLister) ListForStatus(namespace string, pagingParams pager.PagingParams, _a2 *status.ServiceInstanceStatusType) ([]*v1beta1.ServiceInstance, error) {
	ret := _m.Called(namespace, pagingParams, _a2)

	var r0 []*v1beta1.ServiceInstance
	if rf, ok := ret.Get(0).(func(string, pager.PagingParams, *status.ServiceInstanceStatusType) []*v1beta1.ServiceInstance); ok {
		r0 = rf(namespace, pagingParams, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1beta1.ServiceInstance)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, pager.PagingParams, *status.ServiceInstanceStatusType) error); ok {
		r1 = rf(namespace, pagingParams, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
