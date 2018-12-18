package nsbroker_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc_fake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker/automock"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	core_v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
	k8s_testing "k8s.io/client-go/testing"
)

func TestNsBrokerCreateHappyPath(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientset := k8s_fake.NewSimpleClientset()
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)
	mockServiceChecker := &automock.ServiceChecker{}
	defer mockServiceChecker.AssertExpectations(t)
	mockBrokerSyncer := &automock.BrokerSyncer{}
	defer mockBrokerSyncer.AssertExpectations(t)

	serviceName := "reb-ns-for-stage"
	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")
	done := make(chan struct{})
	svcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", serviceName, fixWorkingNs())
	mockServiceChecker.On("WaitUntilIsAvailable", fmt.Sprintf("%s/statusz", svcURL), mock.Anything).Once()
	mockBrokerSyncer.On("SyncBroker", "stage").Run(func(args mock.Arguments) {
		close(done)
	}).Once().Return(nil)

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientset.CoreV1(), mockServiceNameProvider, mockBrokerSyncer, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), spy.NewLogDummy())
	sut = sut.WithServiceChecker(mockServiceChecker)
	// WHEN
	err := sut.Create(fixDestNs())

	// THEN
	require.NoError(t, err)
	actualBroker, err := scFakeClientset.Servicecatalog().ServiceBrokers(fixDestNs()).Get("remote-env-broker", v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "true", actualBroker.Labels["namespaced-remote-env-broker"])
	assert.Equal(t, svcURL, actualBroker.Spec.URL)

	actualService, err := k8sFakeClientset.CoreV1().Services(fixWorkingNs()).Get(serviceName, v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, fixRebSelectorValue(), actualService.Spec.Selector[fixRebSelectorKey()])
	require.Len(t, actualService.Spec.Ports, 1)
	assert.Equal(t, fixTargetPort(), actualService.Spec.Ports[0].TargetPort.IntVal)

	// wait for SyncBroker
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "SyncBroker was not called")
	}
}

func TestNsBrokerCreateErrorsHandlingOnBrokerCreation(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientset := k8s_fake.NewSimpleClientset()
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)
	mockServiceChecker := &automock.ServiceChecker{}
	defer mockServiceChecker.AssertExpectations(t)
	mockBrokerSyncer := &automock.BrokerSyncer{}
	defer mockBrokerSyncer.AssertExpectations(t)

	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")

	scFakeClientset.PrependReactor("create", "servicebrokers", failingReactor(fixError()))
	logSink := spy.NewLogSink()
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientset.CoreV1(), mockServiceNameProvider, mockBrokerSyncer, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Create(fixDestNs())
	// THEN
	assert.Error(t, err)
	assertPerformedActionCreateService(t, k8sFakeClientset.Actions())
	fmt.Println(logSink.DumpAll())
	logSink.AssertLogged(t, logrus.WarnLevel, "Creation of namespaced-broker for namespace [stage] results in error: [some error]. AlreadyExist errors will be ignored.")
}

func TestNsBrokerCreateErrorsHandlingOnServiceCreation(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientset := k8s_fake.NewSimpleClientset()
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)
	mockServiceChecker := &automock.ServiceChecker{}
	defer mockServiceChecker.AssertExpectations(t)
	mockBrokerSyncer := &automock.BrokerSyncer{}
	defer mockBrokerSyncer.AssertExpectations(t)

	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")

	logSink := spy.NewLogSink()
	k8sFakeClientset.PrependReactor("create", "services", failingReactor(fixError()))
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientset.CoreV1(), mockServiceNameProvider, mockBrokerSyncer, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Create(fixDestNs())
	// THEN
	assert.Error(t, err)
	assertPerformedActionCreateBroker(t, scFakeClientset.Actions())
	logSink.AssertLogged(t, logrus.WarnLevel, "Creation of service for namespaced-broker for namespace [stage] results in error: [some error]. AlreadyExist error will be ignored.")
}

func TestNsBrokerCreateAlreadyExistErrorsIgnored(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientSet := k8s_fake.NewSimpleClientset()
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)
	mockServiceChecker := &automock.ServiceChecker{}
	defer mockServiceChecker.AssertExpectations(t)
	mockBrokerSyncer := &automock.BrokerSyncer{}
	defer mockBrokerSyncer.AssertExpectations(t)

	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")

	logSink := spy.NewLogSink()
	k8sFakeClientSet.PrependReactor("create", "services", failingReactor(fixAlreadyExistError()))
	scFakeClientset.PrependReactor("create", "servicebrokers", failingReactor(fixAlreadyExistError()))
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientSet.CoreV1(), mockServiceNameProvider, mockBrokerSyncer, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Create(fixDestNs())
	assert.NoError(t, err)
	logSink.AssertLogged(t, logrus.WarnLevel, "Creation of namespaced-broker for namespace [stage] results in error")
	logSink.AssertLogged(t, logrus.WarnLevel, "Creation of service for namespaced-broker for namespace [stage] results in error")

}

