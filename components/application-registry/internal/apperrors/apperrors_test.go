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

	t.Run("should hide basic credentials from message", func(t *testing.T) {
		// given
		simpleError := UpstreamServerCallFailed("error: https://user:password@dummy.com")
		formattedError := UpstreamServerCallFailed("code: %d, error: %s", 1, "https://user:password@dummy.com")
		appendedError := UpstreamServerCallFailed("error1: https://user:password@dummy.com").Append("error2: https://user:password@dummy.com")
		appendedFormatedError := UpstreamServerCallFailed("error1: https://user:password@dummy.com").Append("code: %d, error2: %s", 1, "https://user:password@dummy.com")

		// then
		assert.Equal(t, "error: https://***:***@dummy.com", simpleError.Error())
		assert.Equal(t, "code: 1, error: https://***:***@dummy.com", formattedError.Error())
		assert.Equal(t, "error2: https://***:***@dummy.com, error1: https://***:***@dummy.com", appendedError.Error())
		assert.Equal(t, "code: 1, error2: https://***:***@dummy.com, error1: https://***:***@dummy.com", appendedFormatedError.Error())
	})
}
