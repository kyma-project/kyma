package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type clusterContextMiddleware struct {
	defaultGroup  string
	defaultTenant string
}

func NewClusterContextMiddleware(tenant, group string) *clusterContextMiddleware {
	return &clusterContextMiddleware{
		defaultGroup:  group,
		defaultTenant: tenant,
	}
}

func (cc *clusterContextMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		clusterContext := cc.readClusterContextFromRequest(r)

		if clusterContext.IsEmpty() {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Required headers not specified."))
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

	if cc.defaultTenant != "" {
		clusterContext.Tenant = cc.defaultTenant
	}

	if cc.defaultGroup != "" {
		clusterContext.Group = cc.defaultGroup
	}

	return clusterContext
}
