package rafter_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimes = 3
const informerResyncPeriod = 30 * time.Second

func TestPluggableContainer(t *testing.T) {
	svcFactory := fake.NewSimpleFakeServiceFactory(informerResyncPeriod)
	require.NotNil(t, svcFactory)

	pluggable, err := rafter.New(svcFactory, rafter.Config{})
	require.NoError(t, err)

	for i := 0; i < testTimes; i++ {
		require.NotPanics(t, func() {
			err := pluggable.Enable()
			require.NoError(t, err)
			<-pluggable.Pluggable.SyncCh

			checkInternalMethod(t, pluggable, true)
			checkExportedMethods(t, pluggable, true)
		})
		require.NotPanics(t, func() {
			err := pluggable.Disable()
			require.NoError(t, err)

			checkInternalMethod(t, pluggable, false)
			checkExportedMethods(t, pluggable, false)
		})
	}
}

func checkInternalMethod(t *testing.T, resolver *rafter.PluggableContainer, enabled bool) {
	assert.NotNil(t, resolver.Resolver)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	val, err := resolver.Resolver.ClusterAssetGroupsQuery(ctx, nil, nil)
	if enabled {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
		require.Nil(t, val)
	}
}

func checkExportedMethods(t *testing.T, resolver *rafter.PluggableContainer, enabled bool) {
	assert.NotNil(t, resolver.Retriever)
	assert.NotNil(t, resolver.Retriever.ClusterAssetGroupGetter)
	assert.NotNil(t, resolver.Retriever.GqlClusterAssetGroupConverter)
	assert.NotNil(t, resolver.Retriever.AssetGroupGetter)
	assert.NotNil(t, resolver.Retriever.GqlAssetGroupConverter)
	assert.NotNil(t, resolver.Retriever.ClusterAssetGetter)
	assert.NotNil(t, resolver.Retriever.SpecificationSvc)

	if enabled {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	val, err := resolver.Resolver.ClusterAssetFilesField(ctx, nil, nil)
	require.Error(t, err)
	require.Nil(t, val)
}
