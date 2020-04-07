package servicecatalogaddons_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	api "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBindingUsageServiceCreate(t *testing.T) {
	// GIVEN
	fakeClient, err := newDynamicClient()
	require.NoError(t, err)
	sut, err := servicecatalogaddons.NewServiceBindingUsageService(fakeClient.Resource(bindingUsageGVR), newSbuFakeInformer(fakeClient), nil, "sbu-name")
	require.NoError(t, err)
	// WHEN
	usage := fixBindingUsage()
	_, err = sut.Create("prod", usage)
	// THEN
	require.NoError(t, err)
	actualUsage, err := fakeClient.Resource(bindingUsageGVR).Namespace(usage.Namespace).Get(usage.Name, v1.GetOptions{})
	require.NoError(t, err)
	assert.NotNil(t, actualUsage)
}

func TestBindingUsageServiceCreateWithGeneratedName(t *testing.T) {
	// GIVEN
	fakeClient, err := newDynamicClient()
	require.NoError(t, err)
	sut, err := servicecatalogaddons.NewServiceBindingUsageService(fakeClient.Resource(bindingUsageGVR), newSbuFakeInformer(fakeClient), nil, "generated-sbu-name")
	require.NoError(t, err)
	sbu := fixBindingUsage()
	sbu.Name = ""
	// WHEN
	_, err = sut.Create("prod", sbu)
	// THEN
	require.NoError(t, err)
	actualUsage, err := fakeClient.Resource(bindingUsageGVR).Namespace("prod").Get("generated-sbu-name", v1.GetOptions{})
	require.NoError(t, err)
	assert.NotNil(t, actualUsage)
}

func TestBindingUsageServiceDelete(t *testing.T) {
	// GIVEN
	bindingUsage := fixBindingUsage()
	fakeClient, err := newDynamicClient(bindingUsage)
	require.NoError(t, err)
	sut, err := servicecatalogaddons.NewServiceBindingUsageService(fakeClient.Resource(bindingUsageGVR), newSbuFakeInformer(fakeClient), nil, bindingUsage.Name)
	require.NoError(t, err)
	// WHEN
	err = sut.Delete(bindingUsage.Namespace, bindingUsage.Name)
	// THEN
	require.NoError(t, err)
	_, err = fakeClient.Resource(bindingUsageGVR).Namespace(bindingUsage.Namespace).Get(bindingUsage.Name, v1.GetOptions{})
	require.True(t, apierrors.IsNotFound(err))
}

func TestBindingUsageServiceFind(t *testing.T) {
	// GIVEN
	bindingUsage := fixBindingUsage()
	fakeClient, err := newDynamicClient(bindingUsage)
	require.NoError(t, err)
	informer := newSbuFakeInformer(fakeClient)
	sut, err := servicecatalogaddons.NewServiceBindingUsageService(fakeClient.Resource(bindingUsageGVR), informer, nil, "sbu-name")
	require.NoError(t, err)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	// WHEN
	actual, err := sut.Find(bindingUsage.Namespace, bindingUsage.Name)
	// THEN
	require.NoError(t, err)
	assert.Equal(t, bindingUsage, actual)
}

func TestBindingUsageServiceList(t *testing.T) {
	// GIVEN
	us1 := fixBindingUsage()
	us2 := fixBindingUsage()
	us2.Name = "second-usage"
	fakeClient, err := newDynamicClient(us1, us2)
	require.NoError(t, err)
	informer := newSbuFakeInformer(fakeClient)
	sut, err := servicecatalogaddons.NewServiceBindingUsageService(fakeClient.Resource(bindingUsageGVR), informer, nil, "")
	require.NoError(t, err)
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

	fakeClient, err := newDynamicClient(us1, us2, us3)
	require.NoError(t, err)
	informer := newSbuFakeInformer(fakeClient)
	bindingFinderLister := &automock.ServiceBindingFinderLister{}
	defer bindingFinderLister.AssertExpectations(t)

	bindingFinderLister.On("ListForServiceInstance", "prod", "redis-instance").Return(
		[]*v1beta1.ServiceBinding{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "binding-redis-1",
					Namespace: "prod",
				},
				Spec: v1beta1.ServiceBindingSpec{
					InstanceRef: v1beta1.LocalObjectReference{
						Name: "redis-instance",
					},
				}},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "binding-redis-2",
					Namespace: "prod",
				},
				Spec: v1beta1.ServiceBindingSpec{
					InstanceRef: v1beta1.LocalObjectReference{
						Name: "redis-instance",
					},
				}},
		}, nil)

	scRetriever := &automock.ServiceCatalogRetriever{}
	scRetriever.On("ServiceBinding").Return(bindingFinderLister)

	sut, err := servicecatalogaddons.NewServiceBindingUsageService(fakeClient.Resource(bindingUsageGVR), informer, scRetriever, "sbu-name")
	require.NoError(t, err)
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
		fakeClient, err := newDynamicClient()
		require.NoError(t, err)
		informer := newSbuFakeInformer(fakeClient)
		bindingFinderLister := &automock.ServiceBindingFinderLister{}
		defer bindingFinderLister.AssertExpectations(t)
		bindingFinderLister.On("ListForServiceInstance", mock.Anything, mock.Anything).Return(nil, errors.New("some error"))

		scRetriever := &automock.ServiceCatalogRetriever{}
		scRetriever.On("ServiceBinding").Return(bindingFinderLister)

		sut, err := servicecatalogaddons.NewServiceBindingUsageService(fakeClient.Resource(bindingUsageGVR), informer, scRetriever, "sbu-name")
		require.NoError(t, err)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
		// WHEN
		_, err = sut.ListForServiceInstance("prod", "redis-instance")
		// THEN
		assert.EqualError(t, err, "while getting ServiceBindings for instance [namespace: prod, name: redis-instance]: some error")
	})
}

func TestBindingUsageServiceListForDeployment(t *testing.T) {
	// GIVEN
	us1 := customBindingUsage("redis-1")
	us2 := customBindingUsage("redis-2")
	us3 := customFunctionBindingUsage("mysql-1")

	fakeClient, err := newDynamicClient(us1, us2, us3)
	require.NoError(t, err)
	informer := newSbuFakeInformer(fakeClient)
	sut, err := servicecatalogaddons.NewServiceBindingUsageService(fakeClient.Resource(bindingUsageGVR), informer, nil, "sbu-name")
	require.NoError(t, err)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	// WHEN
	usages, err := sut.ListByUsageKind("prod", "deployment", "app")
	// THEN
	require.NoError(t, err)
	assert.Len(t, usages, 2)
	assert.Contains(t, usages, us1)
	assert.Contains(t, usages, us2)

}

func fixBindingUsage() *api.ServiceBindingUsage {
	return &api.ServiceBindingUsage{
		TypeMeta: v1.TypeMeta{
			APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
			Kind:       "servicebindingusage",
		},
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

func customBindingUsage(id string) *api.ServiceBindingUsage {
	return &api.ServiceBindingUsage{
		TypeMeta: v1.TypeMeta{
			APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
			Kind:       "servicebindingusage",
		},
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
