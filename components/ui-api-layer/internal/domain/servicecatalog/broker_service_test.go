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

func TestBrokerService_GetServiceBroker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		brokerName := "testExample"
		serviceBroker := fixServiceBroker(brokerName)
		client := fake.NewSimpleClientset(serviceBroker)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer()

		svc := servicecatalog.NewBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		broker, err := svc.Find(brokerName)
		require.NoError(t, err)
		assert.Equal(t, serviceBroker, broker)
	})

	t.Run("NotFound", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer()

		svc := servicecatalog.NewBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		broker, err := svc.Find("doesntExist")
		require.NoError(t, err)
		assert.Nil(t, broker)
	})
}

func TestBrokerService_ListServiceBrokers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		serviceBroker1 := fixServiceBroker("1")
		serviceBroker2 := fixServiceBroker("2")
		serviceBroker3 := fixServiceBroker("3")
		client := fake.NewSimpleClientset(serviceBroker1, serviceBroker2, serviceBroker3)

		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		serviceBrokerInformer := informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer()

		svc := servicecatalog.NewBrokerService(serviceBrokerInformer)

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

		svc := servicecatalog.NewBrokerService(serviceBrokerInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceBrokerInformer)

		var emptyArray []*v1beta1.ClusterServiceBroker
		brokers, err := svc.List(pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, brokers)
	})
}

func fixServiceBroker(name string) *v1beta1.ClusterServiceBroker {
	var mockTimeStamp metav1.Time

	broker := v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: mockTimeStamp,
		},
	}

	return &broker
}
