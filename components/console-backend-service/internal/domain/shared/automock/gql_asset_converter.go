// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"

import v1beta1 "github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"

// GqlAssetConverter is an autogenerated mock type for the GqlAssetConverter type
type GqlAssetConverter struct {
	mock.Mock
}

// ToGQL provides a mock function with given fields: item
func (_m *GqlAssetConverter) ToGQL(item *v1beta1.Asset) (*gqlschema.Asset, error) {
	ret := _m.Called(item)

	var r0 *gqlschema.Asset
	if rf, ok := ret.Get(0).(func(*v1beta1.Asset) *gqlschema.Asset); ok {
		r0 = rf(item)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.Asset)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1beta1.Asset) error); ok {
		r1 = rf(item)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGQLs provides a mock function with given fields: in
func (_m *GqlAssetConverter) ToGQLs(in []*v1beta1.Asset) ([]gqlschema.Asset, error) {
	ret := _m.Called(in)

	var r0 []gqlschema.Asset
	if rf, ok := ret.Get(0).(func([]*v1beta1.Asset) []gqlschema.Asset); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.Asset)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*v1beta1.Asset) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
