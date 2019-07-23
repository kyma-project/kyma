package broker

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"context"

	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	k8sigs "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestClusterServiceBrokerSync_Success(t *testing.T) {
	// given
	clusterServiceBroker := fixClusterServiceBroker()
	require.NoError(t, v1beta1.AddToScheme(scheme.Scheme))
	cli := k8sigs.NewFakeClientWithScheme(scheme.Scheme, clusterServiceBroker)
	csbSyncer := NewServiceBrokerSyncer(cli, clusterServiceBroker.Name, spy.NewLogDummy())

	// when
	err := csbSyncer.Sync()
	require.NoError(t, err)

	// then
	csb := &v1beta1.ClusterServiceBroker{}
	err = cli.Get(context.Background(), types.NamespacedName{Name: clusterServiceBroker.Name}, csb)
	require.NoError(t, err)

	assert.Equal(t, int64(1), csb.Spec.RelistRequests)
	assert.Nil(t, err)
}

func TestClusterServiceBrokerSync_NotExistingBroker(t *testing.T) {
	// given
	require.NoError(t, v1beta1.AddToScheme(scheme.Scheme))
	cli := k8sigs.NewFakeClientWithScheme(scheme.Scheme)
	csbSyncer := NewServiceBrokerSyncer(cli, fixClusterServiceBroker().Name, spy.NewLogDummy())

	// when
	err := csbSyncer.Sync()

	// then
	require.NoError(t, err)
}

func TestServiceBrokerSync_Success(t *testing.T) {
	// given
	serviceBroker := fixServiceBroker()
	require.NoError(t, v1beta1.AddToScheme(scheme.Scheme))
	cli := k8sigs.NewFakeClientWithScheme(scheme.Scheme, serviceBroker)
	csbSyncer := NewServiceBrokerSyncer(cli, serviceBroker.Name, spy.NewLogDummy())

	// when
	err := csbSyncer.SyncServiceBroker(fixDestNs())
	require.NoError(t, err)

	// then
	sb := &v1beta1.ServiceBroker{}
	err = cli.Get(context.Background(), types.NamespacedName{Namespace: fixDestNs(), Name: serviceBroker.Name}, sb)
	require.NoError(t, err)

	assert.Equal(t, int64(1), sb.Spec.RelistRequests)
	assert.Nil(t, err)
}

func TestServiceBrokerSync_NotExistingBroker(t *testing.T) {
	// given
	require.NoError(t, v1beta1.AddToScheme(scheme.Scheme))
	cli := k8sigs.NewFakeClientWithScheme(scheme.Scheme)
	csbSyncer := NewServiceBrokerSyncer(cli, fixServiceBroker().Name, spy.NewLogDummy())

	// when
	err := csbSyncer.SyncServiceBroker(fixDestNs())

	// then
	assert.NoError(t, err)
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

func fixServiceBroker() *v1beta1.ServiceBroker {
	return &v1beta1.ServiceBroker{
		ObjectMeta: v1.ObjectMeta{
			Name:      fixBrokerName(),
			Namespace: fixDestNs(),
			Labels: map[string]string{
				"app": "label",
			},
		},
	}
}
