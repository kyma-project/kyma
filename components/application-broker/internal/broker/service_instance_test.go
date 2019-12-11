package broker_test

import (
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/cache"
)

func TestServiceInstanceFacadeGetServiceInstance(t *testing.T) {
	// GIVEN
	givenServiceInstance := broker.FixServiceInstance()
	cs := fake.NewSimpleClientset(givenServiceInstance)
	informer := createInformer(cs)
	facade := broker.NewServiceInstanceFacade(informer)
	waitForInformerStart(t, informer)

	// WHEN
	obj, err := facade.GetByNamespaceAndExternalID(givenServiceInstance.Namespace, givenServiceInstance.Spec.ExternalID)

	//THEN
	assert.NoError(t, err)
	assert.Equal(t, givenServiceInstance, obj)
}

func TestServiceInstanceFacadeGetServiceInstanceNotFound(t *testing.T) {
	// GIVEN
	givenServiceInstance := broker.FixServiceInstance()
	cs := fake.NewSimpleClientset(givenServiceInstance)
	informer := createInformer(cs)
	facade := broker.NewServiceInstanceFacade(informer)
	waitForInformerStart(t, informer)

	// WHEN
	obj, err := facade.GetByNamespaceAndExternalID(givenServiceInstance.Namespace, "not-existing")

	//THEN
	assert.Error(t, err)
	assert.Nil(t, obj)
}

func createInformer(cs clientset.Interface) cache.SharedIndexInformer {
	informerFactory := externalversions.NewSharedInformerFactory(cs, 0)
	informer := informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer()

	return informer
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
