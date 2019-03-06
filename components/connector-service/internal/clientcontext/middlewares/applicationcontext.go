package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
)

type appContextMiddleware struct {
	*clusterContextMiddleware
}

func (cc appContextMiddleware) isValidCtx(appCtx clientcontext.ApplicationContext) bool {
	return (appCtx.IsEmpty() || (appCtx.ClusterContext.IsEmpty() && bool(cc.clusterContextMiddleware.required))) == false
}

func NewApplicationContextMiddleware(clusterContextMiddleware *clusterContextMiddleware) *appContextMiddleware {
	return &appContextMiddleware{
		clusterContextMiddleware: clusterContextMiddleware,
	}
}

func (cc *appContextMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appContext := clientcontext.ApplicationContext{
			Application:    r.Header.Get(clientcontext.ApplicationHeader),
			ClusterContext: cc.readClusterContextFromRequest(r),
		}

		if cc.isValidCtx(appContext) == false {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Required headers for ApplicationContext not specified."))
			return
		}

		reqWithCtx := r.WithContext(appContext.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
