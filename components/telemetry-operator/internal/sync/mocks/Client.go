// Code generated by mockery v2.12.1. DO NOT EDIT.

package mocks

import (
	context "context"

	client "sigs.k8s.io/controller-runtime/pkg/client"

	meta "k8s.io/apimachinery/pkg/api/meta"

	mock "github.com/stretchr/testify/mock"

	runtime "k8s.io/apimachinery/pkg/runtime"

	testing "testing"

	types "k8s.io/apimachinery/pkg/types"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, obj, opts
func (_m *Client) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, obj)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, client.Object, ...client.CreateOption) error); ok {
		r0 = rf(ctx, obj, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, obj, opts
func (_m *Client) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, obj)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, client.Object, ...client.DeleteOption) error); ok {
		r0 = rf(ctx, obj, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAllOf provides a mock function with given fields: ctx, obj, opts
func (_m *Client) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, obj)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, client.Object, ...client.DeleteAllOfOption) error); ok {
		r0 = rf(ctx, obj, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, key, obj
func (_m *Client) Get(ctx context.Context, key types.NamespacedName, obj client.Object) error {
	ret := _m.Called(ctx, key, obj)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.NamespacedName, client.Object) error); ok {
		r0 = rf(ctx, key, obj)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// List provides a mock function with given fields: ctx, list, opts
func (_m *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, list)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, client.ObjectList, ...client.ListOption) error); ok {
		r0 = rf(ctx, list, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Patch provides a mock function with given fields: ctx, obj, patch, opts
func (_m *Client) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, obj, patch)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, client.Object, client.Patch, ...client.PatchOption) error); ok {
		r0 = rf(ctx, obj, patch, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RESTMapper provides a mock function with given fields:
func (_m *Client) RESTMapper() meta.RESTMapper {
	ret := _m.Called()

	var r0 meta.RESTMapper
	if rf, ok := ret.Get(0).(func() meta.RESTMapper); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(meta.RESTMapper)
		}
	}

	return r0
}

// Scheme provides a mock function with given fields:
func (_m *Client) Scheme() *runtime.Scheme {
	ret := _m.Called()

	var r0 *runtime.Scheme
	if rf, ok := ret.Get(0).(func() *runtime.Scheme); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*runtime.Scheme)
		}
	}

	return r0
}

// Status provides a mock function with given fields:
func (_m *Client) Status() client.StatusWriter {
	ret := _m.Called()

	var r0 client.StatusWriter
	if rf, ok := ret.Get(0).(func() client.StatusWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.StatusWriter)
		}
	}

	return r0
}

// Update provides a mock function with given fields: ctx, obj, opts
func (_m *Client) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, obj)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, client.Object, ...client.UpdateOption) error); ok {
		r0 = rf(ctx, obj, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewClient creates a new instance of Client. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewClient(t testing.TB) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
