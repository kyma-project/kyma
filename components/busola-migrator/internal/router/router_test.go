package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/app"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	t.Parallel()
	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name       string
		router     *chi.Mux
		wantCode   int
		wantHeader *[]string
		request    request
	}{
		{
			name:   "healthz",
			router: New(app.New("some-url", ".")),
			request: request{
				method: http.MethodGet,
				url:    "/healthz",
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "redirect to busola",
			router: New(app.New("some-url", ".")),
			request: request{
				method: http.MethodGet,
				url:    "/console-redirect",
			},
			wantCode:   http.StatusFound,
			wantHeader: &[]string{"Location", "/some-url"},
		},
		{
			name:   "redirect to static page",
			router: New(app.New("some-url", ".")),
			request: request{
				method: http.MethodGet,
				url:    "/not-declared-in-router",
			},
			wantCode:   http.StatusFound,
			wantHeader: &[]string{"Location", "/info/"},
		},
		{
			name:   "static page",
			router: New(app.New("some-url", "../../static")),
			request: request{
				method: http.MethodGet,
				url:    "/info/",
			},
			wantCode:   http.StatusOK,
			wantHeader: &[]string{"Content-Type", "text/html; charset=utf-8"},
		},
	}
	for _, tt := range tests {
		r, _ := http.NewRequest(tt.request.method, tt.request.url, nil)
		w := httptest.NewRecorder()
		tt.router.ServeHTTP(w, r)

		res := w.Result()
		assert.Equal(t, tt.wantCode, res.StatusCode)

		if tt.wantHeader != nil {
			assert.Equal(t, (*tt.wantHeader)[1], res.Header.Get((*tt.wantHeader)[0]))
		}
	}
}
