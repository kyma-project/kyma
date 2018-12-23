// Code generated by mockery v1.0.0
package automock

import chart "k8s.io/helm/pkg/proto/hapi/chart"
import helm "github.com/kyma-project/kyma/components/helm-broker/internal/helm"
import internal "github.com/kyma-project/kyma/components/helm-broker/internal"
import mock "github.com/stretchr/testify/mock"

// HelmClient is an autogenerated mock type for the HelmClient type
type HelmClient struct {
	mock.Mock
}

// Delete provides a mock function with given fields: _a0
func (_m *HelmClient) Delete(_a0 internal.ReleaseName) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(internal.ReleaseName) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InstallOrUpdate provides a mock function with given fields: c, cv, releaseName, namespace
func (_m *HelmClient) InstallOrUpdate(c *chart.Chart, cv internal.ChartValues, releaseName internal.ReleaseName, namespace internal.Namespace) (*helm.ReleaseResponse, error) {
	ret := _m.Called(c, cv, releaseName, namespace)

	var r0 *helm.ReleaseResponse
	if rf, ok := ret.Get(0).(func(*chart.Chart, internal.ChartValues, internal.ReleaseName, internal.Namespace) *helm.ReleaseResponse); ok {
		r0 = rf(c, cv, releaseName, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*helm.ReleaseResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*chart.Chart, internal.ChartValues, internal.ReleaseName, internal.Namespace) error); ok {
		r1 = rf(c, cv, releaseName, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
