package router

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/app"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	fixRouter := func() *chi.Mux {
		return New(app.New("https://busola.url", "/static"))
	}

	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name          string
		request       request
		expectedRoute string
	}{
		{
			name: "healthz",
			request: request{
				method: http.MethodGet,
				url:    "/healthz",
			},
			expectedRoute: "/healthz",
		},
		{
			name: "redirect to busola",
			request: request{
				method: http.MethodGet,
				url:    "/console-redirect",
			},
			expectedRoute: "/console-redirect",
		},
		{
			name: "redirect to static page",
			request: request{
				method: http.MethodGet,
				url:    "/not-declared-in-router",
			},
			expectedRoute: "/*",
		},
		{
			name: "static page",
			request: request{
				method: http.MethodGet,
				url:    "/info/",
			},
			expectedRoute: "/info/*",
		},
	}

	for _, tt := range tests {
		// GIVEN
		rctx := chi.NewRouteContext()

		// WHEN
		match := fixRouter().Match(rctx, tt.request.method, tt.request.url)

		// THEN
		assert.True(t, match)
		assert.Contains(t, rctx.RoutePatterns, tt.expectedRoute)
	}
}
