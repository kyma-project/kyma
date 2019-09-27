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

func TestClusterServiceBrokerService_GetServiceBroker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		brokerName := "testExample"
		serviceBroker := fixClusterServiceBroker(brokerName)
		client := fake.NewSimpleClientset(serviceBroker)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer()

		svc := servicecatalog.NewClusterServiceBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		broker, err := svc.Find(brokerName)
		require.NoError(t, err)
		assert.Equal(t, serviceBroker, broker)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer()

		svc := servicecatalog.NewClusterServiceBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		broker, err := svc.Find("doesntExist")
		require.NoError(t, err)
		assert.Nil(t, broker)
	})
}

func TestClusterServiceBrokerService_ListServiceBrokers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		serviceBroker1 := fixClusterServiceBroker("1")
		serviceBroker2 := fixClusterServiceBroker("2")
		serviceBroker3 := fixClusterServiceBroker("3")
		client := fake.NewSimpleClientset(serviceBroker1, serviceBroker2, serviceBroker3)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer()

		svc := servicecatalog.NewClusterServiceBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		brokers, err := svc.List(pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.ClusterServiceBroker{
			serviceBroker1, serviceBroker2, serviceBroker3,
		}, brokers)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer()

		svc := servicecatalog.NewClusterServiceBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		var emptyArray []*v1beta1.ClusterServiceBroker
		brokers, err := svc.List(pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, brokers)
	})
}

func TestClusterServiceBrokerService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc := servicecatalog.NewClusterServiceBrokerService(fixClusterServiceBrokerInformer())
		instanceListener := listener.NewClusterServiceBroker(nil, nil, nil)
		svc.Subscribe(instanceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc := servicecatalog.NewClusterServiceBrokerService(fixClusterServiceBrokerInformer())
		instanceListener := listener.NewClusterServiceBroker(nil, nil, nil)

		svc.Subscribe(instanceListener)
		svc.Subscribe(instanceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc := servicecatalog.NewClusterServiceBrokerService(fixClusterServiceBrokerInformer())
		instanceListenerA := listener.NewClusterServiceBroker(nil, nil, nil)
		instanceListenerB := listener.NewClusterServiceBroker(nil, nil, nil)

		svc.Subscribe(instanceListenerA)
		svc.Subscribe(instanceListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc := servicecatalog.NewClusterServiceBrokerService(fixClusterServiceBrokerInformer())

		svc.Subscribe(nil)
	})
}

func TestClusterServiceBrokerService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc := servicecatalog.NewClusterServiceBrokerService(fixClusterServiceBrokerInformer())
		instanceListener := listener.NewClusterServiceBroker(nil, nil, nil)
		svc.Subscribe(instanceListener)

		svc.Unsubscribe(instanceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc := servicecatalog.NewClusterServiceBrokerService(fixClusterServiceBrokerInformer())
		instanceListener := listener.NewClusterServiceBroker(nil, nil, nil)
		svc.Subscribe(instanceListener)
		svc.Subscribe(instanceListener)

		svc.Unsubscribe(instanceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc := servicecatalog.NewClusterServiceBrokerService(fixClusterServiceBrokerInformer())
		instanceListenerA := listener.NewClusterServiceBroker(nil, nil, nil)
		instanceListenerB := listener.NewClusterServiceBroker(nil, nil, nil)
		svc.Subscribe(instanceListenerA)
		svc.Subscribe(instanceListenerB)

		svc.Unsubscribe(instanceListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc := servicecatalog.NewClusterServiceBrokerService(fixClusterServiceBrokerInformer())

		svc.Unsubscribe(nil)
	})
}

func fixClusterServiceBroker(name string) *v1beta1.ClusterServiceBroker {
	var mockTimeStamp metav1.Time

	broker := v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: mockTimeStamp,
		},
	}

	return &broker
}

func fixClusterServiceBrokerInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer()

	return informer
}
