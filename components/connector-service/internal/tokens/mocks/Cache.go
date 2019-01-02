// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"
import tokens "github.com/kyma-project/kyma/components/connector-service/internal/tokens"

// Cache is an autogenerated mock type for the Cache type
type Cache struct {
	mock.Mock
}

// Delete provides a mock function with given fields: app
func (_m *Cache) Delete(app string) {
	_m.Called(app)
}

// Get provides a mock function with given fields: app
func (_m *Cache) Get(app string) (*tokens.TokenData, bool) {
	ret := _m.Called(app)

	var r0 *tokens.TokenData
	if rf, ok := ret.Get(0).(func(string) *tokens.TokenData); ok {
		r0 = rf(app)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tokens.TokenData)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(app)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// Put provides a mock function with given fields: app, tokenData
func (_m *Cache) Put(app string, tokenData *tokens.TokenData) {
	_m.Called(app, tokenData)
}
