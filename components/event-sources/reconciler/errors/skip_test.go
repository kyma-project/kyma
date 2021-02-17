package errors

import (
	"errors"
	"testing"
)

func TestSkippableError(t *testing.T) {
	nilErr := (error)(nil)
	if NewSkippable(nilErr).Error() != "" {
		t.Error("Expected empty string")
	}

	const errStr = "some error"

	nonNilErr := errors.New(errStr)
	if NewSkippable(nonNilErr).Error() != errStr {
		t.Errorf("Expected error string to be %q", errStr)
	}
}
