// Code generated by mockery v1.0.0
package automock

import assetstore "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore"
import mock "github.com/stretchr/testify/mock"
import v1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"

// fileSvc is an autogenerated mock type for the fileSvc type
type fileSvc struct {
	mock.Mock
}

// Extract provides a mock function with given fields: statusRef
func (_m *fileSvc) Extract(statusRef *v1alpha2.AssetStatusRef) ([]*assetstore.File, error) {
	ret := _m.Called(statusRef)

	var r0 []*assetstore.File
	if rf, ok := ret.Get(0).(func(*v1alpha2.AssetStatusRef) []*assetstore.File); ok {
		r0 = rf(statusRef)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*assetstore.File)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha2.AssetStatusRef) error); ok {
		r1 = rf(statusRef)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FilterByExtensionsAndExtract provides a mock function with given fields: statusRef, filterExtensions
func (_m *fileSvc) FilterByExtensionsAndExtract(statusRef *v1alpha2.AssetStatusRef, filterExtensions []string) ([]*assetstore.File, error) {
	ret := _m.Called(statusRef, filterExtensions)

	var r0 []*assetstore.File
	if rf, ok := ret.Get(0).(func(*v1alpha2.AssetStatusRef, []string) []*assetstore.File); ok {
		r0 = rf(statusRef, filterExtensions)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*assetstore.File)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha2.AssetStatusRef, []string) error); ok {
		r1 = rf(statusRef, filterExtensions)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
