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

func TestServicePlanService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		nsName := "ns"
		planName := "testExample"
		servicePlan := fixServicePlan(planName, "test", planName, nsName)
		client := fake.NewSimpleClientset(&servicePlan)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer()

		svc, err := servicecatalog.NewServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.Find(planName, nsName)
		require.NoError(t, err)
		assert.Equal(t, &servicePlan, plan)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer()

		svc, err := servicecatalog.NewServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.Find("doesntExist", "ns")
		require.NoError(t, err)
		assert.Nil(t, plan)
	})
}

func TestServicePlanService_FindByExternalNameForClass(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		nsName := "ns"
		className := "test"
		planName := "testExample"
		externalName := "testExternal"
		servicePlan := fixServicePlan(planName, className, externalName, nsName)
		client := fake.NewSimpleClientset(&servicePlan)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer()

		svc, err := servicecatalog.NewServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.FindByExternalName(externalName, className, nsName)
		require.NoError(t, err)
		assert.Equal(t, &servicePlan, plan)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer()

		svc, err := servicecatalog.NewServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plan, err := svc.FindByExternalName("doesntExist", "none", "ns")

		require.NoError(t, err)
		assert.Nil(t, plan)
	})

	t.Run("Error", func(t *testing.T) {
		nsName := "ns"
		className := "duplicateName"
		externalName := "duplicateName"

		servicePlan1 := fixServicePlan("1", className, externalName, nsName)
		servicePlan2 := fixServicePlan("2", className, externalName, nsName)
		servicePlan3 := fixServicePlan("3", className, externalName, nsName)
		client := fake.NewSimpleClientset(&servicePlan1, &servicePlan2, &servicePlan3)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer()

		svc, err := servicecatalog.NewServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		_, err = svc.FindByExternalName(externalName, className, nsName)

		assert.Error(t, err)
	})
}

func TestServicePlanService_ListForClass(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		nsName := "ns"
		className := "testClassName"

		servicePlan1 := fixServicePlan("1", className, "1", nsName)
		servicePlan2 := fixServicePlan("2", className, "2", nsName)
		servicePlan3 := fixServicePlan("3", className, "3", nsName)
		client := fake.NewSimpleClientset(&servicePlan1, &servicePlan2, &servicePlan3)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer()

		svc, err := servicecatalog.NewServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		plans, err := svc.ListForServiceClass(className, nsName)
		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1beta1.ServicePlan{
			&servicePlan1, &servicePlan2, &servicePlan3,
		}, plans)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		servicePlanInformer := informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer()
		svc, err := servicecatalog.NewServicePlanService(servicePlanInformer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, servicePlanInformer)

		var emptyArray []*v1beta1.ServicePlan
		plans, err := svc.ListForServiceClass("doesntExist", "ns")
		require.NoError(t, err)
		assert.ElementsMatch(t, emptyArray, plans)
	})
}

func fixServicePlan(name, relatedServiceClassName, externalName, namespace string) v1beta1.ServicePlan {
	plan := v1beta1.ServicePlan{
		Spec: v1beta1.ServicePlanSpec{
			CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
				ExternalName: externalName,
			},
			ServiceClassRef: v1beta1.LocalObjectReference{
				Name: relatedServiceClassName,
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	return plan
}
