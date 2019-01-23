package middlewares

import (
	"context"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

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
		clusterContext := ClusterContext{
			Tenant: r.Header.Get(TenantHeader),
			Group:  r.Header.Get(GroupHeader),
		}

		if cc.defaultTenant != "" {
			clusterContext.Tenant = cc.defaultTenant
		}

		if cc.defaultGroup != "" {
			clusterContext.Group = cc.defaultGroup
		}

		if clusterContext.IsEmpty() {
			httphelpers.RespondWithError(w, apperrors.BadRequest("Cluster context is empty"))
			return
		}

		reqWithCtx := r.WithContext(context.WithValue(r.Context(), ClusterContextKey, clusterContext))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
