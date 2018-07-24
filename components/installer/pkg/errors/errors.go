package errors

import (
	"log"
)

// ErrorHandlersInterface .
type ErrorHandlersInterface interface {
	CheckError(msg string, err error) bool
	LogError(msg string, err error)
}

// ErrorHandlers .
type ErrorHandlers struct {
}

// CheckError .
func (eh *ErrorHandlers) CheckError(msg string, err error) bool {
	if err != nil {
		log.Printf("%s Details: %s", msg, err.Error())
		return true
	}
	return false
}

// LogError .
func (eh *ErrorHandlers) LogError(msg string, err error) {
	if err != nil {
		log.Printf("%s Details: %s", msg, err.Error())
	}
}
