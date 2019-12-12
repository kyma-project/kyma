// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// serviceChecker is an autogenerated mock type for the serviceChecker type
type serviceChecker struct {
	mock.Mock
}

// WaitUntilIsAvailable provides a mock function with given fields: url, timeout
func (_m *serviceChecker) WaitUntilIsAvailable(url string, timeout time.Duration) {
	_m.Called(url, timeout)
}
