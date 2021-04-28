package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp_HandleHealthy(t *testing.T) {
	// GIVEN
	r, _ := http.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	app := App{}
	handler := http.HandlerFunc(app.HandleHealthy)

	// WHEN
	handler.ServeHTTP(w, r)
	res := w.Result()

	// THEN
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
