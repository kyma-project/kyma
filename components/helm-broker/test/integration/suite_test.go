// +build integration

package integration_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"context"
	"os"

	"github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/testdata"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8s "k8s.io/client-go/kubernetes"
	kubernetes "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"github.com/sirupsen/logrus"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bind"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
)

const (
	pollingInterval = 100 * time.Millisecond

	// bundleID is the ID of the bundle redis in testdata dir
	bundleID = "id-09834-abcd-234"
	addonsConfigName = "addons"
)

func newTestSuite(t *testing.T) *testSuite {
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, apis.AddToScheme(sch))
	require.NoError(t, v1beta1.AddToScheme(sch))
	require.NoError(t, corev1.AddToScheme(sch))

	k8sClientset := kubernetes.NewSimpleClientset()

	cfg := &config.Config{
		TmpDir:                   os.TempDir(),
		Namespace:                "kyma-system",
		Storage:                  testdata.GoldenConfigMemorySingleAll(),
		DevelopMode:              true,
		ClusterServiceBrokerName: "helm-broker",
	}
	storageConfig := storage.ConfigList(cfg.Storage)
	sFact, err := storage.NewFactory(&storageConfig)
	require.NoError(t, err)

	brokerServer := broker.New(sFact.Bundle(), sFact.Chart(), sFact.InstanceOperation(), sFact.Instance(), sFact.InstanceBindData(),
		bind.NewRenderer(), bind.NewResolver(k8sClientset.CoreV1()), nil, logrus.New().WithField("service", "broker"))

	// OSB API Server
	server := httptest.NewServer(brokerServer.CreateHandler())

	// server with addons repository
	repoServer := httptest.NewServer(http.FileServer(http.Dir("testdata")))

	// setup and start kube-apiserver
	environment := &envtest.Environment{}
	restConfig, err := environment.Start()
	require.NoError(t, err)
	_, err = envtest.InstallCRDs(restConfig, envtest.CRDInstallOptions{
		Paths:              []string{"crds/"},
		ErrorIfPathMissing: true,
	})
	require.NoError(t, err)


	mgr := controller.SetupAndStartController(restConfig, &config.ControllerConfig{
		DevelopMode: true, // DevelopMode allows "http" urls
	}, ":8001", sFact, logrus.New().WithField("", ""))

	stopCh := make(chan struct{})
	go func() {
		if err := mgr.Start(stopCh); err != nil {
			t.Errorf("Controller Manager could not start: %v", err.Error())
		}
	}()

	// create a client for managing (cluster) addons configurations
	dynamicClient, err := client.New(restConfig, client.Options{Scheme: sch})

	return &testSuite{
		t: t,

		dynamicClient: dynamicClient,
		repoServer:    repoServer,
		server:        server,
		k8sClient:     k8sClientset,

		stopCh: stopCh,
	}
}

func newOSBClient(url string) (osb.Client, error) {
	config := osb.DefaultClientConfiguration()
	config.URL = url
	config.APIVersion = osb.Version2_13()

	osbClient, err := osb.NewClient(config)
	if err != nil {
		return nil, err
	}

	return osbClient, nil
}

type testSuite struct {
	t          *testing.T
	server     *httptest.Server
	repoServer *httptest.Server

	osbClient     osb.Client
	dynamicClient client.Client
	k8sClient     k8s.Interface

	stopCh chan struct{}
}

func (ts *testSuite) tearDown() {
	ts.server.Close()
	ts.repoServer.Close()
	close(ts.stopCh)
}

func (ts *testSuite) AssertNoServicesInCatalogEndpoint(prefix string) {
	osbClient, err := newOSBClient(fmt.Sprintf("%s/%s", ts.server.URL, prefix))
	require.NoError(ts.t, err)
	resp, err := osbClient.GetCatalog()
	require.NoError(ts.t, err)

	assert.Empty(ts.t, resp.Services)
}

func (ts *testSuite) WaitForEmptyCatalogResponse(prefix string) {
	osbClient, err := newOSBClient(fmt.Sprintf("%s/%s", ts.server.URL, prefix))
	require.NoError(ts.t, err)

	timeoutCh := time.After(3 * time.Second)
	for {
		resp, err := osbClient.GetCatalog()
		require.NoError(ts.t, err)
		if len(resp.Services) == 0 {
			return
		}

		select {
		case <-timeoutCh:
			assert.Fail(ts.t, "The timeout exceeded while waiting for the expected empty OSB catalog response")
			return
		default:
			time.Sleep(pollingInterval)
		}
	}
}

