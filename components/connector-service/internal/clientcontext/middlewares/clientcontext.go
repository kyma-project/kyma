package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
)

type contextMiddleware struct {
	contextHandler clientcontext.ClusterContextStrategy
	idHeader       string
}

func NewContextMiddleware(contextHandler clientcontext.ClusterContextStrategy, idHeader string) *contextMiddleware {
	return &contextMiddleware{
		contextHandler: contextHandler,
		idHeader:       idHeader,
	}
}

func (cc *contextMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appContext := cc.contextHandler.ReadClusterContextFromRequest(r)
		appContext.ID = r.Header.Get(cc.idHeader)

		if !cc.contextHandler.IsValidContext(appContext) {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Required headers for ClientContext not specified."))
			return
		}

		reqWithCtx := r.WithContext(appContext.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
