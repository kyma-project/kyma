// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"
import rsa "crypto/rsa"
import x509 "crypto/x509"

// Provider is an autogenerated mock type for the Provider type
type Provider struct {
	mock.Mock
}

// GetCACertificate provides a mock function with given fields:
func (_m *Provider) GetCACertificate() (*x509.Certificate, error) {
	ret := _m.Called()

	var r0 *x509.Certificate
	if rf, ok := ret.Get(0).(func() *x509.Certificate); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*x509.Certificate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetClientCredentials provides a mock function with given fields:
func (_m *Provider) GetClientCredentials() (*rsa.PrivateKey, *x509.Certificate, error) {
	ret := _m.Called()

	var r0 *rsa.PrivateKey
	if rf, ok := ret.Get(0).(func() *rsa.PrivateKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rsa.PrivateKey)
		}
	}

	var r1 *x509.Certificate
	if rf, ok := ret.Get(1).(func() *x509.Certificate); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*x509.Certificate)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
