package assetstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

const testTimes = 3
const informerResyncPeriod = 10 * time.Second

func TestPluggableContainer(t *testing.T) {
	pluggable, err := assetstore.New(&rest.Config{}, informerResyncPeriod)
	require.NoError(t, err)

	pluggable.SetFakeClient()

	for i := 0; i < testTimes; i++ {
		require.NotPanics(t, func() {
			err := pluggable.Enable()
			require.NoError(t, err)
			<-pluggable.Pluggable.SyncCh

			checkExportedFields(t, pluggable, true)
		})
		require.NotPanics(t, func() {
			err := pluggable.Disable()
			require.NoError(t, err)

			checkExportedFields(t, pluggable, false)
		})
	}
}

func checkExportedFields(t *testing.T, resolver *assetstore.PluggableContainer, enabled bool) {
	assert.NotNil(t, resolver.Resolver)
	require.NotNil(t, resolver.AssetStoreRetriever)
	assert.NotNil(t, resolver.AssetStoreRetriever.ClusterAssetGetter)
	assert.NotNil(t, resolver.AssetStoreRetriever.AssetGetter)
	assert.NotNil(t, resolver.AssetStoreRetriever.GqlClusterAssetConverter)
	assert.NotNil(t, resolver.AssetStoreRetriever.GqlAssetConverter)

	if enabled {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	val, err := resolver.Resolver.ClusterAssetFilesField(ctx, nil, nil)
	require.Error(t, err)
	require.Nil(t, val)
}
