package middlewares

import (
	"context"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

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
			respondWithError(w, apperrors.BadRequest("Application context is empty"))
			return
		}

		reqWithCtx := r.WithContext(context.WithValue(r.Context(), ApplicationContextKey, appContext))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
