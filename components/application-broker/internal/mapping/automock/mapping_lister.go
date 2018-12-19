// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"
import v1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"

// MappingLister is an autogenerated mock type for the MappingLister type
type MappingLister struct {
	mock.Mock
}

// ListApplicationMappings provides a mock function with given fields: application
func (_m *MappingLister) ListApplicationMappings(application string) ([]*v1alpha1.ApplicationMapping, error) {
	ret := _m.Called(application)

	var r0 []*v1alpha1.ApplicationMapping
	if rf, ok := ret.Get(0).(func(string) []*v1alpha1.ApplicationMapping); ok {
		r0 = rf(application)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1alpha1.ApplicationMapping)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(application)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
