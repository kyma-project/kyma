package eventing

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimes = 3

func TestPluggableContainer(t *testing.T) {
	svcFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
	require.NoError(t, err)
	require.NotNil(t, svcFactory)

	pluggable, err := New(svcFactory)
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

func checkInternalMethod(t *testing.T, resolver *PluggableContainer, enabled bool) {
	assert.NotNil(t, resolver.Resolver)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	val, err := resolver.Resolver.CreateTrigger(ctx, fixTrigger(), []gqlschema.OwnerReference{})
	if enabled {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
		require.Nil(t, val)
	}
}

func checkExportedMethods(t *testing.T, resolver *PluggableContainer, enabled bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	val, err := resolver.Resolver.DeleteTrigger(ctx, gqlschema.TriggerMetadataInput{Name: "Name", Namespace: "Namespace"})
	if enabled {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
		require.Nil(t, val)
	}
}

func fixTrigger() gqlschema.TriggerCreateInput {
	uri := "www.test.com"
	name := "Name"
	return gqlschema.TriggerCreateInput{
		Name:      &name,
		Namespace: "Namespace",
		Broker:    "default",
		Subscriber: gqlschema.SubscriberInput{
			URI: &uri,
		},
	}
}
