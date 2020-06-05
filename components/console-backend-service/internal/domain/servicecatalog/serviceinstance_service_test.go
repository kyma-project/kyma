package servicecatalog_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	listener "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
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
		namespace := "testNs"

		serviceInstance := fixServiceInstance(instanceName, namespace)
		serviceInstanceInformer := fixInformer(serviceInstance)

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(t, serviceInstance, instance)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceInstanceInformer := informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer()

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		instance, err := svc.Find("doesntExist", "notExistingNs")
		require.NoError(t, err)
		assert.Nil(t, instance)
	})
}

func TestServiceInstanceService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "ns"
		serviceInstance1 := fixServiceInstance("1", namespace)
		serviceInstance2 := fixServiceInstance("2", namespace)
		serviceInstance3 := fixServiceInstance("3", "ns3")

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		instances, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.ServiceInstance{
			serviceInstance1, serviceInstance2,
		}, instances)
	})

	t.Run("NotFound", func(t *testing.T) {
		serviceInstanceInformer := fixInformer()

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		var emptyArray []*v1beta1.ServiceInstance
		instances, err := svc.List("notExisting", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, instances)
	})
}

func TestServiceInstanceService_ListForStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "ns"
		status := status.ServiceInstanceStatusTypeRunning

		serviceInstance1 := fixServiceInstanceWithStatus("1", namespace)
		serviceInstance2 := fixServiceInstanceWithStatus("2", namespace)
		serviceInstance3 := fixServiceInstanceWithStatus("3", "ns2")

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		instances, err := svc.ListForStatus(namespace, pager.PagingParams{}, &status)
		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.ServiceInstance{
			serviceInstance1, serviceInstance2,
		}, instances)
	})

	t.Run("NotFound", func(t *testing.T) {
		status := status.ServiceInstanceStatusTypeRunning

		serviceInstanceInformer := fixInformer()

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		var emptyArray []*v1beta1.ServiceInstance
		instances, err := svc.ListForStatus("notExisting", pager.PagingParams{}, &status)
		require.NoError(t, err)
		assert.Equal(t, emptyArray, instances)
	})
}

func TestServiceInstanceService_ListForClusterServiceClass(t *testing.T) {
	t.Run("Service Instance exist", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		namespace := "ns"
		serviceInstance1 := fixServiceInstanceWithClusterPlanRef("1", namespace, className, "")
		serviceInstance2 := fixServiceInstanceWithClusterPlanRef("2", namespace, "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithClusterPlanRef("3", "ns2", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)
		expected := []*v1beta1.ServiceInstance{
			serviceInstance1, serviceInstance2, serviceInstance3,
		}

		instances, err := svc.ListForClusterServiceClass(className, externalClassName, nil)
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, instances)
	})

	t.Run("Service Instance don't exist", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		testClassName := "test"
		testExternalClassName := "test"

		serviceInstance1 := fixServiceInstanceWithClusterPlanRef("1", "ns", className, "")
		serviceInstance2 := fixServiceInstanceWithClusterPlanRef("2", "ns2", "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithClusterPlanRef("3", "ns3", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		var emptyArray []*v1beta1.ServiceInstance
		instances, err := svc.ListForClusterServiceClass(testClassName, testExternalClassName, nil)
		require.NoError(t, err)
		assert.Equal(t, emptyArray, instances)
	})

	t.Run("Service Instance exist in namespace", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		namespace := "ns"
		serviceInstance1 := fixServiceInstanceWithClusterPlanRef("1", namespace, className, "")
		serviceInstance2 := fixServiceInstanceWithClusterPlanRef("2", namespace, "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithClusterPlanRef("3", "ns2", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)
		expected := []*v1beta1.ServiceInstance{
			serviceInstance1, serviceInstance2,
		}

		instances, err := svc.ListForClusterServiceClass(className, externalClassName, &namespace)
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, instances)
	})

	t.Run("Service Instance don't exist in namespace", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		testClassName := "test"
		testExternalClassName := "test"

		namespace := "ns"
		serviceInstance1 := fixServiceInstanceWithClusterPlanRef("1", namespace, className, "")
		serviceInstance2 := fixServiceInstanceWithClusterPlanRef("2", "ns2", "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithClusterPlanRef("3", "ns3", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		var emptyArray []*v1beta1.ServiceInstance
		instances, err := svc.ListForClusterServiceClass(testClassName, testExternalClassName, &namespace)
		require.NoError(t, err)
		assert.Equal(t, emptyArray, instances)
	})
}

