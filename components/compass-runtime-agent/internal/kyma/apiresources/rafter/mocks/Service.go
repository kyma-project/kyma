// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	apperrors "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	clusterassetgroup "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// Delete provides a mock function with given fields: id
func (_m *Service) Delete(id string) apperrors.AppError {
	ret := _m.Called(id)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string) apperrors.AppError); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// Put provides a mock function with given fields: id, assets
func (_m *Service) Put(id string, assets []clusterassetgroup.Asset) apperrors.AppError {
	ret := _m.Called(id, assets)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string, []clusterassetgroup.Asset) apperrors.AppError); ok {
		r0 = rf(id, assets)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}