func (ts *testSuite) WaitForServicesInCatalogEndpoint(prefix string) {
	osbClient, err := newOSBClient(fmt.Sprintf("%s/%s", ts.server.URL, prefix))
	require.NoError(ts.t, err)

	timeoutCh := time.After(3 * time.Second)
	for {
		err := ts.checkServiceID(osbClient)
		if err == nil {
			return
		}
		select {
		case <-timeoutCh:
			assert.Failf(ts.t, "The timeout exceeded while waiting for the OSB catalog response, last error: %s", err.Error())
			return
		default:
			time.Sleep(pollingInterval)
		}
	}
}

func (ts *testSuite) checkServiceID(osbClient osb.Client) error {
	osbResponse, err := osbClient.GetCatalog()
	if err != nil {
		return err
	}

	if len(osbResponse.Services) == 1 && osbResponse.Services[0].ID == bundleID {
		return nil
	}

	return fmt.Errorf("unexpected GetCatalogResponse %v", osbResponse)
}

func (ts *testSuite) createClusterAddonsConfiguration() {
	ts.dynamicClient.Create(context.TODO(), &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name: "addons",
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{URL: ts.repoServer.URL + "/index.yaml"},
				},
			},
		},
	})
}

func (ts *testSuite) WaitForClusterAddonsConfigurationStatusReady() {
	var cac v1alpha1.ClusterAddonsConfiguration
	ts.waitForReady(&cac, &(cac.Status.CommonAddonsConfigurationStatus), types.NamespacedName{Name: "addons"})
}

func (ts *testSuite) WaitForAddonsConfigurationStatusReady(namespace string) {
	var ac v1alpha1.AddonsConfiguration
	ts.waitForReady(&ac, &(ac.Status.CommonAddonsConfigurationStatus), types.NamespacedName{Name: "addons", Namespace: namespace})
}

func (ts *testSuite) waitForReady(obj runtime.Object, status *v1alpha1.CommonAddonsConfigurationStatus, nn types.NamespacedName) {
	timeoutCh := time.After(3 * time.Second)
	for {
		err := ts.dynamicClient.Get(context.TODO(), nn, obj)
		require.NoError(ts.t, err)

		if status.Phase == v1alpha1.AddonsConfigurationReady {
			return
		}

		select {
		case <-timeoutCh:
			assert.Fail(ts.t, "The timeout exceeded while waiting for the Phase Ready, current phase: ", string(status.Phase))
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (ts *testSuite) createAddonsConfiguration(namespace string) {
	ts.dynamicClient.Create(context.TODO(), &v1alpha1.AddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name:      "addons",
			Namespace: namespace,
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{URL: ts.repoServer.URL + "/index.yaml"},
				},
			},
		},
	})
}

func (ts *testSuite) removeRepoFromAddonsConfiguration(namespace string) {
	var addonsConfiguration v1alpha1.AddonsConfiguration
	require.NoError(ts.t, ts.dynamicClient.Get(context.TODO(), types.NamespacedName{Name: "addons", Namespace: namespace}, &addonsConfiguration))

	addonsConfiguration.Spec.Repositories = []v1alpha1.SpecRepository{}

	require.NoError(ts.t, ts.dynamicClient.Update(context.TODO(), &addonsConfiguration))
}

func (ts *testSuite) removeRepoFromClusterAddonsConfiguration(namespace string) {
	var clusterAddonsConfiguration v1alpha1.ClusterAddonsConfiguration

	require.NoError(ts.t, ts.dynamicClient.Get(context.TODO(), types.NamespacedName{Name: addonsConfigName}, &clusterAddonsConfiguration))

	clusterAddonsConfiguration.Spec.Repositories = []v1alpha1.SpecRepository{}

	require.NoError(ts.t, ts.dynamicClient.Update(context.TODO(), &clusterAddonsConfiguration))
}

// todo: remove it
type dummyBroker struct {
}

func (*dummyBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"services":[]}`))
}
