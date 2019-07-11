package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
)

type appContextMiddleware struct {
	contextHandler clientcontext.ClusterContextStrategy
}

func NewApplicationContextMiddleware(contextHandler clientcontext.ClusterContextStrategy) *appContextMiddleware {
	return &appContextMiddleware{
		contextHandler: contextHandler,
	}
}

func (cc *appContextMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appContext := cc.contextHandler.ReadClusterContextFromRequest(r)
		appContext.ID = r.Header.Get(clientcontext.ApplicationHeader)
		
		if !cc.contextHandler.IsValidContext(appContext) {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Required headers for ApplicationContext not specified."))
			return
		}

		reqWithCtx := r.WithContext(appContext.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
