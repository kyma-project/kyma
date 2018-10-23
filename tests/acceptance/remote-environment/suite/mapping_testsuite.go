package suite

import (
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/acceptance/pkg/repeat"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"testing"
	"time"
)

type MappingTestSuite struct {
	// TestID is a short id used as suffix in resource names
	TestID   string
	scClient v1beta1.ServicecatalogV1beta1Interface
	reClient v1alpha1.ApplicationconnectorV1alpha1Interface
	nsClient v1.NamespaceInterface

	osbServiceId          string
	remoteEnvironmentName string
	brokerName            string

	MappedNs string
	EmptyNs  string

	t *testing.T
}

func NewMappingTestSuite(t *testing.T) *MappingTestSuite {
	config, err := restclient.InClusterConfig()
	require.NoError(t, err)
	scClient, err := clientset.NewForConfig(config)
	require.NoError(t, err)
	reClient, err := v1alpha1.NewForConfig(config)
	require.NoError(t, err)
	k8sClient, err := v1.NewForConfig(config)
	require.NoError(t, err)

	id := rand.String(4)
	return &MappingTestSuite{
		TestID:   id,
		scClient: scClient.ServicecatalogV1beta1(),
		nsClient: k8sClient.Namespaces(),
		reClient: reClient,

		osbServiceId:          fmt.Sprintf("acc-test-osb-serviceid-%s", id),
		remoteEnvironmentName: fmt.Sprintf("acc-test-re-%s", id),

		brokerName: "remote-env-broker",

		MappedNs: fmt.Sprintf("acc-test-mapping-ns-mapped-%s", id),
		EmptyNs:  fmt.Sprintf("acc-test-mapping-ns-empty-%s", id),

		t: t,
	}
}

func (ts *MappingTestSuite) Setup() {
	ts.createNamespaces()
	ts.createRemoteEnvironment()
}

func (ts *MappingTestSuite) TearDown() {
	err := ts.reClient.RemoteEnvironments().Delete(ts.remoteEnvironmentName, &metav1.DeleteOptions{})
	assert.NoError(ts.t, err)
	err = ts.nsClient.Delete(ts.MappedNs, &metav1.DeleteOptions{})
	assert.NoError(ts.t, err)
	err = ts.nsClient.Delete(ts.EmptyNs, &metav1.DeleteOptions{})
	assert.NoError(ts.t, err)
}

func (ts *MappingTestSuite) WaitForServiceClassWithTimeout(timeout time.Duration) {
	repeat.FuncAtMost(ts.t, func() error {
		_, err := ts.scClient.ServiceClasses(ts.MappedNs).Get(ts.osbServiceId, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "error while getting service class %s", ts.osbServiceId)
		}
		return nil
	}, timeout)
}

func (ts *MappingTestSuite) WaitForServiceBrokerWithTimeout(timeout time.Duration) {
	repeat.FuncAtMost(ts.t, func() error {
		_, err := ts.scClient.ServiceBrokers(ts.MappedNs).Get(ts.brokerName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "error while getting service broker %s", ts.brokerName)
		}
		return nil
	}, timeout)
}

func (ts *MappingTestSuite) EnsureServiceBrokerNotExistWithTimeout(timeout time.Duration) {
	repeat.FuncAtMost(ts.t, func() error {
		_, err := ts.scClient.ServiceBrokers(ts.MappedNs).Get(ts.brokerName, metav1.GetOptions{})
		switch {
		case apierrors.IsNotFound(err):
			return nil
		case err == nil:
			return fmt.Errorf("service broker [%s] in namespace [%s] should not exist", ts.brokerName, ts.MappedNs)
		}
		return errors.Wrapf(err, "error while getting service broker %s", ts.brokerName)
	}, timeout)
}

func (ts *MappingTestSuite) EnsureServiceClassNotExist(namespace string) {
	_, err := ts.scClient.ServiceClasses(namespace).Get(ts.osbServiceId, metav1.GetOptions{})
	assert.True(ts.t, apierrors.IsNotFound(err))
}

func (ts *MappingTestSuite) EnsureServiceBrokerNotExist(namespace string) {
	_, err := ts.scClient.ServiceBrokers(namespace).Get(ts.brokerName, metav1.GetOptions{})
	assert.True(ts.t, apierrors.IsNotFound(err))
}

func (ts *MappingTestSuite) CreateEnvironmentMapping() {
	ts.t.Log("Create EnvironmentMapping")
	_, err := createEnvironmentMapping(ts.reClient.EnvironmentMappings(ts.MappedNs), ts.remoteEnvironmentName)
	require.NoError(ts.t, err)
}

func (ts *MappingTestSuite) DeleteEnvironmentMapping() {
	ts.t.Log("Delete EnvironmentMapping")
	err := ts.reClient.EnvironmentMappings(ts.MappedNs).Delete(ts.remoteEnvironmentName, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func (ts *MappingTestSuite) createNamespaces() {
	ts.t.Logf("Create Namespace %s", ts.MappedNs)
	_, err := ts.nsClient.Create(fixNamespace(ts.MappedNs))
	require.NoError(ts.t, err)

	ts.t.Logf("Create Namespace %s", ts.EmptyNs)
	_, err = ts.nsClient.Create(fixNamespace(ts.EmptyNs))
	require.NoError(ts.t, err)
}

func (ts *MappingTestSuite) createRemoteEnvironment() {
	ts.t.Log("Create RemoteEnvironment")
	displayName := fmt.Sprintf("acc-test-re-name-%s", ts.TestID)
	_, err := createRemoteEnvironment(ts.reClient.RemoteEnvironments(), ts.remoteEnvironmentName, "dummy", ts.osbServiceId, "dummy", displayName)
	require.NoError(ts.t, err)
}
