// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"
import pager "github.com/kyma-project/kyma/components/console-backend-service/internal/pager"

import v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

// serviceClassListGetter is an autogenerated mock type for the serviceClassListGetter type
type serviceClassListGetter struct {
	mock.Mock
}

// Find provides a mock function with given fields: name, namespace
func (_m *serviceClassListGetter) Find(name string, namespace string) (*v1beta1.ServiceClass, error) {
	ret := _m.Called(name, namespace)

	var r0 *v1beta1.ServiceClass
	if rf, ok := ret.Get(0).(func(string, string) *v1beta1.ServiceClass); ok {
		r0 = rf(name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1beta1.ServiceClass)
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

// FindByExternalName provides a mock function with given fields: externalName, namespace
func (_m *serviceClassListGetter) FindByExternalName(externalName string, namespace string) (*v1beta1.ServiceClass, error) {
	ret := _m.Called(externalName, namespace)

	var r0 *v1beta1.ServiceClass
	if rf, ok := ret.Get(0).(func(string, string) *v1beta1.ServiceClass); ok {
		r0 = rf(externalName, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1beta1.ServiceClass)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(externalName, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: namespace, pagingParams
func (_m *serviceClassListGetter) List(namespace string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceClass, error) {
	ret := _m.Called(namespace, pagingParams)

	var r0 []*v1beta1.ServiceClass
	if rf, ok := ret.Get(0).(func(string, pager.PagingParams) []*v1beta1.ServiceClass); ok {
		r0 = rf(namespace, pagingParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1beta1.ServiceClass)
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
