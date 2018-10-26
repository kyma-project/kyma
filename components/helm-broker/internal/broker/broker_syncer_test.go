package broker_test

import (
	"errors"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s_testing "k8s.io/client-go/testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
)

func TestServiceBrokerSync_Success(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixClusterServiceBroker())
	csbSyncer := broker.NewClusterServiceBrokerSyncer(client.ServicecatalogV1beta1().ClusterServiceBrokers(), spy.NewLogDummy())

	// when
	err := csbSyncer.Sync(fixClusterServiceBroker().Name, 2)

	// then
	sb, err := client.ServicecatalogV1beta1().ClusterServiceBrokers().Get(fixClusterServiceBroker().Name, v1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, int64(1), sb.Spec.RelistRequests)
	assert.Nil(t, err)
}

func TestServiceBrokerSync_NotExistingBroker(t *testing.T) {
	// given
	client := fake.NewSimpleClientset()
	csbSyncer := broker.NewClusterServiceBrokerSyncer(client.ServicecatalogV1beta1().ClusterServiceBrokers(), spy.NewLogDummy())

	// when
	err := csbSyncer.Sync(fixClusterServiceBroker().Name, 2)

	// then
	require.NoError(t, err)
}

func TestServiceBrokerSync_SuccessAfterConflictAndRetry(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixClusterServiceBroker())
	i := 0
	client.PrependReactor("update", "clusterservicebrokers", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		if i == 0 {
			i++
			return true, nil, apiErrors.NewConflict(schema.GroupResource{}, "", fixError())
		}
		return false, fixClusterServiceBroker(), nil
	})

	csbSyncer := broker.NewClusterServiceBrokerSyncer(client.Servicecatalog().ClusterServiceBrokers(), spy.NewLogDummy())

	// when
	err := csbSyncer.Sync(fixClusterServiceBroker().Name, 2)

	// then
	sb, err := client.ServicecatalogV1beta1().ClusterServiceBrokers().Get(fixClusterServiceBroker().Name, v1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, int64(1), sb.Spec.RelistRequests)
	assert.Nil(t, err)
}

func fixClusterServiceBroker() *v1beta1.ClusterServiceBroker {
	return &v1beta1.ClusterServiceBroker{
		ObjectMeta: v1.ObjectMeta{
			Name: "broker-name",
			Labels: map[string]string{
				"app": "label",
			},
		},
	}
}

func fixError() error {
	return errors.New("some error")
}
