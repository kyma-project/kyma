// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"

import v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

// clusterServicePlanLister is an autogenerated mock type for the clusterServicePlanLister type
type clusterServicePlanLister struct {
	mock.Mock
}

// ListForClusterServiceClass provides a mock function with given fields: name
func (_m *clusterServicePlanLister) ListForClusterServiceClass(name string) ([]*v1beta1.ClusterServicePlan, error) {
	ret := _m.Called(name)

	var r0 []*v1beta1.ClusterServicePlan
	if rf, ok := ret.Get(0).(func(string) []*v1beta1.ClusterServicePlan); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1beta1.ClusterServicePlan)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
