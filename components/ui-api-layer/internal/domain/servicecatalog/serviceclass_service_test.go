package servicecatalog_test

import (
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceClassService_GetServiceClass(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		environmentName := "env"
		className := "testExample"
		serviceClass := fixServiceClass(className, className, environmentName)
		client := fake.NewSimpleClientset(serviceClass)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer()

		svc := servicecatalog.NewServiceClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		class, err := svc.Find(className, environmentName)
		require.NoError(t, err)
		assert.Equal(t, serviceClass, class)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer()

		svc := servicecatalog.NewServiceClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		class, err := svc.Find("doesntExist", "env")

		require.NoError(t, err)
		assert.Nil(t, class)
	})
}

func TestServiceClassService_FindByExternalName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		className := "testExample"
		externalName := "testExternal"
		environmentName := "exampleEnv"
		serviceClass := fixServiceClass(className, externalName, environmentName)
		client := fake.NewSimpleClientset(serviceClass)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer()

		svc := servicecatalog.NewServiceClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		class, err := svc.FindByExternalName(externalName, environmentName)
		require.NoError(t, err)
		assert.Equal(t, serviceClass, class)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer()

		svc := servicecatalog.NewServiceClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		class, err := svc.FindByExternalName("doesntExist", "env")

		require.NoError(t, err)
		assert.Nil(t, class)
	})

	t.Run("Error", func(t *testing.T) {
		environmentName := "env"
		externalName := "duplicateName"
		client := fake.NewSimpleClientset(
			fixServiceClass("1", externalName, environmentName),
			fixServiceClass("2", externalName, environmentName),
			fixServiceClass("3", externalName, environmentName),
		)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer()

		svc := servicecatalog.NewServiceClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		_, err := svc.FindByExternalName(externalName, environmentName)

		assert.Error(t, err)
	})
}

func TestServiceClassService_ListServiceClasses(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		environmentName := "exampleEnv"
		serviceClass1 := fixServiceClass("1", "1", environmentName)
		serviceClass2 := fixServiceClass("2", "2", environmentName)
		serviceClass3 := fixServiceClass("3", "3", environmentName)
		client := fake.NewSimpleClientset(serviceClass1, serviceClass2, serviceClass3)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer()

		svc := servicecatalog.NewServiceClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		classes, err := svc.List(environmentName, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.ServiceClass{
			serviceClass1, serviceClass2, serviceClass3,
		}, classes)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer()

		svc := servicecatalog.NewServiceClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		var emptyArray []*v1beta1.ServiceClass
		classes, err := svc.List("env", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, classes)
	})
}

func fixServiceClass(name, externalName, environment string) *v1beta1.ServiceClass {
	class := v1beta1.ServiceClass{
		Spec: v1beta1.ServiceClassSpec{
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalName: externalName,
				Tags:         []string{"tag1", "tag2"},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: environment,
		},
	}
	return &class
}
