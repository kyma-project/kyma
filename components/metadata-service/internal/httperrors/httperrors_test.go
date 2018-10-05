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
		error := apperrors.Internal("code internal")
		detailedErrorResponse := false

		//when
		status, body := AppErrorToResponse(error, detailedErrorResponse)

		//then
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, "Internal error.", body.Error)
	})

	t.Run("should print long internal error", func(t *testing.T) {
		//given
		error := apperrors.Internal("long internal error message")
		detailedErrorResponse := true

		//when
		status, body := AppErrorToResponse(error, detailedErrorResponse)

		//then
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, "long internal error message", body.Error)
	})

	t.Run("should print normal non internal error", func(t *testing.T) {
		//given
		error := apperrors.NotFound("normal non internal error message")
		detailedErrorResponse := false

		//when
		status, body := AppErrorToResponse(error, detailedErrorResponse)

		//then
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "normal non internal error message", body.Error)
	})
}