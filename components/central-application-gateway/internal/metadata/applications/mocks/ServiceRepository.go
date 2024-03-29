// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	applications "github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/applications"
	apperrors "github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"

	mock "github.com/stretchr/testify/mock"
)

// ServiceRepository is an autogenerated mock type for the ServiceRepository type
type ServiceRepository struct {
	mock.Mock
}

// GetByEntryName provides a mock function with given fields: appName, serviceName, entryName
func (_m *ServiceRepository) GetByEntryName(appName string, serviceName string, entryName string) (applications.Service, apperrors.AppError) {
	ret := _m.Called(appName, serviceName, entryName)

	var r0 applications.Service
	if rf, ok := ret.Get(0).(func(string, string, string) applications.Service); ok {
		r0 = rf(appName, serviceName, entryName)
	} else {
		r0 = ret.Get(0).(applications.Service)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, string, string) apperrors.AppError); ok {
		r1 = rf(appName, serviceName, entryName)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// GetByServiceName provides a mock function with given fields: appName, serviceName
func (_m *ServiceRepository) GetByServiceName(appName string, serviceName string) (applications.Service, apperrors.AppError) {
	ret := _m.Called(appName, serviceName)

	var r0 applications.Service
	if rf, ok := ret.Get(0).(func(string, string) applications.Service); ok {
		r0 = rf(appName, serviceName)
	} else {
		r0 = ret.Get(0).(applications.Service)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, string) apperrors.AppError); ok {
		r1 = rf(appName, serviceName)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

type mockConstructorTestingTNewServiceRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewServiceRepository creates a new instance of ServiceRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewServiceRepository(t mockConstructorTestingTNewServiceRepository) *ServiceRepository {
	mock := &ServiceRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
