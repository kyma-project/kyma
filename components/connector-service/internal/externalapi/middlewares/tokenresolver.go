package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

type ExtenderConstructor func(token string, tokenResolver tokens.Resolver) (clientcontext.ContextExtender, apperrors.AppError)

type tokenResolverMiddleware struct {
	tokenResolver       tokens.Resolver
	extenderConstructor ExtenderConstructor
}

func NewTokenResolverMiddleware(tokenResolver tokens.Resolver, extenderConstructor ExtenderConstructor) *tokenResolverMiddleware {
	return &tokenResolverMiddleware{
		tokenResolver:       tokenResolver,
		extenderConstructor: extenderConstructor,
	}
}

func (cc *tokenResolverMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			httphelpers.RespondWithError(w, apperrors.Forbidden("Token not provided."))
			return
		}

		ctxExtender, err := cc.extenderConstructor(token, cc.tokenResolver)
		if err != nil {
			if err.Code() == apperrors.CodeNotFound {
				httphelpers.RespondWithError(w, apperrors.Forbidden("Invalid token."))
			} else {
				httphelpers.RespondWithError(w, apperrors.Internal("Failed to resolve token."))
			}

			return
		}

		reqWithCtx := r.WithContext(ctxExtender.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
