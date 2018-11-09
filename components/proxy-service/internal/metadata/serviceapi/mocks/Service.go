// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import apperrors "github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/model"
import remoteenv "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/remoteenv"

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// Read provides a mock function with given fields: _a0
func (_m *Service) Read(_a0 *remoteenv.ServiceAPI) (*model.API, apperrors.AppError) {
	ret := _m.Called(_a0)

	var r0 *model.API
	if rf, ok := ret.Get(0).(func(*remoteenv.ServiceAPI) *model.API); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.API)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(*remoteenv.ServiceAPI) apperrors.AppError); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
