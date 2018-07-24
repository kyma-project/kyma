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

func TestClassService_GetServiceClass(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		className := "testExample"
		serviceClass := fixServiceClass(className, className)
		client := fake.NewSimpleClientset(serviceClass)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer()

		svc := servicecatalog.NewClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		class, err := svc.Find(className)
		require.NoError(t, err)
		assert.Equal(t, serviceClass, class)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer()

		svc := servicecatalog.NewClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		class, err := svc.Find("doesntExist")

		require.NoError(t, err)
		assert.Nil(t, class)
	})
}

func TestClassService_FindByExternalName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		className := "testExample"
		externalName := "testExternal"
		serviceClass := fixServiceClass(className, externalName)
		client := fake.NewSimpleClientset(serviceClass)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer()

		svc := servicecatalog.NewClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		class, err := svc.FindByExternalName(externalName)
		require.NoError(t, err)
		assert.Equal(t, serviceClass, class)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer()

		svc := servicecatalog.NewClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		class, err := svc.FindByExternalName("doesntExist")

		require.NoError(t, err)
		assert.Nil(t, class)
	})

	t.Run("Error", func(t *testing.T) {
		externalName := "duplicateName"
		client := fake.NewSimpleClientset(
			fixServiceClass("1", externalName),
			fixServiceClass("2", externalName),
			fixServiceClass("3", externalName),
		)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer()

		svc := servicecatalog.NewClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		_, err := svc.FindByExternalName(externalName)

		assert.Error(t, err)
	})
}

func TestClassService_ListServiceClasses(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		serviceClass1 := fixServiceClass("1", "1")
		serviceClass2 := fixServiceClass("2", "2")
		serviceClass3 := fixServiceClass("3", "3")
		client := fake.NewSimpleClientset(serviceClass1, serviceClass2, serviceClass3)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer()

		svc := servicecatalog.NewClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		classes, err := svc.List(pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.ClusterServiceClass{
			serviceClass1, serviceClass2, serviceClass3,
		}, classes)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceClassInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer()

		svc := servicecatalog.NewClassService(serviceClassInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceClassInformer)

		var emptyArray []*v1beta1.ClusterServiceClass
		classes, err := svc.List(pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, classes)
	})
}

func fixServiceClass(name, externalName string) *v1beta1.ClusterServiceClass {
	class := v1beta1.ClusterServiceClass{
		Spec: v1beta1.ClusterServiceClassSpec{
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalName: externalName,
				Tags:         []string{"tag1", "tag2"},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &class
}
