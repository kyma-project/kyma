package middlewares

import (
	"context"
	"net/http"
)

// TODO - consider moving to different package
const (
	ClusterContextKey = "ClusterContext"
	TenantHeader      = "Tenant"
	GroupHeader       = "Group"
)

type ClusterContext struct {
	Group  string
	Tenant string
}

// TODO - tests
// IsEmpty returns false if both Group and Tenant are set
func (context ClusterContext) IsEmpty() bool {
	return context.Group == "" || context.Tenant == ""
}

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
			// TODO - error msg + logging
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		reqWithCtx := r.WithContext(context.WithValue(r.Context(), ClusterContextKey, clusterContext))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
