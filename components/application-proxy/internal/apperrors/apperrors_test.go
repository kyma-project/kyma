package apperrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {

	t.Run("should create error with proper code", func(t *testing.T) {
		assert.Equal(t, CodeInternal, Internal("error").Code())
		assert.Equal(t, CodeNotFound, NotFound("error").Code())
		assert.Equal(t, CodeAlreadyExists, AlreadyExists("error").Code())
		assert.Equal(t, CodeWrongInput, WrongInput("error").Code())
		assert.Equal(t, CodeUpstreamServerCallFailed, UpstreamServerCallFailed("error").Code())
	})

	t.Run("should create error with simple message", func(t *testing.T) {
		assert.Equal(t, "error", Internal("error").Error())
		assert.Equal(t, "error", NotFound("error").Error())
		assert.Equal(t, "error", AlreadyExists("error").Error())
		assert.Equal(t, "error", WrongInput("error").Error())
		assert.Equal(t, "error", UpstreamServerCallFailed("error").Error())
	})

	t.Run("should create error with formatted message", func(t *testing.T) {
		assert.Equal(t, "code: 1, error: bug", Internal("code: %d, error: %s", 1, "bug").Error())
		assert.Equal(t, "code: 1, error: bug", NotFound("code: %d, error: %s", 1, "bug").Error())
		assert.Equal(t, "code: 1, error: bug", AlreadyExists("code: %d, error: %s", 1, "bug").Error())
		assert.Equal(t, "code: 1, error: bug", WrongInput("code: %d, error: %s", 1, "bug").Error())
		assert.Equal(t, "code: 1, error: bug", UpstreamServerCallFailed("code: %d, error: %s", 1, "bug").Error())
	})
}
