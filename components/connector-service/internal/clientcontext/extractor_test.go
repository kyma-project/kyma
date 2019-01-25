package clientcontext

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appName = "appName"
	tenant  = "tenant"
	group   = "group"
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
