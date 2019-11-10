// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import mock "github.com/stretchr/testify/mock"
import shared "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

// RafterRetriever is an autogenerated mock type for the RafterRetriever type
type RafterRetriever struct {
	mock.Mock
}

// AssetGroup provides a mock function with given fields:
func (_m *RafterRetriever) AssetGroup() shared.AssetGroupGetter {
	ret := _m.Called()

	var r0 shared.AssetGroupGetter
	if rf, ok := ret.Get(0).(func() shared.AssetGroupGetter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(shared.AssetGroupGetter)
		}
	}

	return r0
}

// AssetGroupConverter provides a mock function with given fields:
func (_m *RafterRetriever) AssetGroupConverter() shared.GqlAssetGroupConverter {
	ret := _m.Called()

	var r0 shared.GqlAssetGroupConverter
	if rf, ok := ret.Get(0).(func() shared.GqlAssetGroupConverter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(shared.GqlAssetGroupConverter)
		}
	}

	return r0
}

// ClusterAssetGroup provides a mock function with given fields:
func (_m *RafterRetriever) ClusterAssetGroup() shared.ClusterAssetGroupGetter {
	ret := _m.Called()

	var r0 shared.ClusterAssetGroupGetter
	if rf, ok := ret.Get(0).(func() shared.ClusterAssetGroupGetter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(shared.ClusterAssetGroupGetter)
		}
	}

	return r0
}

// ClusterAssetGroupConverter provides a mock function with given fields:
func (_m *RafterRetriever) ClusterAssetGroupConverter() shared.GqlClusterAssetGroupConverter {
	ret := _m.Called()

	var r0 shared.GqlClusterAssetGroupConverter
	if rf, ok := ret.Get(0).(func() shared.GqlClusterAssetGroupConverter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(shared.GqlClusterAssetGroupConverter)
		}
	}

	return r0
}

// Specification provides a mock function with given fields:
func (_m *RafterRetriever) Specification() shared.SpecificationGetter {
	ret := _m.Called()

	var r0 shared.SpecificationGetter
	if rf, ok := ret.Get(0).(func() shared.SpecificationGetter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(shared.SpecificationGetter)
		}
	}

	return r0
}
