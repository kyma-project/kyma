package broker

import (
	"fmt"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc_fake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s_testing "k8s.io/client-go/testing"
)

func TestServiceBrokerCreateHappyPath(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.BrokerSyncer{}
	brokerSyncer.On("SyncServiceBroker", fixDestNs()).Once().Return(nil)
	defer brokerSyncer.AssertExpectations(t)

	svcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local/ns/%s", fixService(), fixWorkingNs(), "stage")
	sut := NewBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService())
	// WHEN
	err := sut.Create(fixDestNs())

	// THEN
	require.NoError(t, err)
	actualBroker, err := scFakeClientset.Servicecatalog().ServiceBrokers(fixDestNs()).Get(fixBrokerName(), v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "true", actualBroker.Labels["namespaced-helm-broker"])
	assert.Equal(t, svcURL, actualBroker.Spec.URL)

	require.NoError(t, err)
}

func TestServiceBrokerCreateErrorsHandlingOnBrokerCreation(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()

	brokerSyncer := &automock.BrokerSyncer{}
	brokerSyncer.On("SyncServiceBroker", fixDestNs()).Once().Return(nil)
	defer brokerSyncer.AssertExpectations(t)

	scFakeClientset.PrependReactor("create", "servicebrokers", failingReactor(fixError()))
	sut := NewBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService())
	// WHEN
	err := sut.Create(fixDestNs())
	// THEN
	assert.Error(t, err)
}

func TestServiceBrokerCreateAlreadyExistErrorsIgnored(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.BrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)
	brokerSyncer.On("SyncServiceBroker", fixDestNs()).Once().Return(nil)

	scFakeClientset.PrependReactor("create", "servicebrokers", failingReactor(fixAlreadyExistError()))
	scFakeClientset.PrependReactor("get", "servicebrokers", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.ServiceBroker{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: fixBrokerName(),
				UID:  "1234-abcd",
			},
		}, nil
	})
	sut := NewBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService())
	// WHEN
	err := sut.Create(fixDestNs())
	assert.NoError(t, err)
}

func TestServiceBrokerDeleteHappyPath(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.BrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	sut := NewBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService())
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.NoError(t, err)
	assertPerformedAction(t, "delete", "servicebrokers", scFakeClientset.Actions())
}

func TestServiceBrokerDeleteErrorOnRemovingBroker(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.BrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	scFakeClientset.PrependReactor("delete", "servicebrokers", failingReactor(fixError()))

	sut := NewBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService())
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.Error(t, err)
}

func TestServiceBrokerDeleteNotFoundErrorsIgnored(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.BrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	sut := NewBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService())
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.NoError(t, err)
}

func TestServiceBrokerDoesNotExist(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.BrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	sut := NewBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService())
	// WHEN
	ex, err := sut.Exist(fixDestNs())
	// THEN
	require.NoError(t, err)
	assert.False(t, ex)
}

func TestServiceBrokerExist(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset(&v1beta1.ServiceBroker{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      fixBrokerName(),
			Namespace: fixDestNs(),
		}})

	brokerSyncer := &automock.BrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	sut := NewBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService())
	// WHEN
	ex, err := sut.Exist(fixDestNs())
	// THEN
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestServiceBrokerExistOnError(t *testing.T) {
	// GIVEN
	brokerSyncer := &automock.BrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	scFakeClientset := sc_fake.NewSimpleClientset()
	scFakeClientset.PrependReactor("get", "servicebrokers", failingReactor(fixError()))
	sut := NewBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService())
	// WHEN
	_, err := sut.Exist(fixDestNs())
	// THEN
	assert.EqualError(t, err, "while checking if ServiceBroker [helm-broker] exists in the namespace [stage]: some error")

}

func fixDestNs() string {
	return "stage"
}

func fixService() string {
	return "service"
}

func fixWorkingNs() string {
	return "kyma-system"
}

func failingReactor(retErr error) k8s_testing.ReactionFunc {
	return func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, retErr
	}
}

func fixAlreadyExistError() error {
	return k8s_errors.NewAlreadyExists(schema.GroupResource{}, "")
}

func assertPerformedAction(t *testing.T, verb, resource string, actions []k8s_testing.Action) {
	for _, action := range actions {
		if action.Matches(verb, resource) {
			return
		}
	}
	t.Errorf("Action %s %s not found", verb, resource)
}
