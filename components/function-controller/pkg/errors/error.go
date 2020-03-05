package errors

import (
	"fmt"
)

type code int8

const (
	errEmptyValue code = iota + 1
	errInvalidValue
	errInvalidState
)

var (
	_ error = xtdError{}
)

type xtdError struct {
	msg  interface{}
	code code
}

func (e xtdError) Error() string {
	return fmt.Sprintf("%+v", e.msg)
}

func NewEmptyValue(msg interface{}) error {
	return &xtdError{
		msg:  msg,
		code: errEmptyValue,
	}
}

func NewInvalidValue(msg interface{}) error {
	return &xtdError{
		msg:  msg,
		code: errInvalidValue,
	}
}

func NewInvalidState(msg interface{}) error {
	return &xtdError{
		msg:  msg,
		code: errInvalidState,
	}
}

func hasCode(i interface{}, code code) bool {
	err, ok := i.(*xtdError)
	if !ok {
		return false
	}
	return err.code == code
}

func IsEmptyVal(err error) bool {
	return hasCode(err, errEmptyValue)
}

func IsInvalidVal(err error) bool {
	return hasCode(err, errInvalidValue)
}

func IsInvalidState(err error) bool {
	return hasCode(err, errInvalidState)
}
