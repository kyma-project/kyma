package internal

import (
	"log"
)

type ErrorHandlersInterface interface {
	CheckError(msg string, err error) bool
	LogError(msg string, err error)
}

type ErrorHandlers struct {
}

func (eh *ErrorHandlers) CheckError(msg string, err error) bool {
	if err != nil {
		log.Printf("%s Details: %s", msg, err.Error())
		return true
	}
	return false
}

func (eh *ErrorHandlers) LogError(msg string, err error) {
	if err != nil {
		log.Printf("%s Details: %s", msg, err.Error())
	}
}
