package broker_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type notFoundError struct{}

func (notFoundError) Error() string  { return "element not found" }
func (notFoundError) NotFound() bool { return true }

func TestIsNotForbiddenError(t *testing.T) {
	// GIVEN
	err := errors.New("simple error")
	// WHEN
	ok := broker.IsForbiddenError(err)
	// THEN
	assert.False(t, ok)
}

func TestIsForbiddenError(t *testing.T) {
	// GIVEN
	err := &broker.ForbiddenError{}
	// WHEN
	ok := broker.IsForbiddenError(err)
	// THEN
	assert.True(t, ok)
}

func TestIsNotForbiddenErrorWhenNil(t *testing.T) {
	// WHEN
	ok := broker.IsForbiddenError(nil)
	// THEN
	assert.False(t, ok)

}
