package router

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/app"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	fixRouter := func(uaaEnabled bool) *chi.Mux {
		return New(app.App{UAAEnabled: uaaEnabled})
	}

	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name          string
		request       request
		matchingRoute string
		uaaEnabled    bool
	}{
		{
			name: "healthz",
			request: request{
				method: http.MethodGet,
				url:    "/healthz",
			},
			matchingRoute: "/healthz",
			uaaEnabled:    true,
		},
		{
			name: "redirect to busola",
			request: request{
				method: http.MethodGet,
				url:    "/console-redirect",
			},
			matchingRoute: "/console-redirect",
			uaaEnabled:    true,
		},
		{
			name: "migrate xsuaa",
			request: request{
				method: http.MethodGet,
				url:    "/xsuaa-migrate",
			},
			matchingRoute: "/xsuaa-migrate",
			uaaEnabled:    true,
		},
		{
			name: "xsuaa callback",
			request: request{
				method: http.MethodGet,
				url:    "/callback",
			},
			matchingRoute: "/callback",
			uaaEnabled:    true,
		},
		{
			name: "redirect to static page",
			request: request{
				method: http.MethodGet,
				url:    "/not-declared-in-router",
			},
			matchingRoute: "/*",
			uaaEnabled:    true,
		},
		{
			name: "static page index",
			request: request{
				method: http.MethodGet,
				url:    "/",
			},
			matchingRoute: "/",
			uaaEnabled:    true,
		},
		{
			name: "static page asset",
			request: request{
				method: http.MethodGet,
				url:    "/assets/style.css",
			},
			matchingRoute: "/assets/*",
			uaaEnabled:    true,
		},
		{
			name: "migrate xsuaa when uaa disabled",
			request: request{
				method: http.MethodGet,
				url:    "/xsuaa-migrate",
			},
			matchingRoute: "/*",
			uaaEnabled:    false,
		},
		{
			name: "xsuaa callback when uaa disabled",
			request: request{
				method: http.MethodGet,
				url:    "/callback",
			},
			matchingRoute: "/*",
			uaaEnabled:    false,
		},
	}

	for _, tt := range tests {
		// GIVEN
		rctx := chi.NewRouteContext()

		// WHEN
		match := fixRouter(tt.uaaEnabled).Match(rctx, tt.request.method, tt.request.url)

		// THEN
		assert.True(t, match)
		assert.Contains(t, rctx.RoutePatterns, tt.matchingRoute)
	}
}
