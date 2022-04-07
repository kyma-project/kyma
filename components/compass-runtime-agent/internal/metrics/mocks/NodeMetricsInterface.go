// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	watch "k8s.io/apimachinery/pkg/watch"
)

// NodeMetricsInterface is an autogenerated mock type for the NodeMetricsInterface type
type NodeMetricsInterface struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, name, opts
func (_m *NodeMetricsInterface) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.NodeMetrics, error) {
	ret := _m.Called(ctx, name, opts)

	var r0 *v1beta1.NodeMetrics
	if rf, ok := ret.Get(0).(func(context.Context, string, v1.GetOptions) *v1beta1.NodeMetrics); ok {
		r0 = rf(ctx, name, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1beta1.NodeMetrics)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, v1.GetOptions) error); ok {
		r1 = rf(ctx, name, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, opts
func (_m *NodeMetricsInterface) List(ctx context.Context, opts v1.ListOptions) (*v1beta1.NodeMetricsList, error) {
	ret := _m.Called(ctx, opts)

	var r0 *v1beta1.NodeMetricsList
	if rf, ok := ret.Get(0).(func(context.Context, v1.ListOptions) *v1beta1.NodeMetricsList); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1beta1.NodeMetricsList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, v1.ListOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Watch provides a mock function with given fields: ctx, opts
func (_m *NodeMetricsInterface) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	ret := _m.Called(ctx, opts)

	var r0 watch.Interface
	if rf, ok := ret.Get(0).(func(context.Context, v1.ListOptions) watch.Interface); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(watch.Interface)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, v1.ListOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
