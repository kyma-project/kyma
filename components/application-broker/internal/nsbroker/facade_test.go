package nsbroker_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc_fake "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker/automock"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core_v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
	k8s_testing "k8s.io/client-go/testing"
)

func TestNsBrokerCreateHappyPath(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientset := k8s_fake.NewSimpleClientset()
	mockBrokerSyncer := &automock.BrokerSyncer{}
	defer mockBrokerSyncer.AssertExpectations(t)

	svcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local/%s", fixABService(), fixWorkingNs(), "stage")
	mockBrokerSyncer.On("SyncBroker", "stage").Once().Return(nil)

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientset.CoreV1(), mockBrokerSyncer, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixABService(), fixTargetPort(), spy.NewLogDummy())
	// WHEN
	err := sut.Create(fixDestNs())

	// THEN
	require.NoError(t, err)
	actualBroker, err := scFakeClientset.ServicecatalogV1beta1().ServiceBrokers(fixDestNs()).Get("application-broker", v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "true", actualBroker.Labels["namespaced-application-broker"])
	assert.Equal(t, svcURL, actualBroker.Spec.URL)

	require.NoError(t, err)
}

func TestNsBrokerCreateErrorsHandlingOnBrokerCreation(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()

	k8sFakeClientset := k8s_fake.NewSimpleClientset()
	mockBrokerSyncer := &automock.BrokerSyncer{}
	defer mockBrokerSyncer.AssertExpectations(t)

	scFakeClientset.PrependReactor("create", "servicebrokers", failingReactor(fixError()))
	logSink := spy.NewLogSink()
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientset.CoreV1(), mockBrokerSyncer, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixABService(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Create(fixDestNs())
	// THEN
	assert.Error(t, err)
	fmt.Println(logSink.DumpAll())
	logSink.AssertLogged(t, logrus.WarnLevel, "Creation of namespaced-broker for namespace [stage] results in error: [some error]. AlreadyExist errors will be ignored.")
}

func TestNsBrokerCreateAlreadyExistErrorsIgnored(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientSet := k8s_fake.NewSimpleClientset()
	mockBrokerSyncer := &automock.BrokerSyncer{}
	defer mockBrokerSyncer.AssertExpectations(t)
	mockBrokerSyncer.On("SyncBroker", fixDestNs()).Once().Return(nil)

	logSink := spy.NewLogSink()
	scFakeClientset.PrependReactor("create", "servicebrokers", failingReactor(fixAlreadyExistError()))
	scFakeClientset.PrependReactor("get", "servicebrokers", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.ServiceBroker{
			ObjectMeta: v1.ObjectMeta{
				Name: nsbroker.NamespacedBrokerName,
				UID:  "1234-abcd",
			},
		}, nil
	})
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientSet.CoreV1(), mockBrokerSyncer, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixABService(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Create(fixDestNs())
	assert.NoError(t, err)
	logSink.AssertLogged(t, logrus.InfoLevel, "ServiceBroker for namespace [stage] already exist. Attempt to get resource.")
}

func TestNsBrokerDeleteHappyPath(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientSet := k8s_fake.NewSimpleClientset()

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientSet.CoreV1(), nil, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixABService(), fixTargetPort(), spy.NewLogDummy())
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.NoError(t, err)
	assertPerformedActionRemoveBroker(t, scFakeClientset.Actions())
}

func TestNsBrokerDeleteErrorOnRemovingBroker(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientSet := k8s_fake.NewSimpleClientset(&core_v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      "ab-ns-for-stage",
			Namespace: fixWorkingNs(),
		}})
	scFakeClientset.PrependReactor("delete", "servicebrokers", failingReactor(fixError()))
	logSink := spy.NewLogSink()

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientSet.CoreV1(), nil, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixABService(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.Error(t, err)
	logSink.AssertLogged(t, logrus.WarnLevel, "Deletion of namespaced-broker for namespace [stage] results in error")
}

func TestNsBrokerDeleteNotFoundErrorsIgnored(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientSet := k8s_fake.NewSimpleClientset()
	logSink := spy.NewLogSink()
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientSet.CoreV1(), nil, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixABService(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.NoError(t, err)
}

func TestNsBrokerDoesNotExist(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), nil, nil, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixABService(), fixTargetPort(), spy.NewLogDummy())
	// WHEN
	ex, err := sut.Exist(fixDestNs())
	// THEN
	require.NoError(t, err)
	assert.False(t, ex)
}

func TestNsBrokerExist(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset(&v1beta1.ServiceBroker{
		ObjectMeta: v1.ObjectMeta{
			Name:      "application-broker",
			Namespace: fixDestNs(),
		}})

	k8sFakeClientset := k8s_fake.NewSimpleClientset(
		&core_v1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name:      "ab-ns-for-stage",
				Namespace: fixWorkingNs(),
			},
		},
	)

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientset.CoreV1(), nil, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixABService(), fixTargetPort(), spy.NewLogDummy())
	// WHEN
	ex, err := sut.Exist(fixDestNs())
	// THEN
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestNsBrokerExistOnError(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	scFakeClientset.PrependReactor("get", "servicebrokers", failingReactor(fixError()))
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), nil, nil, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixABService(), fixTargetPort(), spy.NewLogDummy())
	// WHEN
	_, err := sut.Exist(fixDestNs())
	// THEN
	assert.EqualError(t, err, "while checking if ServiceBroker [application-broker] exists in the namespace [stage]: some error")

}

func fixABSelectorKey() string {
	return "app"
}

func fixABSelectorValue() string {
	return "ab"
}

func fixTargetPort() int32 {
	return int32(8080)
}

func fixDestNs() string {
	return "stage"
}

func fixABService() string {
	return "a-b-service"
}

func fixWorkingNs() string {
	return "kyma-system"
}

func failingReactor(retErr error) k8s_testing.ReactionFunc {
	return func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, retErr
	}
}

func fixError() error {
	return errors.New("some error")
}

func fixAlreadyExistError() error {
	return k8s_errors.NewAlreadyExists(schema.GroupResource{}, "")
}

func assertPerformedActionCreateBroker(t *testing.T, actions []k8s_testing.Action) {
	assertPerformedAction(t, "create", "servicebrokers", actions)
}

func assertPerformedActionRemoveBroker(t *testing.T, actions []k8s_testing.Action) {
	assertPerformedAction(t, "delete", "servicebrokers", actions)
}

func assertPerformedAction(t *testing.T, verb, resource string, actions []k8s_testing.Action) {
	for _, action := range actions {
		if action.Matches(verb, resource) {
			return
		}
	}
	t.Errorf("Action %s %s not found", verb, resource)
}
