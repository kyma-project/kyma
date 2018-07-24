package servicecatalog_test

import (
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetServicePlan(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		planName := "testExample"
		servicePlan := fixServicePlan(planName, "test", planName)
		client := fake.NewSimpleClientset(servicePlan)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc := servicecatalog.NewPlanService(servicePlanInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.Find(planName)
		require.NoError(t, err)
		assert.Equal(t, servicePlan, plan)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc := servicecatalog.NewPlanService(servicePlanInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.Find("doesntExist")
		require.NoError(t, err)
		assert.Nil(t, plan)
	})
}

func TestPlanService_FindByExternalName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		className := "test"
		planName := "testExample"
		externalName := "testExternal"
		servicePlan := fixServicePlan(planName, className, externalName)
		client := fake.NewSimpleClientset(servicePlan)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc := servicecatalog.NewPlanService(servicePlanInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.FindByExternalNameForClass(externalName, className)
		require.NoError(t, err)
		assert.Equal(t, servicePlan, plan)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc := servicecatalog.NewPlanService(servicePlanInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.FindByExternalNameForClass("doesntExist", "none")

		require.NoError(t, err)
		assert.Nil(t, plan)
	})

	t.Run("Error", func(t *testing.T) {
		className := "duplicateName"
		externalName := "duplicateName"
		client := fake.NewSimpleClientset(
			fixServicePlan("1", className, externalName),
			fixServicePlan("2", className, externalName),
			fixServicePlan("3", className, externalName),
		)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc := servicecatalog.NewPlanService(servicePlanInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		_, err := svc.FindByExternalNameForClass(externalName, className)

		assert.Error(t, err)
	})
}

func TestListServicePlansForServiceClass(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		className := "testClassName"

		servicePlan1 := fixServicePlan("1", className, "1")
		servicePlan2 := fixServicePlan("2", className, "2")
		servicePlan3 := fixServicePlan("3", className, "3")
		client := fake.NewSimpleClientset(servicePlan1, servicePlan2, servicePlan3)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc := servicecatalog.NewPlanService(servicePlanInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plans, err := svc.ListForClass(className)
		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.ClusterServicePlan{
			servicePlan1, servicePlan2, servicePlan3,
		}, plans)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()
		svc := servicecatalog.NewPlanService(servicePlanInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		var emptyArray []*v1beta1.ClusterServicePlan
		plans, err := svc.ListForClass("doesntExist")
		require.NoError(t, err)
		assert.Equal(t, emptyArray, plans)
	})

}

func fixServicePlan(name, relatedServiceClassName, externalName string) *v1beta1.ClusterServicePlan {
	plan := v1beta1.ClusterServicePlan{
		Spec: v1beta1.ClusterServicePlanSpec{
			CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
				ExternalName: externalName,
			},
			ClusterServiceClassRef: v1beta1.ClusterObjectReference{
				Name: relatedServiceClassName,
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	return &plan
}
