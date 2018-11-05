// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import apperrors "github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
import mock "github.com/stretchr/testify/mock"

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// CreateBasicAuthSecret provides a mock function with given fields: remoteEnvironment, name, username, password, serviceID
func (_m *Service) CreateBasicAuthSecret(remoteEnvironment string, name string, username string, password string, serviceID string) apperrors.AppError {
	ret := _m.Called(remoteEnvironment, name, username, password, serviceID)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) apperrors.AppError); ok {
		r0 = rf(remoteEnvironment, name, username, password, serviceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// CreateOauthSecret provides a mock function with given fields: remoteEnvironment, name, clientID, clientSecret, serviceID
func (_m *Service) CreateOauthSecret(remoteEnvironment string, name string, clientID string, clientSecret string, serviceID string) apperrors.AppError {
	ret := _m.Called(remoteEnvironment, name, clientID, clientSecret, serviceID)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) apperrors.AppError); ok {
		r0 = rf(remoteEnvironment, name, clientID, clientSecret, serviceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// DeleteSecret provides a mock function with given fields: name
func (_m *Service) DeleteSecret(name string) apperrors.AppError {
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

// GetBasicAuthSecret provides a mock function with given fields: remoteEnvironment, name
func (_m *Service) GetBasicAuthSecret(remoteEnvironment string, name string) (string, string, apperrors.AppError) {
	ret := _m.Called(remoteEnvironment, name)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(remoteEnvironment, name)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(string, string) string); ok {
		r1 = rf(remoteEnvironment, name)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 apperrors.AppError
	if rf, ok := ret.Get(2).(func(string, string) apperrors.AppError); ok {
		r2 = rf(remoteEnvironment, name)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(apperrors.AppError)
		}
	}

	return r0, r1, r2
}

// GetOauthSecret provides a mock function with given fields: remoteEnvironment, name
func (_m *Service) GetOauthSecret(remoteEnvironment string, name string) (string, string, apperrors.AppError) {
	ret := _m.Called(remoteEnvironment, name)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(remoteEnvironment, name)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(string, string) string); ok {
		r1 = rf(remoteEnvironment, name)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 apperrors.AppError
	if rf, ok := ret.Get(2).(func(string, string) apperrors.AppError); ok {
		r2 = rf(remoteEnvironment, name)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(apperrors.AppError)
		}
	}

	return r0, r1, r2
}

// UpdateBasicAuthSecret provides a mock function with given fields: remoteEnvironment, name, username, password, serviceID
func (_m *Service) UpdateBasicAuthSecret(remoteEnvironment string, name string, username string, password string, serviceID string) apperrors.AppError {
	ret := _m.Called(remoteEnvironment, name, username, password, serviceID)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) apperrors.AppError); ok {
		r0 = rf(remoteEnvironment, name, username, password, serviceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// UpdateOauthSecret provides a mock function with given fields: remoteEnvironment, name, clientID, clientSecret, serviceID
func (_m *Service) UpdateOauthSecret(remoteEnvironment string, name string, clientID string, clientSecret string, serviceID string) apperrors.AppError {
	ret := _m.Called(remoteEnvironment, name, clientID, clientSecret, serviceID)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) apperrors.AppError); ok {
		r0 = rf(remoteEnvironment, name, clientID, clientSecret, serviceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}
