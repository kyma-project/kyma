package apicontroller_test

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/apicontroller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"testing"
	"time"
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
			<- pluggable.Pluggable.SyncCh

			checkExportedFields(t, pluggable)
		})
		require.NotPanics(t, func() {
			err := pluggable.Disable()
			require.NoError(t, err)

			checkExportedFields(t, pluggable)
		})
	}
}

func checkExportedFields(t *testing.T, resolver *apicontroller.PluggableResolver) {
	assert.NotNil(t, resolver.Resolver)
}
