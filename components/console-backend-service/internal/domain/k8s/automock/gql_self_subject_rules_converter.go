// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

import mock "github.com/stretchr/testify/mock"
import v1 "k8s.io/api/authorization/v1"

// gqlSelfSubjectRulesConverter is an autogenerated mock type for the gqlSelfSubjectRulesConverter type
type gqlSelfSubjectRulesConverter struct {
	mock.Mock
}

// ToGQL provides a mock function with given fields: in
func (_m *gqlSelfSubjectRulesConverter) ToGQL(in *v1.SelfSubjectRulesReview) ([]gqlschema.ResourceRule, error) {
	ret := _m.Called(in)

	var r0 []gqlschema.ResourceRule
	if rf, ok := ret.Get(0).(func(*v1.SelfSubjectRulesReview) []gqlschema.ResourceRule); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.ResourceRule)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.SelfSubjectRulesReview) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
