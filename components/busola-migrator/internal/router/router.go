package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/icza/session"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/app"
)

func New(app app.App) *chi.Mux {
	// disable session middleware logging
	session.Global = session.NewCookieManager(session.NewInMemStoreOptions(&session.InMemStoreOptions{Logger: session.NoopLogger}))

	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.Recoverer)

	// routes
	r.Get("/*", app.HandleInfoRedirect)
	r.Get("/info/*", app.HandleStaticWebsite)
	r.Get("/console-redirect", app.HandleConsoleRedirect)
	r.Get("/xsuaa-migrate", app.HandleXSUAAMigrate)
	r.Get("/callback", app.HandleXSUAACallback)
	r.Get("/healthz", app.HandleHealthy)

	return r
}
