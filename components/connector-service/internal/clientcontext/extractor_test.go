package clientcontext

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appName = "appName"
	tenant  = "tenant"
	group   = "group"
	token   = "token"
)

func Test_ExtractSerializableApplicationContext(t *testing.T) {

	t.Run("should return ApplicationContext", func(t *testing.T) {
		appCtxPayload := ApplicationContext{
			Application:    appName,
			ClusterContext: ClusterContext{Group: group, Tenant: tenant},
		}

		ctx := appCtxPayload.ExtendContext(context.Background())

		serializable, err := ExtractApplicationContext(ctx)
		require.NoError(t, err)

		extractedContext, ok := serializable.(ApplicationContext)
		assert.True(t, ok)

		assert.Equal(t, appCtxPayload, extractedContext)
	})

	t.Run("should fail when there is no ApplicationContext", func(t *testing.T) {
		_, err := ExtractApplicationContext(context.Background())
		require.Error(t, err)

		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})
}

func Test_ExtractSerializableClusterContext(t *testing.T) {
	t.Run("should return ClusterToken", func(t *testing.T) {
		clusterCtxPayload := ClusterContext{Group: group, Tenant: tenant}

		ctx := clusterCtxPayload.ExtendContext(context.Background())

		serializable, err := ExtractClusterContext(ctx)
		require.NoError(t, err)

		extractedContext, ok := serializable.(ClusterContext)
		assert.True(t, ok)

		assert.Equal(t, clusterCtxPayload, extractedContext)
	})

	t.Run("should fail when there is no ClusterContext", func(t *testing.T) {
		_, err := ExtractClusterContext(context.Background())
		require.Error(t, err)

		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})
}

func Test_ResolveApplicationContextExtender(t *testing.T) {

	t.Run("should resolve application context extender", func(t *testing.T) {
		// given
		tokenResolver := &mocks.Resolver{}
		tokenResolver.On("Resolve", token, mock.AnythingOfType("*clientcontext.ApplicationContext")).
			Return(nil)

		// when
		extender, err := ResolveApplicationContextExtender(token, tokenResolver)

		// then
		require.NoError(t, err)
		require.NotNil(t, extender)
		require.IsType(t, ApplicationContext{}, extender)
	})

	t.Run("should return error when failed to resolve", func(t *testing.T) {
		// given
		tokenResolver := &mocks.Resolver{}
		tokenResolver.On("Resolve", token, mock.AnythingOfType("*clientcontext.ApplicationContext")).
			Return(apperrors.Internal("error"))

		// when
		extender, err := ResolveApplicationContextExtender(token, tokenResolver)

		// then
		require.Error(t, err)
		require.Empty(t, extender)
	})
}

func Test_ResolveClusterContextExtender(t *testing.T) {

	t.Run("should return error when failed to resolve", func(t *testing.T) {
		// given
		tokenResolver := &mocks.Resolver{}
		tokenResolver.On("Resolve", token, mock.AnythingOfType("*clientcontext.ClusterContext")).
			Return(apperrors.Internal("error"))

		// when
		extender, err := ResolveClusterContextExtender(token, tokenResolver)

		// then
		require.Error(t, err)
		require.Empty(t, extender)
	})
}
