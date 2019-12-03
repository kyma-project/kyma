// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import mock "github.com/stretchr/testify/mock"

import v1beta1 "github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"

// AssetGroupGetter is an autogenerated mock type for the AssetGroupGetter type
type AssetGroupGetter struct {
	mock.Mock
}

// Find provides a mock function with given fields: namespace, name
func (_m *AssetGroupGetter) Find(namespace string, name string) (*v1beta1.AssetGroup, error) {
	ret := _m.Called(namespace, name)

	var r0 *v1beta1.AssetGroup
	if rf, ok := ret.Get(0).(func(string, string) *v1beta1.AssetGroup); ok {
		r0 = rf(namespace, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1beta1.AssetGroup)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(namespace, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
