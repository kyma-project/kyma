package tokens

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appName = "appName"
	tenant  = "tenant"
	group   = "group"
)

func TestParams_NewApplicationTokenParams(t *testing.T) {

	t.Run("should return ApplictionTokenParams", func(t *testing.T) {
		appCtxPayload := middlewares.ApplicationContext{Application: appName}
		clusterCtxPayload := middlewares.ClusterContext{Group: group, Tenant: tenant}

		ctx := context.WithValue(context.Background(), middlewares.ApplicationContextKey, appCtxPayload)
		ctx = context.WithValue(ctx, middlewares.ClusterContextKey, clusterCtxPayload)

		params, err := NewApplicationTokenParams(ctx)
		require.NoError(t, err)

		appTokenParams, ok := params.(ApplicationTokenParams)
		assert.True(t, ok)

		assert.Equal(t, appName, appTokenParams.Application)
		assert.Equal(t, tenant, appTokenParams.Tenant)
		assert.Equal(t, group, appTokenParams.Group)
	})

	t.Run("should fail when there is no ClusterContext", func(t *testing.T) {
		appCtxPayload := middlewares.ApplicationContext{Application: appName}

		ctx := context.WithValue(context.Background(), middlewares.ApplicationContextKey, appCtxPayload)

		_, err := NewApplicationTokenParams(ctx)
		require.Error(t, err)

		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})

	t.Run("should fail when there is no ApplicationContext", func(t *testing.T) {
		clusterCtxPayload := middlewares.ClusterContext{Group: group, Tenant: tenant}

		ctx := context.WithValue(context.Background(), middlewares.ClusterContextKey, clusterCtxPayload)

		_, err := NewApplicationTokenParams(ctx)
		require.Error(t, err)

		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})
}

func TestParams_NewClusterTokenParams(t *testing.T) {
	t.Run("should return ClusterTokenParams", func(t *testing.T) {
		clusterCtxPayload := middlewares.ClusterContext{Group: group, Tenant: tenant}

		ctx := context.WithValue(context.Background(), middlewares.ClusterContextKey, clusterCtxPayload)

		params, err := NewClusterTokenParams(ctx)
		require.NoError(t, err)

		appTokenParams, ok := params.(ClusterTokenParams)
		assert.True(t, ok)

		assert.Equal(t, tenant, appTokenParams.Tenant)
		assert.Equal(t, group, appTokenParams.Group)
	})

	t.Run("should fail when there is no ClusterContext", func(t *testing.T) {
		_, err := NewApplicationTokenParams(context.Background())
		require.Error(t, err)

		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})
}
