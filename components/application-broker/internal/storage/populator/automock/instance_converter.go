// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	internal "github.com/kyma-project/kyma/components/application-broker/internal"
	mock "github.com/stretchr/testify/mock"

	v1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

// instanceConverter is an autogenerated mock type for the instanceConverter type
type instanceConverter struct {
	mock.Mock
}

// MapServiceInstance provides a mock function with given fields: in
func (_m *instanceConverter) MapServiceInstance(in *v1beta1.ServiceInstance) *internal.Instance {
	ret := _m.Called(in)

	var r0 *internal.Instance
	if rf, ok := ret.Get(0).(func(*v1beta1.ServiceInstance) *internal.Instance); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internal.Instance)
		}
	}

	return r0
}
