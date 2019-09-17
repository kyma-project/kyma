// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"

import v1alpha1 "github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"

// clusterAddonsCfgUpdater is an autogenerated mock type for the clusterAddonsCfgUpdater type
type clusterAddonsCfgUpdater struct {
	mock.Mock
}

// AddRepos provides a mock function with given fields: name, repository
func (_m *clusterAddonsCfgUpdater) AddRepos(name string, repository []gqlschema.AddonsConfigurationRepositoryInput) (*v1alpha1.ClusterAddonsConfiguration, error) {
	ret := _m.Called(name, repository)

	var r0 *v1alpha1.ClusterAddonsConfiguration
	if rf, ok := ret.Get(0).(func(string, []gqlschema.AddonsConfigurationRepositoryInput) *v1alpha1.ClusterAddonsConfiguration); ok {
		r0 = rf(name, repository)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.ClusterAddonsConfiguration)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []gqlschema.AddonsConfigurationRepositoryInput) error); ok {
		r1 = rf(name, repository)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveRepos provides a mock function with given fields: name, reposToRemove
func (_m *clusterAddonsCfgUpdater) RemoveRepos(name string, reposToRemove []string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	ret := _m.Called(name, reposToRemove)

	var r0 *v1alpha1.ClusterAddonsConfiguration
	if rf, ok := ret.Get(0).(func(string, []string) *v1alpha1.ClusterAddonsConfiguration); ok {
		r0 = rf(name, reposToRemove)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.ClusterAddonsConfiguration)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []string) error); ok {
		r1 = rf(name, reposToRemove)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Resync provides a mock function with given fields: name
func (_m *clusterAddonsCfgUpdater) Resync(name string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	ret := _m.Called(name)

	var r0 *v1alpha1.ClusterAddonsConfiguration
	if rf, ok := ret.Get(0).(func(string) *v1alpha1.ClusterAddonsConfiguration); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.ClusterAddonsConfiguration)
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
