package syncer

import (
	"testing"

	"errors"

	"fmt"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s_testing "k8s.io/client-go/testing"
)

func TestServiceBrokerSync_Success(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.Sync(maxSyncRetries)

	// then
	sb, err := client.ServicecatalogV1beta1().ServiceBrokers("test").Get(nsbroker.NamespacedBrokerName, v1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, int64(1), sb.Spec.RelistRequests)
	assert.Nil(t, err)
}

func TestServiceBrokerSync_SuccessAfterRetry(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())

	i := 0
	client.PrependReactor("get", "servicebrokers", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		if i == 0 {
			i++
			return true, nil, fixError()
		}
		return false, fixServiceBroker(), nil
	})

	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	resultErr := nsBrokerSync.Sync(maxSyncRetries)

	// then
	sb, err := client.ServicecatalogV1beta1().ServiceBrokers("test").Get(nsbroker.NamespacedBrokerName, v1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, int64(1), sb.Spec.RelistRequests)
	assert.Error(t, resultErr)
}

func TestServiceBrokerSync_Empty(t *testing.T) {
	// given
	client := fake.NewSimpleClientset()
	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.Sync(maxSyncRetries)

	// then
	assert.Nil(t, err)
}

func TestServiceBrokerSync_ListError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "servicebrokers", failingReactor(fixError()))

	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.Sync(maxSyncRetries)

	// then
	assert.EqualError(t, err, fmt.Sprintf("while listing ServiceBrokers [labelSelector: %s]: %v", fixLabelSelector(), fixError()))
}

func TestServiceBrokerSync_ConflictError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	client.PrependReactor("update", "servicebrokers", failingReactor(apiErrors.NewConflict(schema.GroupResource{}, "", fixError())))

	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.Sync(maxSyncRetries)

	// then
	assert.EqualError(t, err, "1 error occurred:\n\t* could not sync ServiceBroker \"application-broker\" [namespace: test], after 5 tries\n\n")
}

func TestServiceBrokerSync_UpdateError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	client.PrependReactor("update", "servicebrokers", failingReactor(fixError()))

	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.Sync(maxSyncRetries)

	// then
	assert.Error(t, err)
}

func TestServiceBrokerSync_GetError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	client.PrependReactor("get", "servicebrokers", failingReactor(fixError()))

	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.Sync(maxSyncRetries)

	// then
	assert.Error(t, err)
}

func TestServiceBrokerSync_SyncBroker_Success(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())

	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.SyncBroker("test")

	// then
	sb, err := client.ServicecatalogV1beta1().ServiceBrokers("test").Get(nsbroker.NamespacedBrokerName, v1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, int64(1), sb.Spec.RelistRequests)
	assert.NoError(t, err)
}

func TestServiceBrokerSync_SyncBroker_GetError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	client.PrependReactor("get", "servicebrokers", failingReactor(fixError()))

	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.SyncBroker("test")

	// then
	assert.Error(t, err)
}

func TestServiceBrokerSync_SyncBroker_UpdateError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	client.PrependReactor("update", "servicebrokers", failingReactor(fixError()))

	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.SyncBroker("test")

	// then
	assert.Error(t, err)
}

func TestServiceBrokerSync_SyncBroker_ConflictError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	client.PrependReactor("update", "servicebrokers", failingReactor(apiErrors.NewConflict(schema.GroupResource{}, "", fixError())))

	nsBrokerSync := NewServiceBrokerSyncer(client.ServicecatalogV1beta1())

	// when
	err := nsBrokerSync.SyncBroker("test")

	// then
	assert.EqualError(t, err, "could not sync service broker (application-broker) after 5 retries")
}

func fixServiceBroker() *v1beta1.ServiceBroker {
	return &v1beta1.ServiceBroker{
		ObjectMeta: v1.ObjectMeta{
			Name:      nsbroker.NamespacedBrokerName,
			Namespace: "test",
			Labels: map[string]string{
				"namespaced-application-broker": "true",
			},
		},
	}
}

func fixLabelSelector() string {
	return "namespaced-application-broker=true"
}

func failingReactor(retErr error) k8s_testing.ReactionFunc {
	return func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, retErr
	}
}

func fixError() error {
	return errors.New("some error")
}
