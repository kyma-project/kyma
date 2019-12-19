package servicecatalog_test

import (
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/kyma-project/kyma/components/application-broker/internal/servicecatalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const namespace = "working-ns"

func TestServiceInstanceFacadeGetServiceInstance(t *testing.T) {
	// GIVEN
	givenServiceInstance := fixServiceInstance()
	cs := fake.NewSimpleClientset(givenServiceInstance)
	informer := createServiceInstanceInformer(cs)
	facade := servicecatalog.NewFacade(informer, nil)
	waitForInformerStart(t, informer)

	// WHEN
	obj, err := facade.GetByNamespaceAndExternalID(givenServiceInstance.Namespace, givenServiceInstance.Spec.ExternalID)

	//THEN
	assert.NoError(t, err)
	assert.Equal(t, givenServiceInstance, obj)
}

func TestServiceInstanceFacadeGetServiceInstanceNotFound(t *testing.T) {
	// GIVEN
	givenServiceInstance := fixServiceInstance()
	cs := fake.NewSimpleClientset(givenServiceInstance)
	informer := createServiceInstanceInformer(cs)
	facade := servicecatalog.NewFacade(informer, nil)
	waitForInformerStart(t, informer)

	// WHEN
	obj, err := facade.GetByNamespaceAndExternalID(givenServiceInstance.Namespace, "not-existing")

	//THEN
	assert.Error(t, err)
	assert.Nil(t, obj)
}

func TestFacadeExistingAppInstance(t *testing.T) {
	// given
	cs := fake.NewSimpleClientset(fixAppServiceClass(), fixServiceClass(), fixAppServiceInstnace())
	siInformer, scInformer := createInformers(t, cs)
	facade := servicecatalog.NewFacade(siInformer, scInformer)

	// when / then
	exists, err := facade.AnyServiceInstanceExists(namespace)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFacadeNoInstances(t *testing.T) {
	// given
	cs := fake.NewSimpleClientset(fixAppServiceClass(), fixServiceClass())
	siInformer, scInformer := createInformers(t, cs)
	facade := servicecatalog.NewFacade(siInformer, scInformer)

	// when / then
	exists, err := facade.AnyServiceInstanceExists(namespace)

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFacadeNoAppInstances(t *testing.T) {
	// given
	cs := fake.NewSimpleClientset(fixAppServiceClass(), fixServiceClass(), fixServiceInstnaceClusterServiceClass(), fixServiceInstance())
	siInformer, scInformer := createInformers(t, cs)
	facade := servicecatalog.NewFacade(siInformer, scInformer)

	// when / then
	exists, err := facade.AnyServiceInstanceExists(namespace)

	require.NoError(t, err)
	assert.False(t, exists)
}

func createInformers(t *testing.T, cs *fake.Clientset) (cache.SharedIndexInformer, cache.SharedIndexInformer) {
	siInformer := createServiceInstanceInformer(cs)
	scInformer := createServiceClassInformer(cs)
	waitForInformerStart(t, siInformer)
	waitForInformerStart(t, scInformer)
	return siInformer, scInformer
}

func fixAppServiceInstnace() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:      "si-application01",
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			ServiceClassRef: &v1beta1.LocalObjectReference{
				Name: "app-class",
			},
		}}
}

func fixServiceInstance() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:      "si-application01",
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			ServiceClassRef: &v1beta1.LocalObjectReference{
				Name: "sc-service",
			},
		}}
}

func fixServiceInstnaceClusterServiceClass() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:      "si-clusterserviceclass",
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
				Name: "some-class",
			},
		}}
}

func fixAppServiceClass() *v1beta1.ServiceClass {
	return &v1beta1.ServiceClass{
		ObjectMeta: v1.ObjectMeta{
			Name:      "app-class",
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceClassSpec{
			ServiceBrokerName: nsbroker.NamespacedBrokerName,
		}}
}

func fixServiceClass() *v1beta1.ServiceClass {
	return &v1beta1.ServiceClass{
		ObjectMeta: v1.ObjectMeta{
			Name:      "sc-service",
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceClassSpec{
			ServiceBrokerName: "other-broker",
		}}
}

func createServiceInstanceInformer(cs clientset.Interface) cache.SharedIndexInformer {
	informerFactory := externalversions.NewSharedInformerFactory(cs, 0)
	informer := informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer()
	return informer
}

func createServiceClassInformer(cs clientset.Interface) cache.SharedIndexInformer {
	informerFactory := externalversions.NewSharedInformerFactory(cs, 0)
	informer := informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer()
	return informer
}

func waitForInformerStart(t *testing.T, informer cache.SharedIndexInformer) {
	stop := make(chan struct{})
	syncedDone := make(chan struct{})

	go func() {
		if !cache.WaitForCacheSync(stop, informer.HasSynced) {
			t.Fatalf("timeout occurred when waiting to sync siInformer")
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
