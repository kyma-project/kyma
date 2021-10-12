package httphelpers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriterWithStatus_WriteHeader(t *testing.T) {

	testCases := []int{
		http.StatusOK,
		http.StatusBadRequest,
		http.StatusInternalServerError,
	}

	for _, status := range testCases {
		// given
		writer := WriterWithStatus{ResponseWriter: httptest.NewRecorder()}

		// when
		writer.WriteHeader(status)

		// then
		assert.Equal(t, status, writer.status)
	}

}

func TestWriterWithStatus_IsSuccessful(t *testing.T) {

	testCases := []struct {
		status int
		result bool
	}{
		{
			status: http.StatusOK,
			result: true,
		},
		{
			status: http.StatusCreated,
			result: true,
		},
		{
			status: http.StatusBadRequest,
			result: false,
		},
		{
			status: http.StatusInternalServerError,
			result: false,
		},
	}

	for _, test := range testCases {
		// given
		writer := WriterWithStatus{status: test.status}

		// when
		r := writer.IsSuccessful()

		// then
		assert.Equal(t, test.result, r)
	}
}
