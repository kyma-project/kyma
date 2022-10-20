// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// commandRunner is an autogenerated mock type for the commandRunner type
type commandRunner struct {
	mock.Mock
}

// run provides a mock function with given fields: ctx, command, args
func (_m *commandRunner) run(ctx context.Context, command string, args ...string) ([]byte, error) {
	_va := make([]interface{}, len(args))
	for _i := range args {
		_va[_i] = args[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, command)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, string, ...string) []byte); ok {
		r0 = rf(ctx, command, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, ...string) error); ok {
		r1 = rf(ctx, command, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTnewCommandRunner interface {
	mock.TestingT
	Cleanup(func())
}

// newCommandRunner creates a new instance of commandRunner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newCommandRunner(t mockConstructorTestingTnewCommandRunner) *commandRunner {
	mock := &commandRunner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
