package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type clusterContextMiddleware struct {
	required clientcontext.CtxRequiredType
}

func NewClusterContextMiddleware(required clientcontext.CtxRequiredType) *clusterContextMiddleware {
	return &clusterContextMiddleware{
		required: required,
	}
}

func (cc *clusterContextMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		clusterContext := cc.readClusterContextFromRequest(r)

		if clusterContext.IsEmpty() && bool(cc.required) {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Required headers for ClusterContext not specified."))
			return
		}

		reqWithCtx := r.WithContext(clusterContext.ExtendContext(r.Context()))
		handler.ServeHTTP(w, reqWithCtx)
	})
}

func (cc *clusterContextMiddleware) readClusterContextFromRequest(r *http.Request) clientcontext.ClusterContext {
	clusterContext := clientcontext.ClusterContext{
		Tenant: r.Header.Get(clientcontext.TenantHeader),
		Group:  r.Header.Get(clientcontext.GroupHeader),
	}

	return clusterContext
}
