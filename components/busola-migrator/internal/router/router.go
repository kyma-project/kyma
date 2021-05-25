package router

import (
	"github.com/kyma-project/kyma/components/busola-migrator/internal/app"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/icza/session"
)

func New(app app.App) *chi.Mux {
	// disable session middleware logging
	session.Global = session.NewCookieManager(session.NewInMemStoreOptions(&session.InMemStoreOptions{Logger: session.NoopLogger}))

	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.Recoverer)

	// routes
	r.Get("/*", app.HandleIndexRedirect)
	r.Get("/", app.HandleStaticIndex)
	r.Get("/assets/*", app.HandleStaticAssets)
	r.Get("/console-redirect", app.HandleConsoleRedirect)
	r.Get("/healthz", app.HandleHealthy)

	if app.UAAEnabled {
		r.Get("/success", app.HandleStaticSuccess)
		r.Get("/xsuaa-migrate", app.HandleXSUAAMigrate)
		r.Get("/callback", app.HandleXSUAACallback)
	}

	return r
}
