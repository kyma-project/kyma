// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	apperrors "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"

	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// Create provides a mock function with given fields: _a0
func (_m *Repository) Create(_a0 *v1alpha1.Application) (*v1alpha1.Application, apperrors.AppError) {
	ret := _m.Called(_a0)

	var r0 *v1alpha1.Application
	if rf, ok := ret.Get(0).(func(*v1alpha1.Application) *v1alpha1.Application); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Application)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(*v1alpha1.Application) apperrors.AppError); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// Delete provides a mock function with given fields: name, options
func (_m *Repository) Delete(name string, options *v1.DeleteOptions) apperrors.AppError {
	ret := _m.Called(name, options)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string, *v1.DeleteOptions) apperrors.AppError); ok {
		r0 = rf(name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// Get provides a mock function with given fields: name, options
func (_m *Repository) Get(name string, options v1.GetOptions) (*v1alpha1.Application, apperrors.AppError) {
	ret := _m.Called(name, options)

	var r0 *v1alpha1.Application
	if rf, ok := ret.Get(0).(func(string, v1.GetOptions) *v1alpha1.Application); ok {
		r0 = rf(name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Application)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, v1.GetOptions) apperrors.AppError); ok {
		r1 = rf(name, options)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// List provides a mock function with given fields: opts
func (_m *Repository) List(opts v1.ListOptions) (*v1alpha1.ApplicationList, apperrors.AppError) {
	ret := _m.Called(opts)

	var r0 *v1alpha1.ApplicationList
	if rf, ok := ret.Get(0).(func(v1.ListOptions) *v1alpha1.ApplicationList); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.ApplicationList)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(v1.ListOptions) apperrors.AppError); ok {
		r1 = rf(opts)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// Update provides a mock function with given fields: _a0
func (_m *Repository) Update(_a0 *v1alpha1.Application) (*v1alpha1.Application, apperrors.AppError) {
	ret := _m.Called(_a0)

	var r0 *v1alpha1.Application
	if rf, ok := ret.Get(0).(func(*v1alpha1.Application) *v1alpha1.Application); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Application)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(*v1alpha1.Application) apperrors.AppError); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
