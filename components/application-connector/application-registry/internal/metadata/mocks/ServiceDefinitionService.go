// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import (
	apperrors "github.com/kyma-project/kyma/components/application-connector/application-registry/internal/apperrors"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-project/kyma/components/application-connector/application-registry/internal/metadata/model"
)

// ServiceDefinitionService is an autogenerated mock type for the ServiceDefinitionService type
type ServiceDefinitionService struct {
	mock.Mock
}

// Create provides a mock function with given fields: application, serviceDefinition
func (_m *ServiceDefinitionService) Create(application string, serviceDefinition *model.ServiceDefinition) (string, apperrors.AppError) {
	ret := _m.Called(application, serviceDefinition)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, *model.ServiceDefinition) string); ok {
		r0 = rf(application, serviceDefinition)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, *model.ServiceDefinition) apperrors.AppError); ok {
		r1 = rf(application, serviceDefinition)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// Delete provides a mock function with given fields: application, id
func (_m *ServiceDefinitionService) Delete(application string, id string) apperrors.AppError {
	ret := _m.Called(application, id)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string, string) apperrors.AppError); ok {
		r0 = rf(application, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// GetAPI provides a mock function with given fields: application, serviceId
func (_m *ServiceDefinitionService) GetAPI(application string, serviceId string) (*model.API, apperrors.AppError) {
	ret := _m.Called(application, serviceId)

	var r0 *model.API
	if rf, ok := ret.Get(0).(func(string, string) *model.API); ok {
		r0 = rf(application, serviceId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.API)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, string) apperrors.AppError); ok {
		r1 = rf(application, serviceId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// GetAll provides a mock function with given fields: application
func (_m *ServiceDefinitionService) GetAll(application string) ([]model.ServiceDefinition, apperrors.AppError) {
	ret := _m.Called(application)

	var r0 []model.ServiceDefinition
	if rf, ok := ret.Get(0).(func(string) []model.ServiceDefinition); ok {
		r0 = rf(application)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.ServiceDefinition)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string) apperrors.AppError); ok {
		r1 = rf(application)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: application, id
func (_m *ServiceDefinitionService) GetByID(application string, id string) (model.ServiceDefinition, apperrors.AppError) {
	ret := _m.Called(application, id)

	var r0 model.ServiceDefinition
	if rf, ok := ret.Get(0).(func(string, string) model.ServiceDefinition); ok {
		r0 = rf(application, id)
	} else {
		r0 = ret.Get(0).(model.ServiceDefinition)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, string) apperrors.AppError); ok {
		r1 = rf(application, id)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// Update provides a mock function with given fields: application, serviceDef
func (_m *ServiceDefinitionService) Update(application string, serviceDef *model.ServiceDefinition) (model.ServiceDefinition, apperrors.AppError) {
	ret := _m.Called(application, serviceDef)

	var r0 model.ServiceDefinition
	if rf, ok := ret.Get(0).(func(string, *model.ServiceDefinition) model.ServiceDefinition); ok {
		r0 = rf(application, serviceDef)
	} else {
		r0 = ret.Get(0).(model.ServiceDefinition)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, *model.ServiceDefinition) apperrors.AppError); ok {
		r1 = rf(application, serviceDef)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
