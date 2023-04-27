// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	authorization "github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	mock "github.com/stretchr/testify/mock"
)

// StrategyFactory is an autogenerated mock type for the StrategyFactory type
type StrategyFactory struct {
	mock.Mock
}

// Create provides a mock function with given fields: credentials
func (_m *StrategyFactory) Create(credentials *authorization.Credentials) authorization.Strategy {
	ret := _m.Called(credentials)

	var r0 authorization.Strategy
	if rf, ok := ret.Get(0).(func(*authorization.Credentials) authorization.Strategy); ok {
		r0 = rf(credentials)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(authorization.Strategy)
		}
	}

	return r0
}

type mockConstructorTestingTNewStrategyFactory interface {
	mock.TestingT
	Cleanup(func())
}

// NewStrategyFactory creates a new instance of StrategyFactory. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStrategyFactory(t mockConstructorTestingTNewStrategyFactory) *StrategyFactory {
	mock := &StrategyFactory{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
