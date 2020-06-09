// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
)

// apiRuleConv is an autogenerated mock type for the apiRuleConv type
type apiRuleConv struct {
	mock.Mock
}

// ToApiRule provides a mock function with given fields: name, namespace, in
func (_m *apiRuleConv) ToApiRule(name string, namespace string, in gqlschema.APIRuleInput) *v1alpha1.APIRule {
	ret := _m.Called(name, namespace, in)

	var r0 *v1alpha1.APIRule
	if rf, ok := ret.Get(0).(func(string, string, gqlschema.APIRuleInput) *v1alpha1.APIRule); ok {
		r0 = rf(name, namespace, in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.APIRule)
		}
	}

	return r0
}

// ToGQL provides a mock function with given fields: in
func (_m *apiRuleConv) ToGQL(in *v1alpha1.APIRule) (*gqlschema.APIRule, error) {
	ret := _m.Called(in)

	var r0 *gqlschema.APIRule
	if rf, ok := ret.Get(0).(func(*v1alpha1.APIRule) *gqlschema.APIRule); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.APIRule)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha1.APIRule) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGQLs provides a mock function with given fields: in
func (_m *apiRuleConv) ToGQLs(in []*v1alpha1.APIRule) ([]*gqlschema.APIRule, error) {
	ret := _m.Called(in)

	var r0 []*gqlschema.APIRule
	if rf, ok := ret.Get(0).(func([]*v1alpha1.APIRule) []*gqlschema.APIRule); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gqlschema.APIRule)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*v1alpha1.APIRule) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
