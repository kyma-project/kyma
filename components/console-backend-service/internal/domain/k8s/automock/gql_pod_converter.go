// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

import mock "github.com/stretchr/testify/mock"
import v1 "k8s.io/api/core/v1"

// gqlPodConverter is an autogenerated mock type for the gqlPodConverter type
type gqlPodConverter struct {
	mock.Mock
}

// GQLJSONToPod provides a mock function with given fields: in
func (_m *gqlPodConverter) GQLJSONToPod(in gqlschema.JSON) (v1.Pod, error) {
	ret := _m.Called(in)

	var r0 v1.Pod
	if rf, ok := ret.Get(0).(func(gqlschema.JSON) v1.Pod); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(v1.Pod)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(gqlschema.JSON) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGQL provides a mock function with given fields: in
func (_m *gqlPodConverter) ToGQL(in *v1.Pod) (*gqlschema.Pod, error) {
	ret := _m.Called(in)

	var r0 *gqlschema.Pod
	if rf, ok := ret.Get(0).(func(*v1.Pod) *gqlschema.Pod); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.Pod)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.Pod) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGQLs provides a mock function with given fields: in
func (_m *gqlPodConverter) ToGQLs(in []*v1.Pod) ([]gqlschema.Pod, error) {
	ret := _m.Called(in)

	var r0 []gqlschema.Pod
	if rf, ok := ret.Get(0).(func([]*v1.Pod) []gqlschema.Pod); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.Pod)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*v1.Pod) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
