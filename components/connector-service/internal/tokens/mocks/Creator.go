// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import (
	apperrors "github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	mock "github.com/stretchr/testify/mock"
)

// Creator is an autogenerated mock type for the Creator type
type Creator struct {
	mock.Mock
}

// Save provides a mock function with given fields: _a0
func (_m *Creator) Save(_a0 interface{}) (string, apperrors.AppError) {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(interface{}) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(interface{}) apperrors.AppError); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
