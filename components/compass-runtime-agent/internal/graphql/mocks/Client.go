// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import machineboxgraphql "github.com/machinebox/graphql"
import mock "github.com/stretchr/testify/mock"

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// DisableLogging provides a mock function with given fields:
func (_m *Client) DisableLogging() {
	_m.Called()
}

// Do provides a mock function with given fields: req, res
func (_m *Client) Do(req *machineboxgraphql.Request, res interface{}) error {
	ret := _m.Called(req, res)

	var r0 error
	if rf, ok := ret.Get(0).(func(*machineboxgraphql.Request, interface{}) error); ok {
		r0 = rf(req, res)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
