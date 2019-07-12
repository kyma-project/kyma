package broker

import (
	"fmt"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc_fake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8s_testing "k8s.io/client-go/testing"
)

func TestClusterServiceBrokerCreateHappyPath(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.ClusterBrokerSyncer{}
	brokerSyncer.On("Sync").Once().Return(nil)
	defer brokerSyncer.AssertExpectations(t)

	svcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local/cluster", fixService(), fixWorkingNs())
	sut := NewClusterBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService(), fixBrokerName())
	// WHEN
	err := sut.Create()

	// THEN
	require.NoError(t, err)
	actualBroker, err := scFakeClientset.Servicecatalog().ClusterServiceBrokers().Get(fixBrokerName(), v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, svcURL, actualBroker.Spec.URL)

	require.NoError(t, err)
}

func TestClusterServiceBrokerCreateErrorsHandlingOnBrokerCreation(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()

	brokerSyncer := &automock.ClusterBrokerSyncer{}
	brokerSyncer.On("Sync").Once().Return(nil)
	defer brokerSyncer.AssertExpectations(t)

	scFakeClientset.PrependReactor("create", "clusterservicebrokers", failingReactor(fixError()))
	sut := NewClusterBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService(), fixBrokerName())
	// WHEN
	err := sut.Create()
	// THEN
	assert.Error(t, err)
}

func TestClusterServiceBrokerCreateAlreadyExistErrorsIgnored(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.ClusterBrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)
	brokerSyncer.On("Sync").Once().Return(nil)

	scFakeClientset.PrependReactor("create", "clusterservicebrokers", failingReactor(fixAlreadyExistError()))
	scFakeClientset.PrependReactor("get", "clusterservicebrokers", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.ClusterServiceBroker{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: fixBrokerName(),
				UID:  "1234-abcd",
			},
		}, nil
	})
	sut := NewClusterBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService(), fixBrokerName())
	// WHEN
	err := sut.Create()
	assert.NoError(t, err)
}

func TestClusterServiceBrokerDeleteHappyPath(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.ClusterBrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	sut := NewClusterBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService(), fixBrokerName())
	// WHEN
	err := sut.Delete()
	// THEN
	require.NoError(t, err)
	assertPerformedAction(t, "delete", "clusterservicebrokers", scFakeClientset.Actions())
}

func TestClusterServiceBrokerDeleteErrorOnRemovingBroker(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.ClusterBrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	scFakeClientset.PrependReactor("delete", "clusterservicebrokers", failingReactor(fixError()))

	sut := NewClusterBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService(), fixBrokerName())
	// WHEN
	err := sut.Delete()
	// THEN
	require.Error(t, err)
}

func TestClusterServiceBrokerDeleteNotFoundErrorsIgnored(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.ClusterBrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	sut := NewClusterBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService(), fixBrokerName())
	// WHEN
	err := sut.Delete()
	// THEN
	require.NoError(t, err)
}

func TestClusterServiceBrokerDoesNotExist(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	brokerSyncer := &automock.ClusterBrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	sut := NewClusterBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService(), fixBrokerName())
	// WHEN
	ex, err := sut.Exist()
	// THEN
	require.NoError(t, err)
	assert.False(t, ex)
}

func TestClusterServiceBrokerExist(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset(&v1beta1.ClusterServiceBroker{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: fixBrokerName(),
		}})

	brokerSyncer := &automock.ClusterBrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	sut := NewClusterBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService(), fixBrokerName())
	// WHEN
	ex, err := sut.Exist()
	// THEN
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestClusterServiceBrokerExistOnError(t *testing.T) {
	// GIVEN
	brokerSyncer := &automock.ClusterBrokerSyncer{}
	defer brokerSyncer.AssertExpectations(t)

	scFakeClientset := sc_fake.NewSimpleClientset()
	scFakeClientset.PrependReactor("get", "clusterservicebrokers", failingReactor(fixError()))
	sut := NewClusterBrokersFacade(scFakeClientset.ServicecatalogV1beta1(), brokerSyncer, fixWorkingNs(), fixService(), fixBrokerName())
	// WHEN
	_, err := sut.Exist()
	// THEN
	assert.EqualError(t, err, "while checking if ClusterServiceBroker [helm-broker] exists: some error")

}

func fixBrokerName() string {
	return "helm-broker"
}
