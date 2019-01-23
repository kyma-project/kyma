package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dummyKey   = "dummyKey"
	dummyValue = "dummyValue"
)

type dummyContextExtender struct {
}

func (dce dummyContextExtender) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, dummyKey, dummyValue)
}

func dummyExtender(token string, tokenResolver tokens.Resolver) (httpcontext.ContextExtender, apperrors.AppError) {
	return dummyContextExtender{}, nil
}

func TestTokenResolver_Middleware(t *testing.T) {
	t.Run("should resolve token and extend context", func(t *testing.T) {
		token := "token"

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctxValue := ctx.Value(dummyKey)

			assert.Equal(t, dummyValue, ctxValue)
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/?token="+token, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		middleware := NewTokenResolverMiddleware(nil, dummyExtender)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