func TestNsBrokerDeleteHappyPath(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientSet := k8s_fake.NewSimpleClientset()
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)

	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientSet.CoreV1(), mockServiceNameProvider, nil, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), spy.NewLogDummy())
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.NoError(t, err)
	assertPerformedActionRemoveBroker(t, scFakeClientset.Actions())
	assertPerformedActionRemoveService(t, k8sFakeClientSet.Actions())
}

func TestNsBrokerDeleteErrorOnRemovingService(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset(&v1beta1.ServiceBroker{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "remote-env-broker",
			Namespace: fixDestNs(),
		}})
	k8sFakeClientSet := k8s_fake.NewSimpleClientset()
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)

	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")
	k8sFakeClientSet.PrependReactor("delete", "services", failingReactor(fixError()))
	logSink := spy.NewLogSink()

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientSet.CoreV1(), mockServiceNameProvider, nil, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.Error(t, err)
	assertPerformedActionRemoveBroker(t, scFakeClientset.Actions())
	logSink.AssertLogged(t, logrus.WarnLevel, "Deletion of service for namespaced-broker for namespace [stage] results in error: [some error]")
}

func TestNsBrokerDeleteErrorOnRemovingBroker(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientSet := k8s_fake.NewSimpleClientset(&core_v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "reb-ns-for-stage",
			Namespace: fixWorkingNs(),
		}})
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)

	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")
	scFakeClientset.PrependReactor("delete", "servicebrokers", failingReactor(fixError()))
	logSink := spy.NewLogSink()

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientSet.CoreV1(), mockServiceNameProvider, nil, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.Error(t, err)
	assertPerformedActionRemoveService(t, k8sFakeClientSet.Actions())
	logSink.AssertLogged(t, logrus.WarnLevel, "Deletion of namespaced-broker for namespace [stage] results in error")
}

func TestNsBrokerDeleteNotFoundErrorsIgnored(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientSet := k8s_fake.NewSimpleClientset()
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)

	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")
	logSink := spy.NewLogSink()
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientSet.CoreV1(), mockServiceNameProvider, nil, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), logSink.Logger)
	// WHEN
	err := sut.Delete(fixDestNs())
	// THEN
	require.NoError(t, err)
	logSink.AssertLogged(t, logrus.WarnLevel, "Deletion of service for namespaced-broker for namespace [stage] results in error")
	logSink.AssertLogged(t, logrus.WarnLevel, "Deletion of namespaced-broker for namespace [stage] results in error")
}

func TestNsBrokerDoesNotExist(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset()
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), nil, nil, nil, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), spy.NewLogDummy())
	// WHEN
	ex, err := sut.Exist(fixDestNs())
	// THEN
	require.NoError(t, err)
	assert.False(t, ex)
}

func TestNsBrokerDoesNotExistIfOnlyServiceIsAbsent(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset(&v1beta1.ServiceBroker{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "remote-env-broker",
			Namespace: fixDestNs(),
		}})

	k8sFakeClientset := k8s_fake.NewSimpleClientset()
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)

	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientset.CoreV1(), mockServiceNameProvider, nil, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), spy.NewLogDummy())
	// WHEN
	ex, err := sut.Exist(fixDestNs())
	// THEN
	require.NoError(t, err)
	assert.False(t, ex)
}

func TestNsBrokerExist(t *testing.T) {
	// GIVEN
	scFakeClientset := sc_fake.NewSimpleClientset(&v1beta1.ServiceBroker{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "remote-env-broker",
			Namespace: fixDestNs(),
		}})

	k8sFakeClientset := k8s_fake.NewSimpleClientset(
		&core_v1.Service{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "reb-ns-for-stage",
				Namespace: fixWorkingNs(),
			},
		},
	)
	mockServiceNameProvider := &automock.ServiceNameProvider{}
	defer mockServiceNameProvider.AssertExpectations(t)

	mockServiceNameProvider.On("GetServiceNameForNsBroker", "stage").Return("reb-ns-for-stage")

	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientset.CoreV1(), mockServiceNameProvider, nil, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), spy.NewLogDummy())
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
	sut := nsbroker.NewFacade(scFakeClientset.ServicecatalogV1beta1(), nil, nil, nil, fixWorkingNs(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), spy.NewLogDummy())
	// WHEN
	_, err := sut.Exist(fixDestNs())
	// THEN
	assert.EqualError(t, err, "while checking if ServiceBroker [remote-env-broker] exists in the namespace [stage]: some error")

}

func fixRebSelectorKey() string {
	return "app"
}

func fixRebSelectorValue() string {
	return "reb"
}

func fixTargetPort() int32 {
	return int32(8080)
}

func fixDestNs() string {
	return "stage"
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

func assertPerformedActionCreateService(t *testing.T, actions []k8s_testing.Action) {
	assertPerformedAction(t, "create", "services", actions)
}

func assertPerformedActionRemoveService(t *testing.T, actions []k8s_testing.Action) {
	assertPerformedAction(t, "delete", "services", actions)
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
