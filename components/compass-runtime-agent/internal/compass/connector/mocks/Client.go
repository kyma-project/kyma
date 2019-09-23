// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import gqlschema "github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
import mock "github.com/stretchr/testify/mock"

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Configuration provides a mock function with given fields: headers
func (_m *Client) Configuration(headers map[string][]string) (gqlschema.Configuration, error) {
	ret := _m.Called(headers)

	var r0 gqlschema.Configuration
	if rf, ok := ret.Get(0).(func(map[string][]string) gqlschema.Configuration); ok {
		r0 = rf(headers)
	} else {
		r0 = ret.Get(0).(gqlschema.Configuration)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(map[string][]string) error); ok {
		r1 = rf(headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignCSR provides a mock function with given fields: csr, headers
func (_m *Client) SignCSR(csr string, headers map[string][]string) (gqlschema.CertificationResult, error) {
	ret := _m.Called(csr, headers)

	var r0 gqlschema.CertificationResult
	if rf, ok := ret.Get(0).(func(string, map[string][]string) gqlschema.CertificationResult); ok {
		r0 = rf(csr, headers)
	} else {
		r0 = ret.Get(0).(gqlschema.CertificationResult)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, map[string][]string) error); ok {
		r1 = rf(csr, headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
