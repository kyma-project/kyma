// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"

// Collector is an autogenerated mock type for the Collector type
type Collector struct {
	mock.Mock
}

// AddObservation provides a mock function with given fields: pathTemplate, requestMethod, message
func (_m *Collector) AddObservation(pathTemplate string, requestMethod string, message float64) {
	_m.Called(pathTemplate, requestMethod, message)
}
