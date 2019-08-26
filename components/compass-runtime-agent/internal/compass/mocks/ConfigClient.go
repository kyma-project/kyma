// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import certificates "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"

import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"

// ConfigClient is an autogenerated mock type for the ConfigClient type
type ConfigClient struct {
	mock.Mock
}

// FetchConfiguration provides a mock function with given fields: directorURL, credentials
func (_m *ConfigClient) FetchConfiguration(directorURL string, credentials certificates.Credentials) ([]model.Application, error) {
	ret := _m.Called(directorURL, credentials)

	var r0 []model.Application
	if rf, ok := ret.Get(0).(func(string, certificates.Credentials) []model.Application); ok {
		r0 = rf(directorURL, credentials)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, certificates.Credentials) error); ok {
		r1 = rf(directorURL, credentials)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
