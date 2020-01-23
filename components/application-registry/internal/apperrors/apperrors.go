package apperrors

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	CodeInternal                 = 1
	CodeNotFound                 = 2
	CodeAlreadyExists            = 3
	CodeWrongInput               = 4
	CodeUpstreamServerCallFailed = 5

	urlWithBasicAuthRegexpReplaceString = "$1://***:***@$4"
)

var (
	urlWithBasicAuthRegexp = regexp.MustCompile("(.+)://(.+):(.+)@(.+)")
)

type AppError interface {
	Append(string, ...interface{}) AppError
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

func UpstreamServerCallFailed(format string, a ...interface{}) AppError {
	return errorf(CodeUpstreamServerCallFailed, format, a...)
}

func (ae appError) Append(additionalFormat string, a ...interface{}) AppError {
	format := additionalFormat + ", " + ae.Error()
	return errorf(ae.code, format, a...)
}

func (ae appError) Code() int {
	return ae.code
}

func (ae appError) Error() string {
	return hideBasicCredentials(ae.message)
}

func hideBasicCredentials(str string) (output string) {
	strSplitted := strings.Split(str, " ")
	for _, strPart := range strSplitted {
		output = fmt.Sprintf(
			"%s%s ",
			output,
			urlWithBasicAuthRegexp.ReplaceAllString(strPart, urlWithBasicAuthRegexpReplaceString),
		)
	}
	if length := len(output); length > 0 {
		output = output[:length-1]
	}
	return output
}
