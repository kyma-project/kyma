package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"
)

type clusterContextMiddleware struct {
	contextHandler clientcontext.ClusterContextStrategy
}

func NewClusterContextMiddleware(contextHandler clientcontext.ClusterContextStrategy) *clusterContextMiddleware {
	return &clusterContextMiddleware{
		contextHandler: contextHandler,
	}
}

func (cc *clusterContextMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clusterContext := cc.contextHandler.ReadClusterContextFromRequest(r)

		if !cc.contextHandler.IsValidContext(clusterContext) {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Required headers for ClusterContext not specified."))
			return
		}

		reqWithCtx := r.WithContext(clusterContext.ExtendContext(r.Context()))
		handler.ServeHTTP(w, reqWithCtx)
	})
}
