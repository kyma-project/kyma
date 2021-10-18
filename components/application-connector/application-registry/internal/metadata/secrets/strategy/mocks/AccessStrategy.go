// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import apperrors "github.com/kyma-project/kyma/components/application-connector/application-registry/internal/apperrors"
import applications "github.com/kyma-project/kyma/components/application-connector/application-registry/internal/metadata/applications"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-project/kyma/components/application-connector/application-registry/internal/metadata/model"
import strategy "github.com/kyma-project/kyma/components/application-connector/application-registry/internal/metadata/secrets/strategy"

// AccessStrategy is an autogenerated mock type for the AccessStrategy type
type AccessStrategy struct {
	mock.Mock
}

// ToCredentials provides a mock function with given fields: secretData, appCredentials
func (_m *AccessStrategy) ToCredentials(secretData strategy.SecretData, appCredentials *applications.Credentials) (model.CredentialsWithCSRF, apperrors.AppError) {
	ret := _m.Called(secretData, appCredentials)

	var r0 model.CredentialsWithCSRF
	if rf, ok := ret.Get(0).(func(strategy.SecretData, *applications.Credentials) model.CredentialsWithCSRF); ok {
		r0 = rf(secretData, appCredentials)
	} else {
		r0 = ret.Get(0).(model.CredentialsWithCSRF)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(strategy.SecretData, *applications.Credentials) apperrors.AppError); ok {
		r1 = rf(secretData, appCredentials)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
