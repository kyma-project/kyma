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
		return New(app.App{})
	}

	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name          string
		request       request
		matchingRoute string
	}{
		{
			name: "healthz",
			request: request{
				method: http.MethodGet,
				url:    "/healthz",
			},
			matchingRoute: "/healthz",
		},
		{
			name: "redirect to busola",
			request: request{
				method: http.MethodGet,
				url:    "/console-redirect",
			},
			matchingRoute: "/console-redirect",
		},
		{
			name: "migrate xsuaa",
			request: request{
				method: http.MethodGet,
				url:    "/xsuaa-migrate",
			},
			matchingRoute: "/xsuaa-migrate",
		},
		{
			name: "xsuaa callback",
			request: request{
				method: http.MethodGet,
				url:    "/callback",
			},
			matchingRoute: "/callback",
		},
		{
			name: "redirect to static page",
			request: request{
				method: http.MethodGet,
				url:    "/not-declared-in-router",
			},
			matchingRoute: "/*",
		},
		{
			name: "static page",
			request: request{
				method: http.MethodGet,
				url:    "/info/",
			},
			matchingRoute: "/info/*",
		},
	}

	for _, tt := range tests {
		// GIVEN
		rctx := chi.NewRouteContext()

		// WHEN
		match := fixRouter().Match(rctx, tt.request.method, tt.request.url)

		// THEN
		assert.True(t, match)
		assert.Contains(t, rctx.RoutePatterns, tt.matchingRoute)
	}
}
