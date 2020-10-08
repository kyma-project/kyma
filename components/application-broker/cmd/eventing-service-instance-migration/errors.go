package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
)

// TypeNotFoundError is returned when a CRD is missing / a type is not registered.
type TypeNotFoundError string

// Error implements the error interface.
func (e *TypeNotFoundError) Error() string {
	return fmt.Sprintf("CRD %q does not exist", *e)
}

// NewTypeNotFoundError returns a new TypeNotFoundError.
func NewTypeNotFoundError(msg string) error {
	typeErr := TypeNotFoundError(msg)
	return &typeErr
}

// handleAndTerminate logs an error and terminates the program.
func handleAndTerminate(err error, context string) {
	if _, ok := errors.Cause(err).(*TypeNotFoundError); ok {
		log.Printf("Skipping migration: %s", err)
		os.Exit(0)
	}

	log.Fatalf("Error %s: %s", context, err)
}
