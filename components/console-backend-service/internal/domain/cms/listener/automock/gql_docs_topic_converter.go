// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

import mock "github.com/stretchr/testify/mock"
import v1alpha1 "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"

// gqlDocsTopicConverter is an autogenerated mock type for the gqlDocsTopicConverter type
type gqlDocsTopicConverter struct {
	mock.Mock
}

// ToGQL provides a mock function with given fields: in
func (_m *gqlDocsTopicConverter) ToGQL(in *v1alpha1.DocsTopic) (*gqlschema.DocsTopic, error) {
	ret := _m.Called(in)

	var r0 *gqlschema.DocsTopic
	if rf, ok := ret.Get(0).(func(*v1alpha1.DocsTopic) *gqlschema.DocsTopic); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.DocsTopic)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha1.DocsTopic) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
