package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/app"
)

func New(app app.App) *chi.Mux {
	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// routes
	r.With(middleware.Logger).Get("/*", app.HandleInfoRedirect)
	r.With(middleware.Logger).Get("/info/*", app.HandleStaticWebsite)
	r.With(middleware.Logger).Get("/console-redirect", app.HandleConsoleRedirect)
	r.With(middleware.Logger).Get("/xsuaa-migrate", app.HandleXSUAAMigrate)
	r.Get("/healthz", app.HandleHealthy)

	return r
}
