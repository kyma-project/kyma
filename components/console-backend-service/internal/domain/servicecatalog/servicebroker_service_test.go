package servicecatalog_test

import (
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

func TestServiceBrokerService_GetServiceBroker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		brokerName := "testExample"
		nsName := "exampleNs"
		serviceBroker := fixServiceBroker(brokerName, nsName)
		client := fake.NewSimpleClientset(serviceBroker)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ServiceBrokers().Informer()

		svc := servicecatalog.NewServiceBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		broker, err := svc.Find(brokerName, nsName)
		require.NoError(t, err)
		assert.Equal(t, serviceBroker, broker)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ServiceBrokers().Informer()

		svc := servicecatalog.NewServiceBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		broker, err := svc.Find("doesntExist", "ns")
		require.NoError(t, err)
		assert.Nil(t, broker)
	})
}

func TestServiceBrokerService_ListServiceBrokers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		nsName := "exampleNs"
		serviceBroker1 := fixServiceBroker("1", nsName)
		serviceBroker2 := fixServiceBroker("2", nsName)
		serviceBroker3 := fixServiceBroker("3", nsName)
		client := fake.NewSimpleClientset(serviceBroker1, serviceBroker2, serviceBroker3)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ServiceBrokers().Informer()

		svc := servicecatalog.NewServiceBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		brokers, err := svc.List(nsName, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.ServiceBroker{
			serviceBroker1, serviceBroker2, serviceBroker3,
		}, brokers)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ServiceBrokers().Informer()

		svc := servicecatalog.NewServiceBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		var emptyArray []*v1beta1.ServiceBroker
		brokers, err := svc.List("ns", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, brokers)
	})
}

func TestServiceBrokerService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc := servicecatalog.NewServiceBrokerService(fixServiceBrokerInformer())
		instanceListener := listener.NewServiceBroker(nil, nil, nil)
		svc.Subscribe(instanceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc := servicecatalog.NewServiceBrokerService(fixServiceBrokerInformer())
		instanceListener := listener.NewServiceBroker(nil, nil, nil)

		svc.Subscribe(instanceListener)
		svc.Subscribe(instanceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc := servicecatalog.NewServiceBrokerService(fixServiceBrokerInformer())
		instanceListenerA := listener.NewServiceBroker(nil, nil, nil)
		instanceListenerB := listener.NewServiceBroker(nil, nil, nil)

		svc.Subscribe(instanceListenerA)
		svc.Subscribe(instanceListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc := servicecatalog.NewServiceBrokerService(fixServiceBrokerInformer())

		svc.Subscribe(nil)
	})
}

func TestServiceBrokerService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc := servicecatalog.NewServiceBrokerService(fixServiceBrokerInformer())
		instanceListener := listener.NewServiceBroker(nil, nil, nil)
		svc.Subscribe(instanceListener)

		svc.Unsubscribe(instanceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc := servicecatalog.NewServiceBrokerService(fixServiceBrokerInformer())
		instanceListener := listener.NewServiceBroker(nil, nil, nil)
		svc.Subscribe(instanceListener)
		svc.Subscribe(instanceListener)

		svc.Unsubscribe(instanceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc := servicecatalog.NewServiceBrokerService(fixServiceBrokerInformer())
		instanceListenerA := listener.NewServiceBroker(nil, nil, nil)
		instanceListenerB := listener.NewServiceBroker(nil, nil, nil)
		svc.Subscribe(instanceListenerA)
		svc.Subscribe(instanceListenerB)

		svc.Unsubscribe(instanceListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc := servicecatalog.NewServiceBrokerService(fixServiceBrokerInformer())

		svc.Unsubscribe(nil)
	})
}

func fixServiceBroker(name, namespace string) *v1beta1.ServiceBroker {
	var mockTimeStamp metav1.Time

	broker := v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: mockTimeStamp,
		},
	}

	return &broker
}

func fixServiceBrokerInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Servicecatalog().V1beta1().ServiceBrokers().Informer()

	return informer
}
