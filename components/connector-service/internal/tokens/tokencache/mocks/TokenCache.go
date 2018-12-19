// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// TokenCache is an autogenerated mock type for the TokenCache type
type TokenCache struct {
	mock.Mock
}

// Delete provides a mock function with given fields: app
func (_m *TokenCache) Delete(app string) {
	_m.Called(app)
}

// Get provides a mock function with given fields: app
func (_m *TokenCache) Get(app string) (string, bool) {
	ret := _m.Called(app)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(app)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(app)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// Put provides a mock function with given fields: app, token
func (_m *TokenCache) Put(app string, token string) {
	_m.Called(app, token)
}
