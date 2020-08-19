package servicecatalog_test

import (
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/application-broker/internal/servicecatalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestServiceBindingFetcherGetServiceBindingSecretName(t *testing.T) {
	// given
	sbA := fixServiceBinding("A")
	sbB := fixServiceBinding("B")
	sbC := fixServiceBinding("C")

	sc := fake.NewSimpleClientset(sbA, sbB, sbC)
	sbInformer := createServiceBindingInformer(sc)
	fetcher := servicecatalog.NewServiceBindingFetcher(sbInformer)

	waitForInformerStart(t, sbInformer)

	// when
	gotSecretName, err := fetcher.GetServiceBindingSecretName(namespace, sbA.Spec.ExternalID)

	// then
	require.NoError(t, err)
	assert.Equal(t, sbA.Spec.SecretName, gotSecretName)
}

func TestServiceBindingFetcherGetServiceBindingSecretNameError(t *testing.T) {
	// given
	sc := fake.NewSimpleClientset(fixServiceBinding("A"))
	sbInformer := createServiceBindingInformer(sc)
	fetcher := servicecatalog.NewServiceBindingFetcher(sbInformer)

	waitForInformerStart(t, sbInformer)

	// when
	_, err := fetcher.GetServiceBindingSecretName(namespace, "not-existing-id")

	// then
	require.EqualError(t, err, "expected to found one Service Binding but got 0")
}

func createServiceBindingInformer(cs clientset.Interface) cache.SharedIndexInformer {
	informerFactory := externalversions.NewSharedInformerFactory(cs, 0)
	informer := informerFactory.Servicecatalog().V1beta1().ServiceBindings().Informer()
	return informer
}

func fixServiceBinding(externalID string) *v1beta1.ServiceBinding {
	return &v1beta1.ServiceBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      "binding" + externalID,
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceBindingSpec{
			SecretName: "secret-name",
			ExternalID: externalID,
		}}
}
