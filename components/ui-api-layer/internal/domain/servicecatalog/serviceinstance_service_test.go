package servicecatalog_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	listener "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

func TestServiceInstanceService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instanceName := "testExample"
		environment := "testEnv"

		serviceInstance := fixServiceInstance(instanceName, environment)
		serviceInstanceInformer := fixInformer(serviceInstance)

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		instance, err := svc.Find(instanceName, environment)
		require.NoError(t, err)
		assert.Equal(t, serviceInstance, instance)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceInstanceInformer := informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer()

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		instance, err := svc.Find("doesntExist", "notExistingEnv")
		require.NoError(t, err)
		assert.Nil(t, instance)
	})
}

func TestServiceInstanceService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		environment := "env1"
		serviceInstance1 := fixServiceInstance("1", environment)
		serviceInstance2 := fixServiceInstance("2", environment)
		serviceInstance3 := fixServiceInstance("3", "env2")

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		instances, err := svc.List(environment, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.ServiceInstance{
			serviceInstance1, serviceInstance2,
		}, instances)
	})

	t.Run("NotFound", func(t *testing.T) {
		serviceInstanceInformer := fixInformer()

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		var emptyArray []*v1beta1.ServiceInstance
		instances, err := svc.List("notExisting", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, instances)
	})
}

func TestServiceInstanceService_ListForStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		environment := "env1"
		status := status.ServiceInstanceStatusTypeRunning

		serviceInstance1 := fixServiceInstanceWithStatus("1", environment)
		serviceInstance2 := fixServiceInstanceWithStatus("2", environment)
		serviceInstance3 := fixServiceInstanceWithStatus("3", "env2")

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		instances, err := svc.ListForStatus(environment, pager.PagingParams{}, &status)
		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.ServiceInstance{
			serviceInstance1, serviceInstance2,
		}, instances)
	})

	t.Run("NotFound", func(t *testing.T) {
		status := status.ServiceInstanceStatusTypeRunning

		serviceInstanceInformer := fixInformer()

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		var emptyArray []*v1beta1.ServiceInstance
		instances, err := svc.ListForStatus("notExisting", pager.PagingParams{}, &status)
		require.NoError(t, err)
		assert.Equal(t, emptyArray, instances)
	})
}

func TestServiceInstanceService_ListForClusterServiceClass(t *testing.T) {
	t.Run("ServiceInstancesQuery exist", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		environment := "env1"
		serviceInstance1 := fixServiceInstanceWithClusterPlanRef("1", environment, className, "")
		serviceInstance2 := fixServiceInstanceWithClusterPlanRef("2", environment, "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithClusterPlanRef("3", "env2", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)
		expected := []*v1beta1.ServiceInstance{
			serviceInstance1, serviceInstance2, serviceInstance3,
		}

		instances, err := svc.ListForClusterServiceClass(className, externalClassName)
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, instances)
	})

	t.Run("ServiceInstancesQuery don't exist", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		testClassName := "test"
		testExternalClassName := "test"

		serviceInstance1 := fixServiceInstanceWithClusterPlanRef("1", "env1", className, "")
		serviceInstance2 := fixServiceInstanceWithClusterPlanRef("2", "env2", "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithClusterPlanRef("3", "env3", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		var emptyArray []*v1beta1.ServiceInstance
		instances, err := svc.ListForClusterServiceClass(testClassName, testExternalClassName)
		require.NoError(t, err)
		assert.Equal(t, emptyArray, instances)
	})
}

func TestServiceInstanceService_ListForClass(t *testing.T) {
	t.Run("ServiceInstancesQuery exist", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		environment := "env1"
		serviceInstance1 := fixServiceInstanceWithPlanRef("1", environment, className, "")
		serviceInstance2 := fixServiceInstanceWithPlanRef("2", environment, "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithPlanRef("3", "env2", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)
		expected := []*v1beta1.ServiceInstance{
			serviceInstance1, serviceInstance2, serviceInstance3,
		}

		instances, err := svc.ListForServiceClass(className, externalClassName, environment)
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, instances)
	})

	t.Run("ServiceInstancesQuery don't exist", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		testClassName := "test"
		testExternalClassName := "test"

		serviceInstance1 := fixServiceInstanceWithPlanRef("1", "env1", className, "")
		serviceInstance2 := fixServiceInstanceWithPlanRef("2", "env2", "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithPlanRef("3", "env3", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		var emptyArray []*v1beta1.ServiceInstance
		instances, err := svc.ListForClusterServiceClass(testClassName, testExternalClassName)
		require.NoError(t, err)
		assert.Equal(t, emptyArray, instances)
	})
}

func TestServiceInstanceService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := fixServiceInstance("test", "test")
		client := fake.NewSimpleClientset(expected)
		client.PrependReactor("*", "*", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
			return true, expected, nil
		})

		svc := servicecatalog.NewServiceInstanceService(fixInformer(), client)

		params := servicecatalog.NewServiceInstanceCreateParameters("name", "environment", []string{"test", "label"}, "planName", true, "className", true, nil)
		result, err := svc.Create(*params)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

}

