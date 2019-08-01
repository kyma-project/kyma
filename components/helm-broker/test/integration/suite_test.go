// +build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/helm-broker/internal/assetstore"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8s "k8s.io/client-go/kubernetes"
	kubernetes "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/kyma-project/kyma/components/helm-broker/internal/assetstore/automock"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bind"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/testdata"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

const (
	pollingInterval = 100 * time.Millisecond

	// redisAddonID is the ID of the bundle redis in testdata dir
	redisAddonID     = "id-09834-abcd-234"
	accTestAddonID   = "a54abe18-0a84-22e9-ab34-d663bbce3d88"
	addonsConfigName = "addons"

	redisAddonIDGit     = "91c753f0-813b-4bf0-a6b6-f682b1327a21"
	accTestAddonIDGit   = "6308335c-1ace-48ef-a253-47a5c31dd52c"
	addonsConfigNameGit = "git-addons"

	redisRepo           = "index-redis.yaml"
	accTestRepo         = "index-acc-testing.yaml"
	redisAndAccTestRepo = "index.yaml"
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
	logger := logrus.New()

	brokerServer := broker.New(sFact.Addon(), sFact.Chart(), sFact.InstanceOperation(), sFact.Instance(), sFact.InstanceBindData(),
		bind.NewRenderer(), bind.NewResolver(k8sClientset.CoreV1()), nil, logger.WithField("svc", "broker"))

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

	uploadClient := &automock.Client{}
	uploadClient.On("Upload", mock.AnythingOfType("string"), mock.Anything).Return(assetstore.UploadedFile{}, nil)

	mgr := controller.SetupAndStartController(restConfig, &config.ControllerConfig{
		DevelopMode:              true, // DevelopMode allows "http" urls
		ClusterServiceBrokerName: "helm-broker",
		TmpDir:                   cfg.TmpDir,
	}, ":8001", sFact, uploadClient, logger.WithField("svc", "broker"))

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

func (ts *testSuite) assertNoServicesInCatalogEndpoint(prefix string) {
	osbClient, err := newOSBClient(fmt.Sprintf("%s/%s", ts.server.URL, prefix))
	require.NoError(ts.t, err)
	resp, err := osbClient.GetCatalog()
	require.NoError(ts.t, err)

	assert.Empty(ts.t, resp.Services)
}

func (ts *testSuite) waitForEmptyCatalogResponse(prefix string) {
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

func (ts *testSuite) waitForServicesInCatalogEndpoint(prefix string, ids []string) {
	osbClient, err := newOSBClient(fmt.Sprintf("%s/%s", ts.server.URL, prefix))
	require.NoError(ts.t, err)

	timeoutCh := time.After(3 * time.Second)
	for {
		err := ts.checkServiceIDs(osbClient, ids)
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

func (ts *testSuite) checkServiceIDs(osbClient osb.Client, ids []string) error {
	osbResponse, err := osbClient.GetCatalog()
	if err != nil {
		return err
	}

	idsToCheck := make(map[string]struct{})
	for _, id := range ids {
		idsToCheck[id] = struct{}{}
	}

	for _, service := range osbResponse.Services {
		delete(idsToCheck, service.ID)
	}

	if len(idsToCheck) > 0 {
		return fmt.Errorf("unexpected GetCatalogResponse, missing services: %v", idsToCheck)
	}

	return nil
}

func (ts *testSuite) waitForClusterAddonsConfigurationPhase(name string, expectedPhase v1alpha1.AddonsConfigurationPhase) {
	var cac v1alpha1.ClusterAddonsConfiguration
	ts.waitForPhase(&cac, &(cac.Status.CommonAddonsConfigurationStatus), types.NamespacedName{Name: name}, expectedPhase)
}

func (ts *testSuite) waitForAddonsConfigurationPhase(namespace, name string, expectedPhase v1alpha1.AddonsConfigurationPhase) {
	var ac v1alpha1.AddonsConfiguration
	ts.waitForPhase(&ac, &(ac.Status.CommonAddonsConfigurationStatus), types.NamespacedName{Name: name, Namespace: namespace}, expectedPhase)
}

func (ts *testSuite) waitForPhase(obj runtime.Object, status *v1alpha1.CommonAddonsConfigurationStatus, nn types.NamespacedName, expectedPhase v1alpha1.AddonsConfigurationPhase) {
	timeoutCh := time.After(3 * time.Second)
	for {
		err := ts.dynamicClient.Get(context.TODO(), nn, obj)
		require.NoError(ts.t, err)

		if status.Phase == expectedPhase {
			return
		}

		select {
		case <-timeoutCh:
			assert.Fail(ts.t, fmt.Sprintf("The timeout exceeded while waiting for the Phase %s (%q), current phase: %s", expectedPhase, nn.String(), string(status.Phase)))
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (ts *testSuite) deleteAddonsConfiguration(namespace, name string) {
	require.NoError(ts.t, ts.dynamicClient.Delete(context.TODO(), &v1alpha1.AddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		}}))
}

func (ts *testSuite) deleteClusterAddonsConfiguration(name string) {
	require.NoError(ts.t, ts.dynamicClient.Delete(context.TODO(), &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		}}))
}

func (ts *testSuite) createAddonsConfiguration(namespace, name string, source *repositorySource) {
	err := ts.dynamicClient.Create(context.TODO(), &v1alpha1.AddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: source.generateAddonRepositories(),
			},
		},
	})

	if err != nil {
		ts.t.Logf("Failed during creating AddonsConfiguration: %s", err)
	}
}

func (ts *testSuite) createClusterAddonsConfiguration(name string, source *repositorySource) {
	err := ts.dynamicClient.Create(context.TODO(), &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: source.generateAddonRepositories(),
			},
		},
	})

	if err != nil {
		ts.t.Logf("Failed during creating ClusterAddonsConfiguration: %s", err)
	}
}

func (ts *testSuite) updateAddonsConfigurationRepositories(namespace, name string, source *repositorySource) {
	var addonsConfiguration v1alpha1.AddonsConfiguration
	require.NoError(ts.t, ts.dynamicClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, &addonsConfiguration))

	addonsConfiguration.Spec.Repositories = source.generateAddonRepositories()
	require.NoError(ts.t, ts.dynamicClient.Update(context.TODO(), &addonsConfiguration))
}

func (ts *testSuite) updateClusterAddonsConfigurationRepositories(name string, source *repositorySource) {
	var clusterAddonsConfiguration v1alpha1.ClusterAddonsConfiguration
	require.NoError(ts.t, ts.dynamicClient.Get(context.TODO(), types.NamespacedName{Name: name}, &clusterAddonsConfiguration))

	clusterAddonsConfiguration.Spec.Repositories = source.generateAddonRepositories()
	require.NoError(ts.t, ts.dynamicClient.Update(context.TODO(), &clusterAddonsConfiguration))
}
