package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp_HandleXSUAAMigrate(t *testing.T) {
	// GIVEN
	r, _ := http.NewRequest("GET", "/xsuaa-migrate", nil)
	w := httptest.NewRecorder()
	app := New(testBusolaURL, testStaticFilesDir)
	handler := http.HandlerFunc(app.HandleXSUAAMigrate)

	// WHEN
	handler.ServeHTTP(w, r)
	res := w.Result()

	// THEN
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode) // not yet implemented
}
