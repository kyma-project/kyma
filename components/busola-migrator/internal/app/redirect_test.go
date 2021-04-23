package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testBusolaURL = "https://busola-url.com"
)

func TestApp_HandleInfoRedirect(t *testing.T) {
	// GIVEN
	r, _ := http.NewRequest("GET", "/any-url", nil)
	w := httptest.NewRecorder()
	app := App{}
	handler := http.HandlerFunc(app.HandleInfoRedirect)

	// WHEN
	handler.ServeHTTP(w, r)
	res := w.Result()

	// THEN
	assert.Equal(t, http.StatusFound, res.StatusCode)
	assert.Equal(t, "/info/", res.Header.Get("Location"))
}

func TestApp_HandleConsoleRedirect(t *testing.T) {
	// GIVEN
	r, _ := http.NewRequest("GET", "/console-redirect", nil)
	w := httptest.NewRecorder()
	app := App{busolaURL: testBusolaURL}
	handler := http.HandlerFunc(app.HandleConsoleRedirect)

	// WHEN
	handler.ServeHTTP(w, r)
	res := w.Result()

	// THEN
	assert.Equal(t, http.StatusFound, res.StatusCode)
	assert.Equal(t, testBusolaURL, res.Header.Get("Location"))
}