func TestServiceInstanceService_ListForServiceClass(t *testing.T) {
	t.Run("Service Instances exist", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		namespace := "ns"
		serviceInstance1 := fixServiceInstanceWithPlanRef("1", namespace, className, "")
		serviceInstance2 := fixServiceInstanceWithPlanRef("2", namespace, "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithPlanRef("3", "ns2", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)
		expected := []*v1beta1.ServiceInstance{
			serviceInstance1, serviceInstance2,
		}

		instances, err := svc.ListForServiceClass(className, externalClassName, namespace)
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, instances)
	})

	t.Run("Service Instances don't exist", func(t *testing.T) {
		className := "exampleClassName"
		externalClassName := "exampleExternalClassName"

		namespace := "ns"
		testClassName := "test"
		testExternalClassName := "test"

		serviceInstance1 := fixServiceInstanceWithPlanRef("1", "ns", className, "")
		serviceInstance2 := fixServiceInstanceWithPlanRef("2", "ns2", "", externalClassName)
		serviceInstance3 := fixServiceInstanceWithPlanRef("3", "ns3", className, externalClassName)

		serviceInstanceInformer := fixInformer(serviceInstance1, serviceInstance2, serviceInstance3)

		svc, err := servicecatalog.NewServiceInstanceService(serviceInstanceInformer, nil)
		require.NoError(t, err)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInstanceInformer)

		var emptyArray []*v1beta1.ServiceInstance
		instances, err := svc.ListForServiceClass(testClassName, testExternalClassName, namespace)
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

		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), client)
		require.NoError(t, err)

		params := servicecatalog.NewServiceInstanceCreateParameters("name", "namespace", []string{"test", "label"}, "planName", true, "className", true, nil)
		result, err := svc.Create(*params)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

}

func TestServiceInstanceService_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := fixServiceInstance("test", "test")
		client := fake.NewSimpleClientset(instance)
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), client)
		require.NoError(t, err)

		err = svc.Delete("test", "test")

		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		testErr := fmt.Errorf("Test")
		instance := fixServiceInstance("test", "test")
		client := fake.NewSimpleClientset(instance)
		client.PrependReactor("*", "*", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, testErr
		})
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), client)
		require.NoError(t, err)

		err = svc.Delete("test", "test")

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
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)

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
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)

		result := svc.IsBindableWithLocalRefs(class, plan)

		assert.Equal(t, tc.expected, result)
	}
}

func TestServiceInstanceService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)
		instanceListener := listener.NewInstance(nil, nil, nil)
		svc.Subscribe(instanceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)
		instanceListener := listener.NewInstance(nil, nil, nil)

		svc.Subscribe(instanceListener)
		svc.Subscribe(instanceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)
		instanceListenerA := listener.NewInstance(nil, nil, nil)
		instanceListenerB := listener.NewInstance(nil, nil, nil)

		svc.Subscribe(instanceListenerA)
		svc.Subscribe(instanceListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)

		svc.Subscribe(nil)
	})
}

func TestServiceInstanceService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)
		instanceListener := listener.NewInstance(nil, nil, nil)
		svc.Subscribe(instanceListener)

		svc.Unsubscribe(instanceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)
		instanceListener := listener.NewInstance(nil, nil, nil)
		svc.Subscribe(instanceListener)
		svc.Subscribe(instanceListener)

		svc.Unsubscribe(instanceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)
		instanceListenerA := listener.NewInstance(nil, nil, nil)
		instanceListenerB := listener.NewInstance(nil, nil, nil)
		svc.Subscribe(instanceListenerA)
		svc.Subscribe(instanceListenerB)

		svc.Unsubscribe(instanceListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := servicecatalog.NewServiceInstanceService(fixInformer(), nil)
		require.NoError(t, err)

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
