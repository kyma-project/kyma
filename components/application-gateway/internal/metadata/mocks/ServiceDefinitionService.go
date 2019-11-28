// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	apperrors "github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-project/kyma/components/application-gateway/internal/metadata/model"
)

// ServiceDefinitionService is an autogenerated mock type for the ServiceDefinitionService type
type ServiceDefinitionService struct {
	mock.Mock
}

// GetAPI provides a mock function with given fields: serviceId
func (_m *ServiceDefinitionService) GetAPI(serviceId string) (*model.API, apperrors.AppError) {
	ret := _m.Called(serviceId)

	var r0 *model.API
	if rf, ok := ret.Get(0).(func(string) *model.API); ok {
		r0 = rf(serviceId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.API)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string) apperrors.AppError); ok {
		r1 = rf(serviceId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
