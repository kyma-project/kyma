package httperrors

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestHttpErrors_AppErrorToResponse(t *testing.T) {
	t.Run("should print short internal error", func(t *testing.T) {
		//given
		error := apperrors.Internal("some internal error occured")
		detailedErrorResponse := false

		//when
		status, body := AppErrorToResponse(error, detailedErrorResponse)

		//then
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, "Internal error.", body.Error)
	})

	t.Run("should print detailed internal error", func(t *testing.T) {
		//given
		message := "detailed internal error message"
		error := apperrors.Internal(message)
		detailedErrorResponse := true

		//when
		status, body := AppErrorToResponse(error, detailedErrorResponse)

		//then
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, message, body.Error)
	})

	t.Run("should print normal not found error", func(t *testing.T) {
		//given
		message := "normal non internal error message"
		error := apperrors.NotFound(message)
		detailedErrorResponse := false

		//when
		status, body := AppErrorToResponse(error, detailedErrorResponse)

		//then
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, message, body.Error)
	})

	t.Run("should print normal code already exists error", func(t *testing.T) {
		//given
		message := "normal non internal error message"
		error := apperrors.AlreadyExists(message)
		detailedErrorResponse := false

		//when
		status, body := AppErrorToResponse(error, detailedErrorResponse)

		//then
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, message, body.Error)
	})

	t.Run("should print normal bad request error", func(t *testing.T) {
		//given
		message := "normal non internal error message"
		error := apperrors.WrongInput(message)
		detailedErrorResponse := false

		//when
		status, body := AppErrorToResponse(error, detailedErrorResponse)

		//then
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, message, body.Error)
	})
}