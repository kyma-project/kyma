package gateway

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	serviceCatalogApi "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"

	proxyconfig2 "github.com/kyma-project/kyma/components/application-gateway/pkg/proxyconfig"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/gateway/testkit/proxyconfig"

	"github.com/google/uuid"

	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/gateway/mock"
	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
)

const (
	defaultCheckInterval         = 2 * time.Second
	appGatewayHealthCheckTimeout = 45 * time.Second
	gatewayConnectionTimeout     = 20 * time.Second
	apiServerAccessTimeout       = 60 * time.Second
)

type TestSuite struct {
	httpClient            *http.Client
	k8sClient             *kubernetes.Clientset
	namespaceClient       corev1.NamespaceInterface
	secretClient          corev1.SecretInterface
	secretCreator         *proxyconfig.SecretsCreator
	serviceInstanceClient serviceCatalogClient.ServiceInstanceInterface
	config                TestConfig
	appMockServer         *mock.AppMockServer

	serviceInstanceName string
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := ReadConfig()
	require.NoError(t, err)

	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	require.NoError(t, err)

	appMockServer := mock.NewAppMockServer(config.MockServicePort)

	secretClient := coreClientset.CoreV1().Secrets(config.GatewayNamespace)

	secretsCreator := proxyconfig.NewSecretsCreator(secretClient)

	svcCatalogueClientSet, err := serviceCatalogClient.NewForConfig(k8sConfig)
	require.NoError(t, err)

	serviceInstanceName := fmt.Sprintf("gateway-tests-instance-%s", uuid.New().String()[:4])

	return &TestSuite{
		httpClient:            &http.Client{},
		k8sClient:             coreClientset,
		namespaceClient:       coreClientset.CoreV1().Namespaces(),
		secretClient:          secretClient,
		secretCreator:         secretsCreator,
		config:                config,
		appMockServer:         appMockServer,
		serviceInstanceClient: svcCatalogueClientSet.ServiceInstances(config.GatewayNamespace),
		serviceInstanceName:   serviceInstanceName,
	}
}

func (ts *TestSuite) Setup(t *testing.T) {
	ts.WaitForAccessToAPIServer(t)

	ts.appMockServer.Start()

	err := ts.createNamespace()
	require.NoError(t, err)

	err = ts.createServiceInstance()
	require.NoError(t, err)

	ts.CheckApplicationGatewayHealth(t)
}

func (ts *TestSuite) Cleanup(t *testing.T) {
	t.Log("Calling cleanup")

	err := ts.appMockServer.Kill()
	assert.NoError(t, err)

	err = ts.serviceInstanceClient.Delete(ts.serviceInstanceName, &metav1.DeleteOptions{})
	assert.NoError(t, err)

	err = ts.namespaceClient.Delete(ts.config.GatewayNamespace, &metav1.DeleteOptions{})
	assert.NoError(t, err)
}

func (ts *TestSuite) createNamespace() error {
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.config.GatewayNamespace,
		},
	}

	_, err := ts.namespaceClient.Create(namespace)
	return err
}

func (ts *TestSuite) createServiceInstance() error {
	serviceInstance := &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{Name: ts.serviceInstanceName},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: serviceCatalogApi.PlanReference{
				ServiceClassExternalName: "781757f2-8a04-42c0-9015-62905b0f5431",
				ServicePlanExternalName:  "default",
			},
		},
	}

	_, err := ts.serviceInstanceClient.Create(serviceInstance)
	return err
}

// WaitForAccessToAPIServer waits for access to API Server which might be delayed by initialization of Istio sidecar
func (ts *TestSuite) WaitForAccessToAPIServer(t *testing.T) {
	err := tools.WaitForFunction(defaultCheckInterval, apiServerAccessTimeout, func() bool {
		t.Log("Trying to access API Server...")
		_, err := ts.k8sClient.ServerVersion()
		if err != nil {
			t.Log(err.Error())
			return false
		}

		return true
	})

	require.NoError(t, err)
}

func (ts *TestSuite) CheckApplicationGatewayHealth(t *testing.T) {
	t.Log("Checking application gateway health...")

	healthURL := ts.gatewayExternalAPIURL() + "/v1/health"
	err := tools.WaitForFunction(defaultCheckInterval, appGatewayHealthCheckTimeout, func() bool {
		req, err := http.NewRequest(http.MethodGet, healthURL, nil)
		if err != nil {
			return false
		}

		res, err := ts.httpClient.Do(req)
		if err != nil {
			return false
		}

		if res.StatusCode != http.StatusOK {
			return false
		}

		return true
	})

	require.NoError(t, err, "Failed to check health of Application Gateway.")
}

func (ts *TestSuite) CreateSecret(t *testing.T, apiName string, proxyConfig proxyconfig2.ProxyDestinationConfig) string {
	secretName := fmt.Sprintf("test-%s-%s", ts.config.GatewayNamespace, apiName)

	err := ts.secretCreator.NewSecret(secretName, apiName, proxyConfig)
	require.NoError(t, err)

	return secretName
}

func (ts *TestSuite) DeleteSecret(t *testing.T, secretName string) {
	err := ts.secretClient.Delete(secretName, &metav1.DeleteOptions{})
	assert.NoError(t, err)
}

func (ts *TestSuite) CallAPIThroughGateway(t *testing.T, secretName, apiName, path string) *http.Response {
	url := fmt.Sprintf("%s/secret/%s/api/%s/%s", ts.gatewayProxyURL(), secretName, apiName, path)

	var resp *http.Response

	err := tools.WaitForFunction(defaultCheckInterval, gatewayConnectionTimeout, func() bool {
		t.Logf("Accessing Gateway at: %s", url)
		var err error

		resp, err = http.Get(url)
		if err != nil {
			t.Logf("Failed to access Gateway: %s", err.Error())
			return false
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusServiceUnavailable {
			t.Logf("Invalid response from Gateway, status: %d.", resp.StatusCode)
			bytes, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			t.Log(string(bytes))
			t.Logf("Gateway is not ready. Retrying.")
			return false
		}

		return true
	})
	require.NoError(t, err)

	return resp
}

func (ts *TestSuite) gatewayExternalAPIURL() string {
	return fmt.Sprintf("http://%s-gateway.%s.svc.cluster.local:8081", ts.config.GatewayNamespace, ts.config.GatewayNamespace)
}

func (ts *TestSuite) gatewayProxyURL() string {
	return fmt.Sprintf("http://%s-gateway.%s.svc.cluster.local:8080", ts.config.GatewayNamespace, ts.config.GatewayNamespace)
}

func (ts *TestSuite) GetMockServiceURL() string {
	return ts.config.MockServiceURL
}
