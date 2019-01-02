package servicecatalog_test

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"testing"
	"time"
)

const testTimes = 3
const informerResyncPeriod = 10 * time.Second

func TestPluggableContainer(t *testing.T) {
	pluggable, err := servicecatalog.New(&rest.Config{}, informerResyncPeriod, nil, nil, nil)
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

func checkExportedFields(t *testing.T, resolver *servicecatalog.PluggableContainer) {
	assert.NotNil(t, resolver.Resolver)
	assert.NotNil(t, resolver.ServiceBindingUsageLister)
	assert.NotNil(t, resolver.ServiceBindingGetter)
}

