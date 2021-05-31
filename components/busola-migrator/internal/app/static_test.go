package app

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/app/automock"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/renderer"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestApp_HandleStaticAssets(t *testing.T) {
	// GIVEN
	testStaticAssetsDir := "../../static/assets"
	app := App{
		fsAssets: http.Dir(testStaticAssetsDir),
	}
	handler := http.HandlerFunc(app.HandleStaticAssets)

	t.Run("success when accessing asset", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/assets/style.css", nil)
		r = r.WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, &chi.Context{RoutePatterns: []string{"/assets/*"}}))
		w := httptest.NewRecorder()

		// WHEN
		handler.ServeHTTP(w, r)
		res := w.Result()

		// THEN
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "text/css; charset=utf-8", res.Header.Get("Content-Type"))
	})

	t.Run("redirect to root path when accessing directory", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/assets/", nil)
		r = r.WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, &chi.Context{RoutePatterns: []string{"/assets/*"}}))
		w := httptest.NewRecorder()

		// WHEN
		handler.ServeHTTP(w, r)
		res := w.Result()

		// THEN
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/", res.Header.Get("Location"))
	})
}

func TestApp_HandleStaticIndex(t *testing.T) {
	// GIVEN
	r, _ := http.NewRequest("GET", "/", nil)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()

		testBody := "<b>TEST</b>"
		mockRenderer := &automock.HTMLRenderer{}
		mockRenderer.On("RenderTemplate", w, renderer.TemplateNameIndex, mock.Anything).
			Run(func(args mock.Arguments) {
				w := args.Get(0).(io.Writer)
				w.Write([]byte(testBody))
			}).
			Return(nil).
			Once()

		app := App{
			htmlRenderer: mockRenderer,
		}
		handler := http.HandlerFunc(app.HandleStaticIndex)

		// WHEN
		handler.ServeHTTP(w, r)
		res := w.Result()

		// THEN
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", res.Header.Get("Content-Type"))
		resBody, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, testBody, string(resBody))
	})

	t.Run("error while rendering template", func(t *testing.T) {
		w := httptest.NewRecorder()

		testError := errors.New("test error")

		mockRenderer := &automock.HTMLRenderer{}
		mockRenderer.On("RenderTemplate", w, renderer.TemplateNameIndex, mock.Anything).
			Return(testError).
			Once()

		app := App{
			htmlRenderer: mockRenderer,
		}
		handler := http.HandlerFunc(app.HandleStaticIndex)

		// WHEN
		handler.ServeHTTP(w, r)
		res := w.Result()

		// THEN
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))
		resBody, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Contains(t, string(resBody), "Unable to render webpage")
	})
}

func TestApp_HandleStaticSuccess(t *testing.T) {
	// GIVEN
	r, _ := http.NewRequest("GET", "/", nil)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()

		testBody := "<b>TEST</b>"
		mockRenderer := &automock.HTMLRenderer{}
		mockRenderer.On("RenderTemplate", w, renderer.TemplateNameSuccess, mock.Anything).
			Run(func(args mock.Arguments) {
				w := args.Get(0).(io.Writer)
				w.Write([]byte(testBody))
			}).
			Return(nil).
			Once()

		app := App{
			htmlRenderer: mockRenderer,
		}
		handler := http.HandlerFunc(app.HandleStaticSuccess)

		// WHEN
		handler.ServeHTTP(w, r)
		res := w.Result()

		// THEN
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", res.Header.Get("Content-Type"))
		resBody, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, testBody, string(resBody))
	})

	t.Run("error while rendering template", func(t *testing.T) {
		w := httptest.NewRecorder()

		testError := errors.New("test error")

		mockRenderer := &automock.HTMLRenderer{}
		mockRenderer.On("RenderTemplate", w, renderer.TemplateNameSuccess, mock.Anything).
			Return(testError).
			Once()

		app := App{
			htmlRenderer: mockRenderer,
		}
		handler := http.HandlerFunc(app.HandleStaticSuccess)

		// WHEN
		handler.ServeHTTP(w, r)
		res := w.Result()

		// THEN
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))
		resBody, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Contains(t, string(resBody), "Unable to render webpage")
	})
}
