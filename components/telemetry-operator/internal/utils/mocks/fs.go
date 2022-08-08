// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils"
	mock "github.com/stretchr/testify/mock"
)

// FileSystem is an autogenerated mock type for the FileSystem type
type FileSystem struct {
	mock.Mock
}

// CreateAndWrite provides a mock function with given fields: s
func (_m *FileSystem) CreateAndWrite(s utils.File) error {
	ret := _m.Called(s)

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.File) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveDirectory provides a mock function with given fields: path
func (_m *FileSystem) RemoveDirectory(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewFileSystem interface {
	mock.TestingT
	Cleanup(func())
}

// NewFileSystem creates a new instance of FileSystem. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFileSystem(t mockConstructorTestingTNewFileSystem) *FileSystem {
	mock := &FileSystem{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
