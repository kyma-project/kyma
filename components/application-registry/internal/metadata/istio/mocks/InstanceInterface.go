// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha2 "github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"
)

// InstanceInterface is an autogenerated mock type for the InstanceInterface type
type InstanceInterface struct {
	mock.Mock
}

// Create provides a mock function with given fields: _a0, _a1, _a2
func (_m *InstanceInterface) Create(_a0 context.Context, _a1 *v1alpha2.Instance, _a2 v1.CreateOptions) (*v1alpha2.Instance, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 *v1alpha2.Instance
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha2.Instance, v1.CreateOptions) *v1alpha2.Instance); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha2.Instance)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *v1alpha2.Instance, v1.CreateOptions) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: _a0, _a1, _a2
func (_m *InstanceInterface) Delete(_a0 context.Context, _a1 string, _a2 v1.DeleteOptions) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, v1.DeleteOptions) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
