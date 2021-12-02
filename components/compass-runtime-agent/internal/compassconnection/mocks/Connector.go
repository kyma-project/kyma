// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	certificates "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	compassconnection "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compassconnection"

	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
)

// Connector is an autogenerated mock type for the Connector type
type Connector struct {
	mock.Mock
}

// EstablishConnection provides a mock function with given fields: connectorURL, token
func (_m *Connector) EstablishConnection(connectorURL string, token string) (compassconnection.EstablishedConnection, error) {
	ret := _m.Called(connectorURL, token)

	var r0 compassconnection.EstablishedConnection
	if rf, ok := ret.Get(0).(func(string, string) compassconnection.EstablishedConnection); ok {
		r0 = rf(connectorURL, token)
	} else {
		r0 = ret.Get(0).(compassconnection.EstablishedConnection)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(connectorURL, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MaintainConnection provides a mock function with given fields: renewCert, credentialsExist
func (_m *Connector) MaintainConnection(renewCert bool, credentialsExist bool) (*certificates.Credentials, v1alpha1.ManagementInfo, error) {
	ret := _m.Called(renewCert, credentialsExist)

	var r0 *certificates.Credentials
	if rf, ok := ret.Get(0).(func(bool, bool) *certificates.Credentials); ok {
		r0 = rf(renewCert, credentialsExist)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*certificates.Credentials)
		}
	}

	var r1 v1alpha1.ManagementInfo
	if rf, ok := ret.Get(1).(func(bool, bool) v1alpha1.ManagementInfo); ok {
		r1 = rf(renewCert, credentialsExist)
	} else {
		r1 = ret.Get(1).(v1alpha1.ManagementInfo)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(bool, bool) error); ok {
		r2 = rf(renewCert, credentialsExist)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
