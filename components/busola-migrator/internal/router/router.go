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
	r.Use(middleware.Logger) // TODO: Disable logging (at least for k8s probes)
	r.Use(middleware.Recoverer)

	// routes
	r.Get("/*", app.HandleStaticWebsite)
	r.Get("/busola-redirect", app.HandleRedirect)
	r.Get("/xsuaa-migrate", app.HandleXSUAAMigrate)
	r.Get("/healthz", app.HandleHealthy)

	return r
}
