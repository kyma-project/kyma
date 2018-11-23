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

	t.Run("should append apperrors without changing error code", func(t *testing.T) {
		//given
		createdInternalErr := Internal("Some Internal apperror, %s", "Some pkg err")
		createdNotFoundErr := NotFound("Some NotFound apperror, %s", "Some pkg err")
		createdAlreadyExistsErr := AlreadyExists("Some AlreadyExists apperror, %s", "Some pkg err")
		createdWrongInputErr := WrongInput("Some WrongInput apperror, %s", "Some pkg err")
		createdUpstreamServerCallFailedErr := UpstreamServerCallFailed("Some UpstreamServerCallFailed apperror, %s", "Some pkg err")

		//when
		appendedInternalErr := createdInternalErr.Append("Some additional message")
		appendedNotFoundErr := createdNotFoundErr.Append("Some additional message")
		appendedAlreadyExistsErr := createdAlreadyExistsErr.Append("Some additional message")
		appendedWrongInputErr := createdWrongInputErr.Append("Some additional message")
		appendedUpstreamServerCallFailedErr := createdUpstreamServerCallFailedErr.Append("Some additional message")

		//then
		assert.Equal(t, CodeInternal, appendedInternalErr.Code())
		assert.Equal(t, CodeNotFound, appendedNotFoundErr.Code())
		assert.Equal(t, CodeAlreadyExists, appendedAlreadyExistsErr.Code())
		assert.Equal(t, CodeWrongInput, appendedWrongInputErr.Code())
		assert.Equal(t, CodeUpstreamServerCallFailed, appendedUpstreamServerCallFailedErr.Code())
	})

	t.Run("should append apperrors and chain messages correctly", func(t *testing.T) {
		//given
		createdInternalErr := Internal("Some Internal apperror, %s", "Some pkg err")
		createdNotFoundErr := NotFound("Some NotFound apperror, %s", "Some pkg err")
		createdAlreadyExistsErr := AlreadyExists("Some AlreadyExists apperror, %s", "Some pkg err")
		createdWrongInputErr := WrongInput("Some WrongInput apperror, %s", "Some pkg err")
		createdUpstreamServerCallFailedErr := UpstreamServerCallFailed("Some UpstreamServerCallFailed apperror, %s", "Some pkg err")

		//when
		appendedInternalErr := createdInternalErr.Append("Some additional message: %s", "error")
		appendedNotFoundErr := createdNotFoundErr.Append("Some additional message: %s", "error")
		appendedAlreadyExistsErr := createdAlreadyExistsErr.Append("Some additional message: %s", "error")
		appendedWrongInputErr := createdWrongInputErr.Append("Some additional message: %s", "error")
		appendedUpstreamServerCallFailedErr := createdUpstreamServerCallFailedErr.Append("Some additional message: %s", "error")

		//then
		assert.Equal(t, "Some additional message: error, Some Internal apperror, Some pkg err", appendedInternalErr.Error())
		assert.Equal(t, "Some additional message: error, Some NotFound apperror, Some pkg err", appendedNotFoundErr.Error())
		assert.Equal(t, "Some additional message: error, Some AlreadyExists apperror, Some pkg err", appendedAlreadyExistsErr.Error())
		assert.Equal(t, "Some additional message: error, Some WrongInput apperror, Some pkg err", appendedWrongInputErr.Error())
		assert.Equal(t, "Some additional message: error, Some UpstreamServerCallFailed apperror, Some pkg err", appendedUpstreamServerCallFailedErr.Error())
	})
}
