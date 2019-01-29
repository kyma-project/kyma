// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"

import v1alpha1 "github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"

// gqlBackendModuleConverter is an autogenerated mock type for the gqlBackendModuleConverter type
type gqlBackendModuleConverter struct {
	mock.Mock
}

// ToGQL provides a mock function with given fields: in
func (_m *gqlBackendModuleConverter) ToGQL(in *v1alpha1.BackendModule) (*gqlschema.BackendModule, error) {
	ret := _m.Called(in)

	var r0 *gqlschema.BackendModule
	if rf, ok := ret.Get(0).(func(*v1alpha1.BackendModule) *gqlschema.BackendModule); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.BackendModule)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha1.BackendModule) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGQLs provides a mock function with given fields: in
func (_m *gqlBackendModuleConverter) ToGQLs(in []*v1alpha1.BackendModule) ([]gqlschema.BackendModule, error) {
	ret := _m.Called(in)

	var r0 []gqlschema.BackendModule
	if rf, ok := ret.Get(0).(func([]*v1alpha1.BackendModule) []gqlschema.BackendModule); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.BackendModule)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*v1alpha1.BackendModule) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
