package syncer

import (
	"testing"

	"errors"

	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s_testing "k8s.io/client-go/testing"
)

func TestServiceBrokerSync_Success(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	nsBrokerSync := NewServiceBrokerSyncer(client.Servicecatalog())

	// when
	err := nsBrokerSync.Sync(fixLabelSelector(), maxSyncRetries)

	// then
	sb, err := client.Servicecatalog().ServiceBrokers("test").Get("fix-a", v1.GetOptions{})
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

	nsBrokerSync := NewServiceBrokerSyncer(client.Servicecatalog())

	// when
	resultErr := nsBrokerSync.Sync(fixLabelSelector(), maxSyncRetries)

	// then
	sb, err := client.Servicecatalog().ServiceBrokers("test").Get("fix-a", v1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, int64(1), sb.Spec.RelistRequests)
	assert.Error(t, resultErr)
}

func TestServiceBrokerSync_Empty(t *testing.T) {
	// given
	client := fake.NewSimpleClientset()
	nsBrokerSync := NewServiceBrokerSyncer(client.Servicecatalog())

	// when
	err := nsBrokerSync.Sync(fixLabelSelector(), maxSyncRetries)

	// then
	assert.Nil(t, err)
}

func TestServiceBrokerSync_ListError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "servicebrokers", failingReactor(fixError()))

	nsBrokerSync := NewServiceBrokerSyncer(client.Servicecatalog())

	// when
	err := nsBrokerSync.Sync(fixLabelSelector(), maxSyncRetries)

	// then
	assert.EqualError(t, err, fmt.Sprintf("while listing ServiceBrokers [labelSelector: %s]: %v", fixLabelSelector(), fixError()))
}

func TestServiceBrokerSync_ConflictError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	client.PrependReactor("update", "servicebrokers", failingReactor(apiErrors.NewConflict(schema.GroupResource{}, "", fixError())))

	nsBrokerSync := NewServiceBrokerSyncer(client.Servicecatalog())

	// when
	err := nsBrokerSync.Sync(fixLabelSelector(), maxSyncRetries)

	// then
	assert.EqualError(t, err, "1 error occurred:\n\n* could not sync ServiceBroker \"fix-a\" [namespace: test], after 5 tries")
}

func TestServiceBrokerSync_UpdateError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	client.PrependReactor("update", "servicebrokers", failingReactor(fixError()))

	nsBrokerSync := NewServiceBrokerSyncer(client.Servicecatalog())

	// when
	err := nsBrokerSync.Sync(fixLabelSelector(), maxSyncRetries)

	// then
	assert.Error(t, err)
}

func TestServiceBrokerSync_GetError(t *testing.T) {
	// given
	client := fake.NewSimpleClientset(fixServiceBroker())
	client.PrependReactor("get", "servicebrokers", failingReactor(fixError()))

	nsBrokerSync := NewServiceBrokerSyncer(client.Servicecatalog())

	// when
	err := nsBrokerSync.Sync(fixLabelSelector(), maxSyncRetries)

	// then
	assert.Error(t, err)
}

func fixServiceBroker() *v1beta1.ServiceBroker {
	return &v1beta1.ServiceBroker{
		ObjectMeta: v1.ObjectMeta{
			Name:      "fix-a",
			Namespace: "test",
			Labels: map[string]string{
				"app": "label",
			},
		},
	}
}

func fixLabelSelector() string {
	return "app=label"
}

func failingReactor(retErr error) k8s_testing.ReactionFunc {
	return func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, retErr
	}
}

func passingReactor(sb *v1beta1.ServiceBroker) k8s_testing.ReactionFunc {
	return func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		return true, sb, nil
	}
}

func fixError() error {
	return errors.New("some error")
}
