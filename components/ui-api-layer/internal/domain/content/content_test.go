package content_test

import (
	"testing"

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

			checkExportedFields(t, pluggable)
		})
		require.NotPanics(t, func() {
			err := pluggable.Disable()
			require.NoError(t, err)

			checkExportedFields(t, pluggable)
		})
	}
}

func checkExportedFields(t *testing.T, resolver *content.PluggableContainer) {
	assert.NotNil(t, resolver.Resolver)
	assert.NotNil(t, resolver.ApiSpecGetter)
	assert.NotNil(t, resolver.AsyncApiSpecGetter)
	assert.NotNil(t, resolver.ContentGetter)
}
