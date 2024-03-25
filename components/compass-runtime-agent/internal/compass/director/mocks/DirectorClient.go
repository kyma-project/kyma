// Code generated by mockery v2.30.1. DO NOT EDIT.

package mocks

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	director "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/director"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

// DirectorClient is an autogenerated mock type for the DirectorClient type
type DirectorClient struct {
	mock.Mock
}

// FetchConfiguration provides a mock function with given fields: ctx
func (_m *DirectorClient) FetchConfiguration(ctx context.Context) ([]model.Application, graphql.Labels, error) {
	ret := _m.Called(ctx)

	var r0 []model.Application
	var r1 graphql.Labels
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]model.Application, graphql.Labels, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []model.Application); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Application)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) graphql.Labels); ok {
		r1 = rf(ctx)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(graphql.Labels)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context) error); ok {
		r2 = rf(ctx)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetRuntime provides a mock function with given fields: ctx
func (_m *DirectorClient) GetRuntime(ctx context.Context) (graphql.RuntimeExt, error) {
	ret := _m.Called(ctx)

	var r0 graphql.RuntimeExt
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (graphql.RuntimeExt, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) graphql.RuntimeExt); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(graphql.RuntimeExt)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetRuntimeStatusCondition provides a mock function with given fields: ctx, statusCondition
func (_m *DirectorClient) SetRuntimeStatusCondition(ctx context.Context, statusCondition graphql.RuntimeStatusCondition) error {
	ret := _m.Called(ctx, statusCondition)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, graphql.RuntimeStatusCondition) error); ok {
		r0 = rf(ctx, statusCondition)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetURLsLabels provides a mock function with given fields: ctx, urlsCfg, actualLabels
func (_m *DirectorClient) SetURLsLabels(ctx context.Context, urlsCfg director.RuntimeURLsConfig, actualLabels graphql.Labels) (graphql.Labels, error) {
	ret := _m.Called(ctx, urlsCfg, actualLabels)

	var r0 graphql.Labels
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, director.RuntimeURLsConfig, graphql.Labels) (graphql.Labels, error)); ok {
		return rf(ctx, urlsCfg, actualLabels)
	}
	if rf, ok := ret.Get(0).(func(context.Context, director.RuntimeURLsConfig, graphql.Labels) graphql.Labels); ok {
		r0 = rf(ctx, urlsCfg, actualLabels)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(graphql.Labels)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, director.RuntimeURLsConfig, graphql.Labels) error); ok {
		r1 = rf(ctx, urlsCfg, actualLabels)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewDirectorClient creates a new instance of DirectorClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDirectorClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *DirectorClient {
	mock := &DirectorClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
