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
	token      = "token"
)

type dummyContextExtender struct {
}

func (dce dummyContextExtender) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, dummyKey, dummyValue)
}

func dummyExtender(token string, tokenResolver tokens.Resolver) (httpcontext.ContextExtender, apperrors.AppError) {
	return dummyContextExtender{}, nil
}

func notFoundExtender(token string, tokenResolver tokens.Resolver) (httpcontext.ContextExtender, apperrors.AppError) {
	return dummyContextExtender{}, apperrors.NotFound("Not found")
}

func internalErrorExtender(token string, tokenResolver tokens.Resolver) (httpcontext.ContextExtender, apperrors.AppError) {
	return dummyContextExtender{}, apperrors.Internal("Error")
}

func TestTokenResolver_Middleware(t *testing.T) {
	t.Run("should resolve token and extend context", func(t *testing.T) {
		// given
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

	t.Run("should return 403 when there is no token sent", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		middleware := NewTokenResolverMiddleware(nil, nil)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("should return 403 when token is not found", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		middleware := NewTokenResolverMiddleware(nil, notFoundExtender)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("should return 500 when internal error occured", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/?token="+token, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		middleware := NewTokenResolverMiddleware(nil, internalErrorExtender)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
