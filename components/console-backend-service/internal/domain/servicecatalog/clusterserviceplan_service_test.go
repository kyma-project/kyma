package servicecatalog_test

import (
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterServicePlanService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		planName := "testExample"
		servicePlan := fixClusterServicePlan(planName, "test", planName)
		client := fake.NewSimpleClientset(&servicePlan)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc, err := servicecatalog.NewClusterServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.Find(planName)
		require.NoError(t, err)
		assert.Equal(t, &servicePlan, plan)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc, err := servicecatalog.NewClusterServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.Find("doesntExist")
		require.NoError(t, err)
		assert.Nil(t, plan)
	})
}

func TestClusterServicePlanService_FindByExternalNameForClass(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		className := "test"
		planName := "testExample"
		externalName := "testExternal"
		servicePlan := fixClusterServicePlan(planName, className, externalName)
		client := fake.NewSimpleClientset(&servicePlan)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc, err := servicecatalog.NewClusterServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.FindByExternalName(externalName, className)
		require.NoError(t, err)
		assert.Equal(t, &servicePlan, plan)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc, err := servicecatalog.NewClusterServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.FindByExternalName("doesntExist", "none")

		require.NoError(t, err)
		assert.Nil(t, plan)
	})

	t.Run("Error", func(t *testing.T) {
		className := "duplicateName"
		externalName := "duplicateName"
		servicePlan1 := fixClusterServicePlan("1", className, externalName)
		servicePlan2 := fixClusterServicePlan("2", className, externalName)
		servicePlan3 := fixClusterServicePlan("3", className, externalName)
		client := fake.NewSimpleClientset(
			&servicePlan1, &servicePlan2, &servicePlan3,
		)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc, err := servicecatalog.NewClusterServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		_, err = svc.FindByExternalName(externalName, className)

		assert.Error(t, err)
	})
}

func TestClusterServicePlanService_ListForClass(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		className := "testClassName"

		servicePlan1 := fixClusterServicePlan("1", className, "1")
		servicePlan2 := fixClusterServicePlan("2", className, "2")
		servicePlan3 := fixClusterServicePlan("3", className, "3")
		client := fake.NewSimpleClientset(&servicePlan1, &servicePlan2, &servicePlan3)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()

		svc, err := servicecatalog.NewClusterServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plans, err := svc.ListForClusterServiceClass(className)
		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1beta1.ClusterServicePlan{
			&servicePlan1, &servicePlan2, &servicePlan3,
		}, plans)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer()
		svc, err := servicecatalog.NewClusterServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		var emptyArray []*v1beta1.ClusterServicePlan
		plans, err := svc.ListForClusterServiceClass("doesntExist")
		require.NoError(t, err)
		assert.Equal(t, emptyArray, plans)
	})

}

func fixClusterServicePlan(name, relatedServiceClassName, externalName string) v1beta1.ClusterServicePlan {
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

	return plan
}
