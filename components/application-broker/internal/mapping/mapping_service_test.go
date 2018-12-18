package mapping

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestMappingService_ListEnvironmentMappings(t *testing.T) {
	// GIVEN
	em1 := fixEM("em1", "prod")
	em2 := fixEM("em2", "prod")
	emQa := fixEM("em3", "qa")
	client := fake.NewSimpleClientset(em1, em2, emQa)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	emInformer := informerFactory.Applicationconnector().V1alpha1().EnvironmentMappings().Informer()

	waitForInformerStart(t, emInformer)
	svc := newMappingService(emInformer)

	// WHEN
	result, err := svc.ListEnvironmentMappings("prod")

	// THEN
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, em1)
	assert.Contains(t, result, em2)
}

func fixEM(name string, namespace string) *v1alpha1.EnvironmentMapping {
	return &v1alpha1.EnvironmentMapping{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func waitForInformerStart(t *testing.T, informer cache.SharedIndexInformer) {
	stop := make(chan struct{})
	syncedDone := make(chan struct{})

	go func() {
		if !cache.WaitForCacheSync(stop, informer.HasSynced) {
			t.Fatalf("timeout occurred when waiting to sync informer")
		}
		close(syncedDone)
	}()

	go informer.Run(stop)

	select {
	case <-time.After(time.Second):
		close(stop)
	case <-syncedDone:
	}
}
