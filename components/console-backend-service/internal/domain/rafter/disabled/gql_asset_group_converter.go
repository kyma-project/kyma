// Code generated by failery v1.0.0. DO NOT EDIT.

package disabled

import (
	gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	v1beta1 "github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

// gqlAssetGroupConverter is an autogenerated failing mock type for the gqlAssetGroupConverter type
type gqlAssetGroupConverter struct {
	err error
}

// NewGqlAssetGroupConverter creates a new gqlAssetGroupConverter type instance
func NewGqlAssetGroupConverter(err error) *gqlAssetGroupConverter {
	return &gqlAssetGroupConverter{err: err}
}

// ToGQL provides a failing mock function with given fields: in
func (_m *gqlAssetGroupConverter) ToGQL(in *v1beta1.AssetGroup) (*gqlschema.AssetGroup, error) {
	var r0 *gqlschema.AssetGroup
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ToGQLs provides a failing mock function with given fields: in
func (_m *gqlAssetGroupConverter) ToGQLs(in []*v1beta1.AssetGroup) ([]gqlschema.AssetGroup, error) {
	var r0 []gqlschema.AssetGroup
	var r1 error
	r1 = _m.err

	return r0, r1
}
