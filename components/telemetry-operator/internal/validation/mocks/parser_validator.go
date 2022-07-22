// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	v1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	mock "github.com/stretchr/testify/mock"
)

// ParserValidator is an autogenerated mock type for the ParserValidator type
type ParserValidator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: logPipeline
func (_m *ParserValidator) Validate(logPipeline *v1alpha1.LogParser) error {
	ret := _m.Called(logPipeline)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha1.LogParser) error); ok {
		r0 = rf(logPipeline)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewParserValidator interface {
	mock.TestingT
	Cleanup(func())
}

// NewParserValidator creates a new instance of ParserValidator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewParserValidator(t mockConstructorTestingTNewParserValidator) *ParserValidator {
	mock := &ParserValidator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
