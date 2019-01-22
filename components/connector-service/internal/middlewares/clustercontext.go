package middlewares

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
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
			respondWithError(w, apperrors.BadRequest("Cluster context is empty"))
			return
		}

		reqWithCtx := r.WithContext(context.WithValue(r.Context(), ClusterContextKey, clusterContext))

		handler.ServeHTTP(w, reqWithCtx)
	})
}

func respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
}

func respondWithError(w http.ResponseWriter, apperr apperrors.AppError) {
	statusCode, responseBody := httperrors.AppErrorToResponse(apperr)

	respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)

}
