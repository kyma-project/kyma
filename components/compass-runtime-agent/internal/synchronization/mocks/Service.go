// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import apperrors "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
import compass "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
import mock "github.com/stretchr/testify/mock"
import synchronization "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization"

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// Apply provides a mock function with given fields: applications
func (_m *Service) Apply(applications []compass.Application) ([]synchronization.Result, apperrors.AppError) {
	ret := _m.Called(applications)

	var r0 []synchronization.Result
	if rf, ok := ret.Get(0).(func([]compass.Application) []synchronization.Result); ok {
		r0 = rf(applications)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]synchronization.Result)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func([]compass.Application) apperrors.AppError); ok {
		r1 = rf(applications)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
