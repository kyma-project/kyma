// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"
import pager "github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"

import v1alpha1 "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"

// usageKindServices is an autogenerated mock type for the usageKindServices type
type usageKindServices struct {
	mock.Mock
}

// List provides a mock function with given fields: params
func (_m *usageKindServices) List(params pager.PagingParams) ([]*v1alpha1.UsageKind, error) {
	ret := _m.Called(params)

	var r0 []*v1alpha1.UsageKind
	if rf, ok := ret.Get(0).(func(pager.PagingParams) []*v1alpha1.UsageKind); ok {
		r0 = rf(params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1alpha1.UsageKind)
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

// ListUsageKindResources provides a mock function with given fields: usageKind, environment
func (_m *usageKindServices) ListUsageKindResources(usageKind string, environment string) ([]gqlschema.UsageKindResource, error) {
	ret := _m.Called(usageKind, environment)

	var r0 []gqlschema.UsageKindResource
	if rf, ok := ret.Get(0).(func(string, string) []gqlschema.UsageKindResource); ok {
		r0 = rf(usageKind, environment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.UsageKindResource)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(usageKind, environment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
