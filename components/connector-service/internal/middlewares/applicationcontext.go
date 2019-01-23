package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type appContextMiddleware struct {
	clusterContextMiddleware
}

func NewApplicationContextMiddleware() *appContextMiddleware {
	return &appContextMiddleware{}
}

func (cc *appContextMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appContext := httpcontext.ApplicationContext{
			Application:    r.Header.Get(httpcontext.ApplicationHeader),
			ClusterContext: cc.readClusterContextFromRequest(r),
		}

		if appContext.IsEmpty() {
			httphelpers.RespondWithError(w, apperrors.BadRequest("Application context is empty"))
			return
		}

		reqWithCtx := r.WithContext(appContext.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
