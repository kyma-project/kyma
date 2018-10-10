package apperrors

import "fmt"

const (
	CodeInternal      = 1
	CodeNotFound      = 2
	CodeAlreadyExists = 3
	CodeWrongInput    = 4
)

type AppError interface {
	Code() int
	Error() string
}

type appError struct {
	code    int
	message string
}

func errorf(code int, format string, a ...interface{}) AppError {
	return appError{code: code, message: fmt.Sprintf(format, a...)}
}

func Internal(format string, a ...interface{}) AppError {
	return errorf(CodeInternal, format, a...)
}

func NotFound(format string, a ...interface{}) AppError {
	return errorf(CodeNotFound, format, a...)
}

func AlreadyExists(format string, a ...interface{}) AppError {
	return errorf(CodeAlreadyExists, format, a...)
}

func WrongInput(format string, a ...interface{}) AppError {
	return errorf(CodeWrongInput, format, a...)
}

func (ae appError) Code() int {
	return ae.code
}

func (ae appError) Error() string {
	return ae.message
}
