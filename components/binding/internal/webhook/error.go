package webhook

import "fmt"

// Error holds info about error message and its status code
type Error struct {
	errMsg  string
	errCode int32
}

// NewError returns new error
func NewError(code int32, msg string) *Error {
	return &Error{
		errMsg:  msg,
		errCode: code,
	}
}

// NewError returns new error
func NewErrorf(code int32, msg string, args ...interface{}) *Error {
	return &Error{
		errMsg:  fmt.Sprintf(msg, args...),
		errCode: code,
	}
}

// Error returns error message
func (m *Error) Error() string {
	return m.errMsg
}

// Code returns error status code
func (m *Error) Code() int32 {
	return m.errCode
}