func TestServiceInstanceService_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := fixServiceInstance("test", "test")
		client := fake.NewSimpleClientset(instance)
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), client)

		err := svc.Delete("test", "test")

		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		testErr := fmt.Errorf("Test")
		instance := fixServiceInstance("test", "test")
		client := fake.NewSimpleClientset(instance)
		client.PrependReactor("*", "*", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, testErr
		})
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), client)

		err := svc.Delete("test", "test")

		assert.Equal(t, testErr, err)
	})
}

func TestServiceInstanceService_IsBindableWithClusterRefs(t *testing.T) {
	trueVal := true
	falseVal := false

	for _, tc := range []struct {
		classBindable bool
		planBindable  *bool
		expected      bool
	}{
		{true, &trueVal, true},
		{false, &falseVal, false},
		{true, nil, true},
		{false, nil, false},
		{true, &falseVal, false},
		{false, &trueVal, true},
	} {
		class := &v1beta1.ClusterServiceClass{
			Spec: v1beta1.ClusterServiceClassSpec{
				CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
					Bindable: tc.classBindable,
				},
			},
		}
		plan := &v1beta1.ClusterServicePlan{
			Spec: v1beta1.ClusterServicePlanSpec{
				CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
					Bindable: tc.planBindable,
				},
			},
		}
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)

		result := svc.IsBindableWithClusterRefs(class, plan)

		assert.Equal(t, tc.expected, result)
	}
}

func TestServiceInstanceService_IsBindableWithLocalRefs(t *testing.T) {
	trueVal := true
	falseVal := false

	for _, tc := range []struct {
		classBindable bool
		planBindable  *bool
		expected      bool
	}{
		{true, &trueVal, true},
		{false, &falseVal, false},
		{true, nil, true},
		{false, nil, false},
		{true, &falseVal, false},
		{false, &trueVal, true},
	} {
		class := &v1beta1.ServiceClass{
			Spec: v1beta1.ServiceClassSpec{
				CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
					Bindable: tc.classBindable,
				},
			},
		}
		plan := &v1beta1.ServicePlan{
			Spec: v1beta1.ServicePlanSpec{
				CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
					Bindable: tc.planBindable,
				},
			},
		}
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)

		result := svc.IsBindableWithLocalRefs(class, plan)

		assert.Equal(t, tc.expected, result)
	}
}

func TestServiceInstanceService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		instanceListener := listener.NewInstance(nil, nil, nil)
		svc.Subscribe(instanceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		instanceListener := listener.NewInstance(nil, nil, nil)

		svc.Subscribe(instanceListener)
		svc.Subscribe(instanceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		instanceListenerA := listener.NewInstance(nil, nil, nil)
		instanceListenerB := listener.NewInstance(nil, nil, nil)

		svc.Subscribe(instanceListenerA)
		svc.Subscribe(instanceListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)

		svc.Subscribe(nil)
	})
}

func TestServiceInstanceService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		instanceListener := listener.NewInstance(nil, nil, nil)
		svc.Subscribe(instanceListener)

		svc.Unsubscribe(instanceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		instanceListener := listener.NewInstance(nil, nil, nil)
		svc.Subscribe(instanceListener)
		svc.Subscribe(instanceListener)

		svc.Unsubscribe(instanceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		instanceListenerA := listener.NewInstance(nil, nil, nil)
		instanceListenerB := listener.NewInstance(nil, nil, nil)
		svc.Subscribe(instanceListenerA)
		svc.Subscribe(instanceListenerB)

		svc.Unsubscribe(instanceListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc := servicecatalog.NewServiceInstanceService(fixInformer(), nil)

		svc.Unsubscribe(nil)
	})
}

func fixServiceInstance(name, namespace string) *v1beta1.ServiceInstance {
	instance := v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassName:         "",
				ClusterServiceClassExternalName: "",
			},
		},
	}

	return &instance
}

func fixServiceInstanceWithClusterPlanRef(name, namespace, className, externalClassName string) *v1beta1.ServiceInstance {
	plan := v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassName:         className,
				ClusterServiceClassExternalName: externalClassName,
			},
		},
	}

	return &plan
}

func fixServiceInstanceWithPlanRef(name, namespace, className, externalClassName string) *v1beta1.ServiceInstance {
	plan := v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassName:         className,
				ServiceClassExternalName: externalClassName,
			},
		},
	}

	return &plan
}

func fixServiceInstanceWithStatus(name, namespace string) *v1beta1.ServiceInstance {
	plan := v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: v1beta1.ServiceInstanceStatus{
			AsyncOpInProgress: false,
			Conditions: []v1beta1.ServiceInstanceCondition{
				{
					Type:    v1beta1.ServiceInstanceConditionReady,
					Status:  v1beta1.ConditionTrue,
					Message: "Working",
					Reason:  "Testing",
				},
			},
		},
	}

	return &plan
}

func fixInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer()

	return informer
}
