package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/httphelpers"
)

type appContextMiddleware struct {
	contextHandler clientcontext.ClusterContextStrategy
}

func (cc appContextMiddleware) isValidCtx(appCtx clientcontext.ApplicationContext) bool {
	return !appCtx.IsEmpty() && cc.contextHandler.IsValidContext(appCtx.ClusterContext)
}

func NewApplicationContextMiddleware(contextHandler clientcontext.ClusterContextStrategy) *appContextMiddleware {
	return &appContextMiddleware{
		contextHandler: contextHandler,
	}
}

func (cc *appContextMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appContext := clientcontext.ApplicationContext{
			Application:    r.Header.Get(clientcontext.ApplicationHeader),
			ClusterContext: cc.contextHandler.ReadClusterContextFromRequest(r),
		}

		if !cc.isValidCtx(appContext) {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Required headers for ApplicationContext not specified."))
			return
		}

		reqWithCtx := r.WithContext(appContext.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
