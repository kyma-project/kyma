// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"
import pkix "crypto/x509/pkix"
import rsa "crypto/rsa"

// CSRProvider is an autogenerated mock type for the CSRProvider type
type CSRProvider struct {
	mock.Mock
}

// CreateCSR provides a mock function with given fields: subject
func (_m *CSRProvider) CreateCSR(subject pkix.Name) (string, *rsa.PrivateKey, error) {
	ret := _m.Called(subject)

	var r0 string
	if rf, ok := ret.Get(0).(func(pkix.Name) string); ok {
		r0 = rf(subject)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 *rsa.PrivateKey
	if rf, ok := ret.Get(1).(func(pkix.Name) *rsa.PrivateKey); ok {
		r1 = rf(subject)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*rsa.PrivateKey)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(pkix.Name) error); ok {
		r2 = rf(subject)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
