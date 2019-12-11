package nsbroker_test

import (
	"fmt"
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc_fake "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	scbeta "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const integrationNS = "kyma-integration"

func TestMigrationServiceHappyPath(t *testing.T) {
	// GIVEN
	ns1 := "stage"
	ns2 := "production"

	ts := newTestSuite(t)
	// create namespaces
	ts.createNamespace(ns1)
	ts.createNamespace(ns2)
	ts.createNamespace(integrationNS)

	// create legacy setup (which will be migrated)
	ts.createLegacyBrokerAndService(ns1)
	ts.createLegacyBrokerAndService(ns2)

	//create other services, which must not be removed
	ts.createService("qa", "other-service")
	ts.createService(integrationNS, "other-service")
	ts.createService(integrationNS, "application-broker")

	// create other ServiceBroker which must not be touched
	ts.createServiceBroker("qa", "other-broker", "http://other-broker.kyma-integration.svc.cluster.local")

	ts.assertServicesInNamespace(integrationNS, "other-service", "application-broker")

	migrationService := ts.newMigrationService()

	// WHEN
	migrationService.Migrate()

	// THEN
	ts.assertServicesInNamespace(integrationNS, "other-service", "application-broker")
	ts.assertServicesInNamespace("qa", "other-service")
	ts.assertServicesInNamespace("production")

	ts.assertServiceBrokerURL("stage", nsbroker.NamespacedBrokerName, fmt.Sprintf("http://application-broker.%s.svc.cluster.local/stage", integrationNS))
	ts.assertServiceBrokerURL("production", nsbroker.NamespacedBrokerName, fmt.Sprintf("http://application-broker.%s.svc.cluster.local/production", integrationNS))
	ts.assertServiceBrokerURL("qa", "other-broker", "http://other-broker.kyma-integration.svc.cluster.local")

	// second execution of migration should not change anything
	// WHEN
	migrationService.Migrate()

	// THEN
	ts.assertServicesInNamespace(integrationNS, "other-service", "application-broker")
	ts.assertServicesInNamespace("qa", "other-service")

	ts.assertServiceBrokerURL("stage", nsbroker.NamespacedBrokerName, fmt.Sprintf("http://application-broker.%s.svc.cluster.local/stage", integrationNS))
	ts.assertServiceBrokerURL("production", nsbroker.NamespacedBrokerName, fmt.Sprintf("http://application-broker.%s.svc.cluster.local/production", integrationNS))
	ts.assertServiceBrokerURL("qa", "other-broker", "http://other-broker.kyma-integration.svc.cluster.local")
}

func TestMigrationServiceNoServiceDeletion(t *testing.T) {
	// GIVEN
	ns1 := "stage"

	ts := newTestSuite(t)
	// create namespaces
	ts.createNamespace(ns1)
	ts.createNamespace(integrationNS)
	ts.createService(integrationNS, "application-broker")

	ts.createABServiceBroker(ns1, "ec-prod", fmt.Sprintf("http://application-broker.kyma-integration.svc.cluster.local/%s", ns1))
	ts.createABServiceBroker(integrationNS, "ec-prod", "http://application-broker.kyma-integration.svc.cluster.local/")

	migrationService := ts.newMigrationService()

	// WHEN
	migrationService.Migrate()

	// THEN
	ts.assertServicesInNamespace(integrationNS, "application-broker")
	ts.assertServiceBrokerURL(ns1, "ec-prod", fmt.Sprintf("http://application-broker.kyma-integration.svc.cluster.local/%s", ns1))
}

type testSuite struct {
	brokerGetter       scbeta.ServiceBrokersGetter
	servicesGetter     typedCorev1.ServicesGetter
	namespaceInterface typedCorev1.NamespaceInterface

	legacyFacade *legacyFacade

	t *testing.T
}

func newTestSuite(t *testing.T) *testSuite {
	scFakeClientset := sc_fake.NewSimpleClientset()
	k8sFakeClientset := k8s_fake.NewSimpleClientset()
	nsBrokerService, err := broker.NewNsBrokerService()
	require.NoError(t, err)
	return &testSuite{
		servicesGetter:     k8sFakeClientset.CoreV1(),
		brokerGetter:       scFakeClientset.ServicecatalogV1beta1(),
		legacyFacade:       newLegacyFacade(scFakeClientset.ServicecatalogV1beta1(), k8sFakeClientset.CoreV1(), nsBrokerService, fixWorkingNs(), fixABSelectorKey(), fixABSelectorValue(), fixTargetPort()),
		namespaceInterface: k8sFakeClientset.CoreV1().Namespaces(),
		t:                  t,
	}
}

func (ts *testSuite) newMigrationService() *nsbroker.MigrationService {
	ms, err := nsbroker.NewMigrationService(ts.servicesGetter, ts.brokerGetter, integrationNS, "application-broker", spy.NewLogDummy())
	require.NoError(ts.t, err)
	return ms
}

func (ts *testSuite) createLegacyBrokerAndService(namespaceName string) {
	err := ts.legacyFacade.Create(namespaceName)
	require.NoError(ts.t, err)
}

func (ts *testSuite) createService(namespace, name string) {
	_, err := ts.servicesGetter.Services(namespace).Create(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	require.NoError(ts.t, err)
}

func (ts *testSuite) createNamespace(name string) {
	_, err := ts.namespaceInterface.Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	require.NoError(ts.t, err)
}

func (ts *testSuite) assertServicesInNamespace(namespace string, expectedServices ...string) {
	list, err := ts.servicesGetter.Services(namespace).List(metav1.ListOptions{})
	require.NoError(ts.t, err)
	assert.Equal(ts.t, len(expectedServices), len(list.Items))

	for _, s := range list.Items {
		assert.Contains(ts.t, expectedServices, s.Name)
	}
}
func (ts *testSuite) assertServiceBrokerURL(namespace string, name string, url string) {
	broker, err := ts.brokerGetter.ServiceBrokers(namespace).Get(name, metav1.GetOptions{})
	require.NoError(ts.t, err)

	assert.Equal(ts.t, url, broker.Spec.URL)
}

func (ts *testSuite) createServiceBroker(ns, name, url string) {
	_, err := ts.brokerGetter.ServiceBrokers(ns).Create(&v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	})
	require.NoError(ts.t, err)
}

func (ts *testSuite) createABServiceBroker(ns, name, url string) {
	_, err := ts.brokerGetter.ServiceBrokers(ns).Create(&v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				nsbroker.BrokerLabelKey: nsbroker.BrokerLabelValue,
			},
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	})
	require.NoError(ts.t, err)
}
