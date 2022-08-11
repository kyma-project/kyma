// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

// DryRunner is an autogenerated mock type for the DryRunner type
type DryRunner struct {
	mock.Mock
}

// DryRunPipeline provides a mock function with given fields: ctx, pipeline
func (_m *DryRunner) RunPipeline(ctx context.Context, pipeline *v1alpha1.LogPipeline) error {
	ret := _m.Called(ctx, pipeline)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.LogPipeline) error); ok {
		r0 = rf(ctx, pipeline)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewDryRunner interface {
	mock.TestingT
	Cleanup(func())
}

// NewDryRunner creates a new instance of DryRunner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDryRunner(t mockConstructorTestingTNewDryRunner) *DryRunner {
	mock := &DryRunner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
