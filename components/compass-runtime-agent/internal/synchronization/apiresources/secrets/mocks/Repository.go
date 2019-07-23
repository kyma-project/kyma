// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import apperrors "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
import mock "github.com/stretchr/testify/mock"

import strategy "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/apiresources/secrets/strategy"
import types "k8s.io/apimachinery/pkg/types"

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// Create provides a mock function with given fields: application, appUID, name, serviceID, data
func (_m *Repository) Create(application string, appUID types.UID, name string, serviceID string, data strategy.SecretData) apperrors.AppError {
	ret := _m.Called(application, appUID, name, serviceID, data)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string, types.UID, string, string, strategy.SecretData) apperrors.AppError); ok {
		r0 = rf(application, appUID, name, serviceID, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// Delete provides a mock function with given fields: name
func (_m *Repository) Delete(name string) apperrors.AppError {
	ret := _m.Called(name)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string) apperrors.AppError); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// Get provides a mock function with given fields: name
func (_m *Repository) Get(name string) (strategy.SecretData, apperrors.AppError) {
	ret := _m.Called(name)

	var r0 strategy.SecretData
	if rf, ok := ret.Get(0).(func(string) strategy.SecretData); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(strategy.SecretData)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string) apperrors.AppError); ok {
		r1 = rf(name)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// Upsert provides a mock function with given fields: application, appUID, name, secretID, data
func (_m *Repository) Upsert(application string, appUID types.UID, name string, secretID string, data strategy.SecretData) apperrors.AppError {
	ret := _m.Called(application, appUID, name, secretID, data)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string, types.UID, string, string, strategy.SecretData) apperrors.AppError); ok {
		r0 = rf(application, appUID, name, secretID, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}
