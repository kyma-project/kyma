package httperrors

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/stretchr/testify/assert"
)

func TestHttpErrors_AppErrorToResponse(t *testing.T) {
	t.Run("should print short internal error", func(t *testing.T) {
		//given
		err := apperrors.Internal("some internal error occured")

		//when
		status, body := AppErrorToResponse(err, false)

		//then
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, "Internal error.", body.Error)
	})

	t.Run("should print detailed internal error", func(t *testing.T) {
		//given
		message := "detailed internal error message"
		err := apperrors.Internal(message)

		//when
		status, body := AppErrorToResponse(err, true)

		//then
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, message, body.Error)
	})

	t.Run("should print normal not found error", func(t *testing.T) {
		//given
		message := "normal non internal error message"
		err := apperrors.NotFound(message)

		//when
		status, body := AppErrorToResponse(err, false)

		//then
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, message, body.Error)
	})

	t.Run("should print normal code already exists error", func(t *testing.T) {
		//given
		message := "normal non internal error message"
		err := apperrors.AlreadyExists(message)

		//when
		status, body := AppErrorToResponse(err, false)

		//then
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, message, body.Error)
	})

	t.Run("should print normal bad request error", func(t *testing.T) {
		//given
		message := "normal non internal error message"
		err := apperrors.WrongInput(message)

		//when
		status, body := AppErrorToResponse(err, false)

		//then
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, message, body.Error)
	})

	t.Run("should print normal bad gateway error", func(t *testing.T) {
		//given
		message := "normal non internal error message"
		err := apperrors.UpstreamServerCallFailed(message)

		//when
		status, body := AppErrorToResponse(err, false)

		//then
		assert.Equal(t, http.StatusBadGateway, status)
		assert.Equal(t, message, body.Error)
	})
}
