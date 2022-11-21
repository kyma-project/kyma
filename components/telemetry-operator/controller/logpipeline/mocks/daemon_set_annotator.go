// Code generated by mockery v2.14.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "k8s.io/apimachinery/pkg/types"
)

// DaemonSetAnnotator is an autogenerated mock type for the DaemonSetAnnotator type
type DaemonSetAnnotator struct {
	mock.Mock
}

// SetAnnotation provides a mock function with given fields: ctx, name, key, value
func (_m *DaemonSetAnnotator) SetAnnotation(ctx context.Context, name types.NamespacedName, key string, value string) (bool, error) {
	ret := _m.Called(ctx, name, key, value)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, types.NamespacedName, string, string) bool); ok {
		r0 = rf(ctx, name, key, value)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.NamespacedName, string, string) error); ok {
		r1 = rf(ctx, name, key, value)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewDaemonSetAnnotator interface {
	mock.TestingT
	Cleanup(func())
}

// NewDaemonSetAnnotator creates a new instance of DaemonSetAnnotator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDaemonSetAnnotator(t mockConstructorTestingTNewDaemonSetAnnotator) *DaemonSetAnnotator {
	mock := &DaemonSetAnnotator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
