// Code generated by failery v1.0.0. DO NOT EDIT.

package disabled

import (
	resource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	v1beta1 "github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

// clusterAssetSvc is an autogenerated failing mock type for the clusterAssetSvc type
type clusterAssetSvc struct {
	err error
}

// NewClusterAssetSvc creates a new clusterAssetSvc type instance
func NewClusterAssetSvc(err error) *clusterAssetSvc {
	return &clusterAssetSvc{err: err}
}

// Find provides a failing mock function with given fields: name
func (_m *clusterAssetSvc) Find(name string) (*v1beta1.ClusterAsset, error) {
	var r0 *v1beta1.ClusterAsset
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ListForClusterAssetGroupByType provides a failing mock function with given fields: assetGroupName, types
func (_m *clusterAssetSvc) ListForClusterAssetGroupByType(assetGroupName string, types []string) ([]*v1beta1.ClusterAsset, error) {
	var r0 []*v1beta1.ClusterAsset
	var r1 error
	r1 = _m.err

	return r0, r1
}

// Subscribe provides a failing mock function with given fields: listener
func (_m *clusterAssetSvc) Subscribe(listener resource.Listener) {
}

// Unsubscribe provides a failing mock function with given fields: listener
func (_m *clusterAssetSvc) Unsubscribe(listener resource.Listener) {
}
