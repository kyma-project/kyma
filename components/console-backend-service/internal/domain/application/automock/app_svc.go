// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"
import pager "github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
import resource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
import v1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

// appSvc is an autogenerated mock type for the appSvc type
type appSvc struct {
	mock.Mock
}

// Create provides a mock function with given fields: name, description, labels
func (_m *appSvc) Create(name string, description string, labels gqlschema.Labels) (*v1alpha1.Application, error) {
	ret := _m.Called(name, description, labels)

	var r0 *v1alpha1.Application
	if rf, ok := ret.Get(0).(func(string, string, gqlschema.Labels) *v1alpha1.Application); ok {
		r0 = rf(name, description, labels)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, gqlschema.Labels) error); ok {
		r1 = rf(name, description, labels)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: name
func (_m *appSvc) Delete(name string) error {
	ret := _m.Called(name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Disable provides a mock function with given fields: namespace, name
func (_m *appSvc) Disable(namespace string, name string) error {
	ret := _m.Called(namespace, name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(namespace, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Enable provides a mock function with given fields: namespace, name, services
func (_m *appSvc) Enable(namespace string, name string, services []*gqlschema.ApplicationMappingService) (*applicationconnectorv1alpha1.ApplicationMapping, error) {
	ret := _m.Called(namespace, name, services)

	var r0 *applicationconnectorv1alpha1.ApplicationMapping
	if rf, ok := ret.Get(0).(func(string, string, []*gqlschema.ApplicationMappingService) *applicationconnectorv1alpha1.ApplicationMapping); ok {
		r0 = rf(namespace, name, services)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*applicationconnectorv1alpha1.ApplicationMapping)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, []*gqlschema.ApplicationMappingService) error); ok {
		r1 = rf(namespace, name, services)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Find provides a mock function with given fields: name
func (_m *appSvc) Find(name string) (*v1alpha1.Application, error) {
	ret := _m.Called(name)

	var r0 *v1alpha1.Application
	if rf, ok := ret.Get(0).(func(string) *v1alpha1.Application); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Application)
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

// GetConnectionURL provides a mock function with given fields: _a0
func (_m *appSvc) GetConnectionURL(_a0 string) (string, error) {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: params
func (_m *appSvc) List(params pager.PagingParams) ([]*v1alpha1.Application, error) {
	ret := _m.Called(params)

	var r0 []*v1alpha1.Application
	if rf, ok := ret.Get(0).(func(pager.PagingParams) []*v1alpha1.Application); ok {
		r0 = rf(params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1alpha1.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(pager.PagingParams) error); ok {
		r1 = rf(params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListApplicationMapping provides a mock function with given fields: name
func (_m *appSvc) ListApplicationMapping(name string) ([]*applicationconnectorv1alpha1.ApplicationMapping, error) {
	ret := _m.Called(name)

	var r0 []*applicationconnectorv1alpha1.ApplicationMapping
	if rf, ok := ret.Get(0).(func(string) []*applicationconnectorv1alpha1.ApplicationMapping); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*applicationconnectorv1alpha1.ApplicationMapping)
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

// ListInNamespace provides a mock function with given fields: namespace
func (_m *appSvc) ListInNamespace(namespace string) ([]*v1alpha1.Application, error) {
	ret := _m.Called(namespace)

	var r0 []*v1alpha1.Application
	if rf, ok := ret.Get(0).(func(string) []*v1alpha1.Application); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1alpha1.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListNamespacesFor provides a mock function with given fields: appName
func (_m *appSvc) ListNamespacesFor(appName string) ([]string, error) {
	ret := _m.Called(appName)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(appName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(appName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Subscribe provides a mock function with given fields: listener
func (_m *appSvc) Subscribe(listener resource.Listener) {
	_m.Called(listener)
}

// Unsubscribe provides a mock function with given fields: listener
func (_m *appSvc) Unsubscribe(listener resource.Listener) {
	_m.Called(listener)
}

// Update provides a mock function with given fields: name, description, labels
func (_m *appSvc) Update(name string, description string, labels gqlschema.Labels) (*v1alpha1.Application, error) {
	ret := _m.Called(name, description, labels)

	var r0 *v1alpha1.Application
	if rf, ok := ret.Get(0).(func(string, string, gqlschema.Labels) *v1alpha1.Application); ok {
		r0 = rf(name, description, labels)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, gqlschema.Labels) error); ok {
		r1 = rf(name, description, labels)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateApplicationMapping provides a mock function with given fields: namespace, name, services
func (_m *appSvc) UpdateApplicationMapping(namespace string, name string, services []*gqlschema.ApplicationMappingService) (*applicationconnectorv1alpha1.ApplicationMapping, error) {
	ret := _m.Called(namespace, name, services)

	var r0 *applicationconnectorv1alpha1.ApplicationMapping
	if rf, ok := ret.Get(0).(func(string, string, []*gqlschema.ApplicationMappingService) *applicationconnectorv1alpha1.ApplicationMapping); ok {
		r0 = rf(namespace, name, services)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*applicationconnectorv1alpha1.ApplicationMapping)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, []*gqlschema.ApplicationMappingService) error); ok {
		r1 = rf(namespace, name, services)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
