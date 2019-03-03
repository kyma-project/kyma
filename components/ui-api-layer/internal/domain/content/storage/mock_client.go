// Code generated by mockery v1.0.0. DO NOT EDIT.
package storage

import io "io"
import mock "github.com/stretchr/testify/mock"

// mockClient is an autogenerated mock type for the client type
type mockClient struct {
	mock.Mock
}

// IsInvalidBeginningCharacterError provides a mock function with given fields: err
func (_m *mockClient) IsInvalidBeginningCharacterError(err error) bool {
	ret := _m.Called(err)

	var r0 bool
	if rf, ok := ret.Get(0).(func(error) bool); ok {
		r0 = rf(err)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsNotExistsError provides a mock function with given fields: err
func (_m *mockClient) IsNotExistsError(err error) bool {
	ret := _m.Called(err)

	var r0 bool
	if rf, ok := ret.Get(0).(func(error) bool); ok {
		r0 = rf(err)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// NotificationChannel provides a mock function with given fields: bucketName, stop
func (_m *mockClient) NotificationChannel(bucketName string, stop <-chan struct{}) <-chan notification {
	ret := _m.Called(bucketName, stop)

	var r0 <-chan notification
	if rf, ok := ret.Get(0).(func(string, <-chan struct{}) <-chan notification); ok {
		r0 = rf(bucketName, stop)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan notification)
		}
	}

	return r0
}

// Object provides a mock function with given fields: bucketName, objectName
func (_m *mockClient) Object(bucketName string, objectName string) (io.ReadCloser, error) {
	ret := _m.Called(bucketName, objectName)

	var r0 io.ReadCloser
	if rf, ok := ret.Get(0).(func(string, string) io.ReadCloser); ok {
		r0 = rf(bucketName, objectName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(bucketName, objectName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
