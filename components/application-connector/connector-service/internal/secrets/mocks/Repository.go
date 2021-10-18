// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	apperrors "github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"

	mock "github.com/stretchr/testify/mock"

	types "k8s.io/apimachinery/pkg/types"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *Repository) Get(_a0 context.Context, _a1 types.NamespacedName) (map[string][]byte, apperrors.AppError) {
	ret := _m.Called(_a0, _a1)

	var r0 map[string][]byte
	if rf, ok := ret.Get(0).(func(context.Context, types.NamespacedName) map[string][]byte); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]byte)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(context.Context, types.NamespacedName) apperrors.AppError); ok {
		r1 = rf(_a0, _a1)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
