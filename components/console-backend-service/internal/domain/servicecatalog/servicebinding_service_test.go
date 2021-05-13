package servicecatalog_test

import (
	"context"
	"testing"
	"time"

	api "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestBindingServiceCreate(t *testing.T) {
	// GIVEN
	fakeClient := fake.NewSimpleClientset()
	sut, err := servicecatalog.NewServiceBindingService(fakeClient.ServicecatalogV1beta1(), fixBindingInformer(fakeClient), "sb-generated-name")
	require.NoError(t, err)
	// WHEN
	actualBinding, err := sut.Create("production", fixServiceBindingToRedis())
	// THEN
	require.NoError(t, err)
	bindingFromClientSet, err := fakeClient.ServicecatalogV1beta1().ServiceBindings("production").Get(context.Background(), "redis-binding", v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, bindingFromClientSet, actualBinding)
}

func TestBindingServiceCreateWithGeneratedName(t *testing.T) {
	// GIVEN
	fakeClient := fake.NewSimpleClientset()
	sut, err := servicecatalog.NewServiceBindingService(fakeClient.ServicecatalogV1beta1(), fixBindingInformer(fakeClient), "sb-generated-name")
	require.NoError(t, err)
	sb := fixServiceBindingToRedis()
	sb.Name = ""

	// WHEN
	actualBinding, err := sut.Create("production", sb)
	// THEN
	require.NoError(t, err)
	bindingFromClientSet, err := fakeClient.ServicecatalogV1beta1().ServiceBindings("production").Get(context.Background(), "sb-generated-name", v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, bindingFromClientSet, actualBinding)
}

func TestBindingServiceDelete(t *testing.T) {
	// GIVEN
	fakeClient := fake.NewSimpleClientset(fixServiceBindingToRedis())
	sut, err := servicecatalog.NewServiceBindingService(fakeClient.ServicecatalogV1beta1(), fixBindingInformer(fakeClient), "")
	require.NoError(t, err)

	// WHEN
	err = sut.Delete("production", "redis-binding")
	// THEN
	require.NoError(t, err)
	_, err = fakeClient.ServicecatalogV1beta1().ServiceBindings("production").Get(context.Background(), "redis-binding", v1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err))

}

func TestBindingServiceFind(t *testing.T) {
	// GIVEN
	client := fake.NewSimpleClientset(fixServiceBindingToRedis())
	informer := fixBindingInformer(client)
	sut, err := servicecatalog.NewServiceBindingService(client.ServicecatalogV1beta1(), informer, "")
	require.NoError(t, err)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	// WHEN
	actual, err := sut.Find("production", "redis-binding")
	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixServiceBindingToRedis(), actual)
}

func TestBindingServiceListForServiceInstance(t *testing.T) {
	// GIVEN
	client := fake.NewSimpleClientset(fixServiceBindingToRedis(), fixServiceBindingToSql())
	informer := fixBindingInformer(client)
	sut, err := servicecatalog.NewServiceBindingService(client.ServicecatalogV1beta1(), informer, "")
	require.NoError(t, err)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	// WHEN
	actualBindings, err := sut.ListForServiceInstance("production", "redis")
	// THEN
	require.NoError(t, err)
	assert.Len(t, actualBindings, 1)
	assert.Contains(t, actualBindings, fixServiceBindingToRedis())
}

func fixServiceBindingToRedis() *api.ServiceBinding {
	return &api.ServiceBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      "redis-binding",
			Namespace: "production",
		},
		Spec: api.ServiceBindingSpec{
			InstanceRef: api.LocalObjectReference{
				Name: "redis",
			},
		},
	}
}

func fixServiceBindingToSql() *api.ServiceBinding {
	return &api.ServiceBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      "sql-binding",
			Namespace: "production",
		},
		Spec: api.ServiceBindingSpec{
			InstanceRef: api.LocalObjectReference{
				Name: "sql",
			},
		},
	}
}

func fixBindingInformer(client *fake.Clientset) cache.SharedIndexInformer {
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Servicecatalog().V1beta1().ServiceBindings().Informer()

	return informer
}
