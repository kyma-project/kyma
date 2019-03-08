package content_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimes = 3

func TestPluggableResolver(t *testing.T) {
	pluggable, err := content.New(content.Config{
		Address: "test-content.kyma.local",
		Bucket:  "test",
	})
	require.NoError(t, err)

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

func checkExportedFields(t *testing.T, resolver *content.PluggableContainer, enabled bool) {
	assert.NotNil(t, resolver.Resolver)
	require.NotNil(t, resolver.ContentRetriever)
	assert.NotNil(t, resolver.ContentRetriever.ApiSpec())
	assert.NotNil(t, resolver.ContentRetriever.OpenApiSpec())
	assert.NotNil(t, resolver.ContentRetriever.ODataSpec())
	assert.NotNil(t, resolver.ContentRetriever.AsyncApiSpec())
	assert.NotNil(t, resolver.ContentRetriever.Content())

	if enabled {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	val, err := resolver.Resolver.ContentQuery(ctx, "test", "test")
	require.Error(t, err)
	require.Nil(t, val)
}
