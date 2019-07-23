// Code generated by mockery v1.0.0
package automock

import internal "github.com/kyma-project/kyma/components/helm-broker/internal"
import mock "github.com/stretchr/testify/mock"

// ClusterDocsProvider is an autogenerated mock type for the ClusterDocsProvider type
type ClusterDocsProvider struct {
	mock.Mock
}

// EnsureClusterDocsTopic provides a mock function with given fields: bundle
func (_m *ClusterDocsProvider) EnsureClusterDocsTopic(bundle *internal.Addon) error {
	ret := _m.Called(bundle)

	var r0 error
	if rf, ok := ret.Get(0).(func(*internal.Addon) error); ok {
		r0 = rf(bundle)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnsureClusterDocsTopicRemoved provides a mock function with given fields: id
func (_m *ClusterDocsProvider) EnsureClusterDocsTopicRemoved(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
