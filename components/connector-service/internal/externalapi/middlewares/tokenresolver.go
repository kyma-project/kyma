package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

type ExtenderConstructor func() clientcontext.ContextExtender

type tokenResolverMiddleware struct {
	tokenManager        tokens.Manager
	extenderConstructor ExtenderConstructor
}

func NewTokenResolverMiddleware(tokenManager tokens.Manager, extenderConstructor ExtenderConstructor) *tokenResolverMiddleware {
	return &tokenResolverMiddleware{
		tokenManager:        tokenManager,
		extenderConstructor: extenderConstructor,
	}
}

func (cc *tokenResolverMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			httphelpers.RespondWithErrorAndLog(w, apperrors.Forbidden("Token not provided."))
			return
		}

		connectorClientContext := cc.extenderConstructor()

		err := cc.tokenManager.Resolve(token, connectorClientContext)
		if err != nil {
			if err.Code() == apperrors.CodeNotFound {
				httphelpers.RespondWithErrorAndLog(w, apperrors.Forbidden("Invalid token."))
			} else {
				httphelpers.RespondWithErrorAndLog(w, apperrors.Internal("Failed to resolve token."))
			}

			return
		}

		reqWithCtx := r.WithContext(connectorClientContext.ExtendContext(r.Context()))
		writerWithStatus := httphelpers.WriterWithStatus{ResponseWriter: w}

		handler.ServeHTTP(&writerWithStatus, reqWithCtx)

		if writerWithStatus.IsSuccessful() {
			cc.tokenManager.Delete(token)
		}
	})
}
