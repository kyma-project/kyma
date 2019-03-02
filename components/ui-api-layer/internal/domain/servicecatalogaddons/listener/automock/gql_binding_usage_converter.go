// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import gqlschema "github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"

import mock "github.com/stretchr/testify/mock"
import v1alpha1 "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"

// gqlBindingUsageConverter is an autogenerated mock type for the gqlBindingUsageConverter type
type gqlBindingUsageConverter struct {
	mock.Mock
}

// ToGQL provides a mock function with given fields: in
func (_m *gqlBindingUsageConverter) ToGQL(in *v1alpha1.ServiceBindingUsage) (*gqlschema.ServiceBindingUsage, error) {
	ret := _m.Called(in)

	var r0 *gqlschema.ServiceBindingUsage
	if rf, ok := ret.Get(0).(func(*v1alpha1.ServiceBindingUsage) *gqlschema.ServiceBindingUsage); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.ServiceBindingUsage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha1.ServiceBindingUsage) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
