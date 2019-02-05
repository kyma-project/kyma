package apicontroller_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/apicontroller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

const testTimes = 3
const informerResyncPeriod = 10 * time.Second

func TestPluggableResolver(t *testing.T) {
	pluggable, err := apicontroller.New(&rest.Config{}, informerResyncPeriod)
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

func checkExportedFields(t *testing.T, resolver *apicontroller.PluggableResolver, enabled bool) {
	assert.NotNil(t, resolver.Resolver)
	namespace := "default"

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	val, err := resolver.Resolver.APIsQuery(ctx, namespace, nil, nil)
	if enabled {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
		require.Nil(t, val)
	}
}
