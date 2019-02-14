package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type appContextMiddleware struct {
	*clusterContextMiddleware
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

		if appContext.IsEmpty() {
			httphelpers.RespondWithError(w, apperrors.BadRequest("Required headers not specified."))
			return
		}

		reqWithCtx := r.WithContext(appContext.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
