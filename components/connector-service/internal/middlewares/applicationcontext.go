package middlewares

import (
	"context"
	"net/http"
)

const (
	ApplicationHeader     = "Application"
	ApplicationContextKey = "ApplicationContext"
)

type ApplicationContext struct {
	Application string
}

// TODO - tests
// IsEmpty returns false if both Group and Tenant are set
func (context ApplicationContext) IsEmpty() bool {
	return context.Application == ""
}

type appContextMiddleware struct {
}

func NewAppContextMiddleware() *appContextMiddleware {
	return &appContextMiddleware{}
}

func (cc *appContextMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appContext := ApplicationContext{
			Application: r.Header.Get(ApplicationHeader),
		}

		if appContext.IsEmpty() {
			// TODO - error msg + logging
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		reqWithCtx := r.WithContext(context.WithValue(r.Context(), ApplicationContextKey, appContext))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
