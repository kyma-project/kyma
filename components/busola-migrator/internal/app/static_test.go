package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestApp_HandleStaticWebsite(t *testing.T) {
	// GIVEN
	r, _ := http.NewRequest("GET", "/info/", nil)
	r = r.WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, &chi.Context{RoutePatterns: []string{"/info/*"}}))
	w := httptest.NewRecorder()

	testStaticFilesDir := "../../static"
	app := App{
		fsRoot: http.Dir(testStaticFilesDir),
	}
	handler := http.HandlerFunc(app.HandleStaticWebsite)

	// WHEN
	handler.ServeHTTP(w, r)
	res := w.Result()

	// THEN
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", res.Header.Get("Content-Type"))
}
