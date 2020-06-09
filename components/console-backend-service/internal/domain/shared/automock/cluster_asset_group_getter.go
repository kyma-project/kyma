// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	mock "github.com/stretchr/testify/mock"

	v1beta1 "github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

// ClusterAssetGroupGetter is an autogenerated mock type for the ClusterAssetGroupGetter type
type ClusterAssetGroupGetter struct {
	mock.Mock
}

// Find provides a mock function with given fields: name
func (_m *ClusterAssetGroupGetter) Find(name string) (*v1beta1.ClusterAssetGroup, error) {
	ret := _m.Called(name)

	var r0 *v1beta1.ClusterAssetGroup
	if rf, ok := ret.Get(0).(func(string) *v1beta1.ClusterAssetGroup); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1beta1.ClusterAssetGroup)
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
