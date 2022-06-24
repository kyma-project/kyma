// Code generated by mockery v2.13.. DO NOT EDIT.

package mocks

import (
	v1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	mock "github.com/stretchr/testify/mock"
)

// PluginValidator is an autogenerated mock type for the PluginValidator type
type PluginValidator struct {
	mock.Mock
}

// ContainsCustomPlugin provides a mock function with given fields: logPipeline
func (_m *PluginValidator) ContainsCustomPlugin(logPipeline *v1alpha1.LogPipeline) bool {
	ret := _m.Called(logPipeline)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*v1alpha1.LogPipeline) bool); ok {
		r0 = rf(logPipeline)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Validate provides a mock function with given fields: logPipeline, logPipelines
func (_m *PluginValidator) Validate(logPipeline *v1alpha1.LogPipeline, logPipelines *v1alpha1.LogPipelineList) error {
	ret := _m.Called(logPipeline, logPipelines)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha1.LogPipeline, *v1alpha1.LogPipelineList) error); ok {
		r0 = rf(logPipeline, logPipelines)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewPluginValidator interface {
	mock.TestingT
	Cleanup(func())
}

// NewPluginValidator creates a new instance of PluginValidator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPluginValidator(t mockConstructorTestingTNewPluginValidator) *PluginValidator {
	mock := &PluginValidator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
