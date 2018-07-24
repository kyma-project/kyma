package servicecatalog_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestBindingUsageServiceCreate(t *testing.T) {
	// GIVEN
	fakeClient := fake.NewSimpleClientset()
	sut := servicecatalog.NewServiceBindingUsageService(fakeClient.ServicecatalogV1alpha1(), fixBindingUsageInformer(fakeClient), nil)
	// WHEN
	_, err := sut.Create("prod", fixBindingUsage())
	// THEN
	require.NoError(t, err)
	actualUsage, err := fakeClient.ServicecatalogV1alpha1().ServiceBindingUsages("prod").Get("usage", v1.GetOptions{})
	require.NoError(t, err)
	assert.NotNil(t, actualUsage)

}

func TestBindingUsageServiceDelete(t *testing.T) {
	// GIVEN
	fakeClient := fake.NewSimpleClientset(fixBindingUsage())
	sut := servicecatalog.NewServiceBindingUsageService(fakeClient.ServicecatalogV1alpha1(), fixBindingUsageInformer(fakeClient), nil)
	// WHEN
	err := sut.Delete("prod", "usage")
	// THEN
	require.NoError(t, err)
	_, err = fakeClient.ServicecatalogV1alpha1().ServiceBindingUsages("prod").Get("usage", v1.GetOptions{})
	require.True(t, apierrors.IsNotFound(err))
}

func TestBindingUsageServiceFind(t *testing.T) {
	// GIVEN
	fakeClient := fake.NewSimpleClientset(fixBindingUsage())
	informer := fixBindingUsageInformer(fakeClient)
	sut := servicecatalog.NewServiceBindingUsageService(fakeClient.ServicecatalogV1alpha1(), informer, nil)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	// WHEN
	actual, err := sut.Find("prod", "usage")
	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixBindingUsage(), actual)
}

func TestBindingUsageServiceList(t *testing.T) {
	// GIVEN
	us1 := fixBindingUsage()
	us2 := fixBindingUsage()
	us2.Name = "second-usage"
	fakeClient := fake.NewSimpleClientset(us1, us2)
	informer := fixBindingUsageInformer(fakeClient)
	sut := servicecatalog.NewServiceBindingUsageService(fakeClient.ServicecatalogV1alpha1(), informer, nil)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	// WHEN
	actualUsages, err := sut.List("prod")
	// THEN
	require.NoError(t, err)
	assert.Len(t, actualUsages, 2)
	assert.Contains(t, actualUsages, us1)
	assert.Contains(t, actualUsages, us2)

}

func TestBindingUsageServiceListForServiceInstance(t *testing.T) {
	// GIVEN
	us1 := customBindingUsage("redis-1")
	us2 := customBindingUsage("redis-2")
	us3 := customBindingUsage("mysql-1")

	fakeClient := fake.NewSimpleClientset(us1, us2, us3)
	informer := fixBindingUsageInformer(fakeClient)
	mockBindingFacade := automock.NewServiceBindingOperations()
	defer mockBindingFacade.AssertExpectations(t)

	mockBindingFacade.On("ListForServiceInstance", "prod", "redis-instance").Return(
		[]*v1beta1.ServiceBinding{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "binding-redis-1",
					Namespace: "prod",
				},
				Spec: v1beta1.ServiceBindingSpec{
					ServiceInstanceRef: v1beta1.LocalObjectReference{
						Name: "redis-instance",
					},
				}},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "binding-redis-2",
					Namespace: "prod",
				},
				Spec: v1beta1.ServiceBindingSpec{
					ServiceInstanceRef: v1beta1.LocalObjectReference{
						Name: "redis-instance",
					},
				}},
		}, nil)

	sut := servicecatalog.NewServiceBindingUsageService(fakeClient.ServicecatalogV1alpha1(), informer, mockBindingFacade)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	// WHEN
	usages, err := sut.ListForServiceInstance("prod", "redis-instance")
	// THEN
	require.NoError(t, err)
	assert.Len(t, usages, 2)
	assert.Contains(t, usages, us1)
	assert.Contains(t, usages, us2)
}

func TestBindingUsageServiceListForServiceInstanceErrors(t *testing.T) {
	t.Run("on getting bindings", func(t *testing.T) {
		// GIVEN
		fakeClient := fake.NewSimpleClientset()
		informer := fixBindingUsageInformer(fakeClient)
		mockBindingFacade := automock.NewServiceBindingOperations()
		defer mockBindingFacade.AssertExpectations(t)
		mockBindingFacade.On("ListForServiceInstance", mock.Anything, mock.Anything).Return(nil, errors.New("some error"))
		sut := servicecatalog.NewServiceBindingUsageService(fakeClient.ServicecatalogV1alpha1(), informer, mockBindingFacade)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
		// WHEN
		_, err := sut.ListForServiceInstance("prod", "redis-instance")
		// THEN
		assert.EqualError(t, err, "while getting ServiceBindings for instance [env: prod, name: redis-instance]: some error")
	})
}

func TestBindingUsageServiceListForDeployment(t *testing.T) {
	// GIVEN
	us1 := customBindingUsage("redis-1")
	us2 := customBindingUsage("redis-2")
	us3 := customFunctionBindingUsage("mysql-1")

	fakeClient := fake.NewSimpleClientset(us1, us2, us3)
	informer := fixBindingUsageInformer(fakeClient)
	sut := servicecatalog.NewServiceBindingUsageService(fakeClient.ServicecatalogV1alpha1(), informer, nil)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	// WHEN
	usages, err := sut.ListForDeployment("prod", "deployment", "app")
	// THEN
	require.NoError(t, err)
	assert.Len(t, usages, 2)
	assert.Contains(t, usages, us1)
	assert.Contains(t, usages, us2)

}

func fixBindingUsage() *api.ServiceBindingUsage {
	return &api.ServiceBindingUsage{
		ObjectMeta: v1.ObjectMeta{
			Name:      "usage",
			Namespace: "prod",
		},
		Spec: api.ServiceBindingUsageSpec{
			UsedBy: api.LocalReferenceByKindAndName{
				Kind: "deployment",
				Name: "app",
			},
			ServiceBindingRef: api.LocalReferenceByName{
				Name: "binding",
			},
		},
	}
}

func fixBindingUsageInformer(client *fake.Clientset) cache.SharedIndexInformer {
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Servicecatalog().V1alpha1().ServiceBindingUsages().Informer()

	return informer
}

func customBindingUsage(id string) *api.ServiceBindingUsage {
	return &api.ServiceBindingUsage{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("usage-%s", id),
			Namespace: "prod",
		},
		Spec: api.ServiceBindingUsageSpec{
			UsedBy: api.LocalReferenceByKindAndName{
				Kind: "deployment",
				Name: "app",
			},
			ServiceBindingRef: api.LocalReferenceByName{
				Name: fmt.Sprintf("binding-%s", id),
			},
		},
	}
}

func customFunctionBindingUsage(id string) *api.ServiceBindingUsage {
	usage := customBindingUsage(id)
	usage.Spec.UsedBy.Kind = "function"

	return usage
}
