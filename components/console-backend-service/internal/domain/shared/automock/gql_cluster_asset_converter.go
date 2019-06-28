// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"

import v1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"

// GqlClusterAssetConverter is an autogenerated mock type for the GqlClusterAssetConverter type
type GqlClusterAssetConverter struct {
	mock.Mock
}

// ToGQL provides a mock function with given fields: item
func (_m *GqlClusterAssetConverter) ToGQL(item *v1alpha2.ClusterAsset) (*gqlschema.ClusterAsset, error) {
	ret := _m.Called(item)

	var r0 *gqlschema.ClusterAsset
	if rf, ok := ret.Get(0).(func(*v1alpha2.ClusterAsset) *gqlschema.ClusterAsset); ok {
		r0 = rf(item)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.ClusterAsset)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha2.ClusterAsset) error); ok {
		r1 = rf(item)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGQLs provides a mock function with given fields: in
func (_m *GqlClusterAssetConverter) ToGQLs(in []*v1alpha2.ClusterAsset) ([]gqlschema.ClusterAsset, error) {
	ret := _m.Called(in)

	var r0 []gqlschema.ClusterAsset
	if rf, ok := ret.Get(0).(func([]*v1alpha2.ClusterAsset) []gqlschema.ClusterAsset); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.ClusterAsset)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*v1alpha2.ClusterAsset) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
