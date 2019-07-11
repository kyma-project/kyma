// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"

import v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

// gqlServicePlanConverter is an autogenerated mock type for the gqlServicePlanConverter type
type gqlServicePlanConverter struct {
	mock.Mock
}

// ToGQL provides a mock function with given fields: item
func (_m *gqlServicePlanConverter) ToGQL(item *v1beta1.ServicePlan) (*gqlschema.ServicePlan, error) {
	ret := _m.Called(item)

	var r0 *gqlschema.ServicePlan
	if rf, ok := ret.Get(0).(func(*v1beta1.ServicePlan) *gqlschema.ServicePlan); ok {
		r0 = rf(item)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.ServicePlan)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1beta1.ServicePlan) error); ok {
		r1 = rf(item)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGQLs provides a mock function with given fields: in
func (_m *gqlServicePlanConverter) ToGQLs(in []*v1beta1.ServicePlan) ([]gqlschema.ServicePlan, error) {
	ret := _m.Called(in)

	var r0 []gqlschema.ServicePlan
	if rf, ok := ret.Get(0).(func([]*v1beta1.ServicePlan) []gqlschema.ServicePlan); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.ServicePlan)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*v1beta1.ServicePlan) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
