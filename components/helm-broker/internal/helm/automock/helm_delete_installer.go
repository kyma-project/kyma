// Code generated by mockery v1.0.0
package automock

import chart "k8s.io/helm/pkg/proto/hapi/chart"
import helm "k8s.io/helm/pkg/helm"

import mock "github.com/stretchr/testify/mock"
import services "k8s.io/helm/pkg/proto/hapi/services"

// HelmDeleteInstaller is an autogenerated mock type for the HelmDeleteInstaller type
type HelmDeleteInstaller struct {
	mock.Mock
}

// DeleteRelease provides a mock function with given fields: rlsName, opts
func (_m *HelmDeleteInstaller) DeleteRelease(rlsName string, opts ...helm.DeleteOption) (*services.UninstallReleaseResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, rlsName)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *services.UninstallReleaseResponse
	if rf, ok := ret.Get(0).(func(string, ...helm.DeleteOption) *services.UninstallReleaseResponse); ok {
		r0 = rf(rlsName, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*services.UninstallReleaseResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, ...helm.DeleteOption) error); ok {
		r1 = rf(rlsName, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InstallReleaseFromChart provides a mock function with given fields: _a0, ns, opts
func (_m *HelmDeleteInstaller) InstallReleaseFromChart(_a0 *chart.Chart, ns string, opts ...helm.InstallOption) (*services.InstallReleaseResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, ns)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *services.InstallReleaseResponse
	if rf, ok := ret.Get(0).(func(*chart.Chart, string, ...helm.InstallOption) *services.InstallReleaseResponse); ok {
		r0 = rf(_a0, ns, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*services.InstallReleaseResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*chart.Chart, string, ...helm.InstallOption) error); ok {
		r1 = rf(_a0, ns, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListReleases provides a mock function with given fields: opts
func (_m *HelmDeleteInstaller) ListReleases(opts ...helm.ReleaseListOption) (*services.ListReleasesResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *services.ListReleasesResponse
	if rf, ok := ret.Get(0).(func(...helm.ReleaseListOption) *services.ListReleasesResponse); ok {
		r0 = rf(opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*services.ListReleasesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(...helm.ReleaseListOption) error); ok {
		r1 = rf(opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateReleaseFromChart provides a mock function with given fields: rlsName, _a1, opts
func (_m *HelmDeleteInstaller) UpdateReleaseFromChart(rlsName string, _a1 *chart.Chart, opts ...helm.UpdateOption) (*services.UpdateReleaseResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, rlsName, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *services.UpdateReleaseResponse
	if rf, ok := ret.Get(0).(func(string, *chart.Chart, ...helm.UpdateOption) *services.UpdateReleaseResponse); ok {
		r0 = rf(rlsName, _a1, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*services.UpdateReleaseResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *chart.Chart, ...helm.UpdateOption) error); ok {
		r1 = rf(rlsName, _a1, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
