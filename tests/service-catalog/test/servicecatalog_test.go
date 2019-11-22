// +build acceptance

package test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	mappingClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/service-catalog/utils/istio"
	osb "github.com/pmorie/go-open-service-broker-client/v2"

	scc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	v1alpha12 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/tests/service-catalog/utils/repeat"
	"github.com/kyma-project/kyma/tests/service-catalog/utils/report"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	apimerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	helmBrokerURLEnvName                = "HELM_BROKER_URL"
	applicationBrokerURLEnvName         = "APPLICATION_BROKER_URL"
	istioAuthorizationRuleEnabledEnvVar = "BROKER_ISTIO_RBAC_ENABLED"

	addonId       = "a54abe18-0a84-22e9-ab34-d663bbce3d88"
	addonsRepoURL = "https://github.com/kyma-project/bundles/releases/download/latest/index-acc-testing.yaml"

	anyNamespaceName = "default"
)

func TestBrokerHasIstioRbacAuthorizationRules(t *testing.T) {
	if !(os.Getenv(istioAuthorizationRuleEnabledEnvVar) == "true") {
		t.Skipf("Test skipped due to %s environment variable value.", istioAuthorizationRuleEnabledEnvVar)
	}
	type brokerInfo struct {
		url           string
		labelSelector string
	}

	for testName, broker := range map[string]brokerInfo{
		"Helm Broker":        {url: os.Getenv(helmBrokerURLEnvName), labelSelector: "app=helm-broker"},
		"Application Broker": {url: os.Getenv(applicationBrokerURLEnvName), labelSelector: "app=application-broker"},
		// "<next broker>": os.Getenv("<broker url env name>"),
	} {
		t.Run(testName, func(t *testing.T) {
			err := repeat.FuncAtMost(func() error {
				isForbidden, statusCode, err := isCatalogForbidden(broker.url)
				if err != nil {
					return errors.Wrap(err, "while getting catalog")
				}
				if !isForbidden {
					return fmt.Errorf("%s catalog response must be forbidden, status code: %d", testName, statusCode)
				}
				return nil
			}, time.Minute)
			assert.NoError(t, err)

			if err != nil || os.Getenv("DUMP_ISTIO_CONFIG") == "true" {
				u, err := url.Parse(broker.url)
				require.NoError(t, err)

				istio.DumpConfig(t, u.Host, broker.labelSelector)

			}
		})
	}
}

func TestHelmBrokerAddonsConfiguration(t *testing.T) {
	// given
	namespace := fmt.Sprintf("test-acc-addonsconfig-%s", rand.String(4))
	addonsConfig := &v1alpha12.AddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testing-addons-config",
			Namespace: namespace,
		},
		Spec: v1alpha12.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha12.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha12.SpecRepository{{URL: addonsRepoURL}},
			},
		},
	}

	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)

	sch, err := v1alpha12.SchemeBuilder.Build()
	require.NoError(t, err)

	dynamicClient, err := client.New(k8sConfig, client.Options{Scheme: sch})
	require.NoError(t, err)

	k8sClient, err := corev1.NewForConfig(k8sConfig)
	require.NoError(t, err)

	scClient, err := scc.NewForConfig(k8sConfig)
	require.NoError(t, err)

	t.Logf("Creating Namespace %s", namespace)
	_, err = k8sClient.Namespaces().Create(fixNamespace(namespace))
	require.NoError(t, err)

	defer func() {
		err = k8sClient.Namespaces().Delete(namespace, &metav1.DeleteOptions{})
		assert.NoError(t, err)
	}()

	// when
	err = dynamicClient.Create(context.TODO(), addonsConfig)
	require.NoError(t, err)
	defer func() {
		if t.Failed() {
			namespaceReport := report.NewReport(t,
				k8sConfig,
				report.WithHB(), report.WithSC())
			namespaceReport.PrintJsonReport(namespace)
		}
		err = dynamicClient.Delete(context.TODO(), addonsConfig)
		assert.NoError(t, err)
	}()
	// then
	repeat.AssertFuncAtMost(t, func() error {
		_, err := scClient.ServicecatalogV1beta1().ServiceClasses(namespace).Get(addonId, metav1.GetOptions{})
		return err
	}, time.Second*20)
}

func TestHelmBrokerClusterAddonsConfiguration(t *testing.T) {
	// given
	randomID := rand.String(4)
	addonsConfig := &v1alpha12.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("acceptance-test-addonsconfig-%s", randomID),
		},
		Spec: v1alpha12.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha12.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha12.SpecRepository{{URL: addonsRepoURL}},
			},
		},
	}

	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)

	sch, err := v1alpha12.SchemeBuilder.Build()
	require.NoError(t, err)

	dynamicClient, err := client.New(k8sConfig, client.Options{Scheme: sch})
	require.NoError(t, err)

	scClient, err := scc.NewForConfig(k8sConfig)
	require.NoError(t, err)

	// when
	err = dynamicClient.Create(context.TODO(), addonsConfig)
	require.NoError(t, err)

	defer func() {
		if t.Failed() {
			namespaceReport := report.NewReport(t,
				k8sConfig,
				report.WithHB())
			namespaceReport.PrintJsonReport(anyNamespaceName)
		}
		err = dynamicClient.Delete(context.TODO(), addonsConfig)
		assert.NoError(t, err)
	}()

	// then
	repeat.AssertFuncAtMost(t, func() error {
		_, err := scClient.ServicecatalogV1beta1().ClusterServiceClasses().Get(addonId, metav1.GetOptions{})
		return err
	}, time.Second*30)
}

func TestServiceCatalogResourcesAreCleanUp(t *testing.T) {
	// given
	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)

	aClient, err := appClient.NewForConfig(k8sConfig)
	require.NoError(t, err)

	mClient, err := mappingClient.NewForConfig(k8sConfig)
	require.NoError(t, err)

	k8sClient, err := corev1.NewForConfig(k8sConfig)
	require.NoError(t, err)

	scClient, err := scc.NewForConfig(k8sConfig)
	require.NoError(t, err)

	name := fmt.Sprintf("test-acc-app-%s", rand.String(4))
	namespace := fmt.Sprintf("test-acc-ns-broker-%s", rand.String(4))

	defer func() {
		if t.Failed() {
			namespaceReport := report.NewReport(t,
				k8sConfig,
				report.WithK8s(),
				report.WithApp(),
				report.WithSC())
			namespaceReport.PrintJsonReport(namespace)
		}
	}()

	t.Log("Creating Application")
	app, err := aClient.ApplicationconnectorV1alpha1().Applications().Create(fixApplication(name))
	require.NoError(t, err)

	t.Logf("Creating Namespace %s", namespace)
	_, err = k8sClient.Namespaces().Create(fixNamespace(namespace))
	require.NoError(t, err)

	t.Log("Creating ApplicationMapping")
	am, err := mClient.ApplicationconnectorV1alpha1().ApplicationMappings(namespace).Create(fixApplicationMapping(name))
	require.NoError(t, err)
	// when
	t.Logf("Removing Namespace %s", namespace)
	err = k8sClient.Namespaces().Delete(namespace, &metav1.DeleteOptions{})
	assert.NoError(t, err)

	t.Log("Removing Application")
	err = aClient.ApplicationconnectorV1alpha1().Applications().Delete(app.Name, &metav1.DeleteOptions{})
	assert.NoError(t, err)

	t.Log("Assert namespace is removed")
	repeat.AssertFuncAtMost(t, func() error {
		_, err := k8sClient.Namespaces().Get(namespace, metav1.GetOptions{})
		if apimerr.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("%s is not removed", namespace)
	}, time.Minute*2)

	// then
	t.Log("Check resources are removed")
	_, err = aClient.ApplicationconnectorV1alpha1().Applications().Get(app.Name, metav1.GetOptions{})
	assert.True(t, apimerr.IsNotFound(err))

	_, err = mClient.ApplicationconnectorV1alpha1().ApplicationMappings(namespace).Get(am.Name, metav1.GetOptions{})
	assert.True(t, apimerr.IsNotFound(err))

	listServices, err := k8sClient.Services(namespace).List(metav1.ListOptions{})
	assert.Empty(t, listServices.Items)

	listServiceBrokers, err := scClient.ServicecatalogV1beta1().ServiceBrokers(namespace).List(metav1.ListOptions{})
	assert.Empty(t, listServiceBrokers.Items)
}

func isCatalogForbidden(url string) (bool, int, error) {
	config := osb.DefaultClientConfiguration()
	config.URL = fmt.Sprintf("%s/cluster", url)

	client, err := osb.NewClient(config)
	if err != nil {
		return false, 0, errors.Wrapf(err, "while creating osb client for broker with URL: %s", url)
	}
	var statusCode int
	isForbiddenError := func(err error) bool {
		statusCodeError, ok := err.(osb.HTTPStatusCodeError)
		if !ok {
			return false
		}
		statusCode = statusCodeError.StatusCode
		return statusCodeError.StatusCode == http.StatusForbidden
	}

	_, err = client.GetCatalog()
	switch {
	case err == nil:
		return false, http.StatusOK, nil
	case isForbiddenError(err):
		return true, statusCode, nil
	default:
		return false, 0, errors.Wrapf(err, "while getting catalog from broker with URL: %s", url)
	}
}

func fixNamespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func fixApplication(name string) *appTypes.Application {
	return &appTypes.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appTypes.ApplicationSpec{
			Description:      "Application used by acceptance test",
			AccessLabel:      "fix-access",
			SkipInstallation: true,
			Services: []appTypes.Service{
				{
					ID:   "id-00000-1234-test",
					Name: "provider-4951",
					Labels: map[string]string{
						"connected-app": name,
					},
					ProviderDisplayName: "provider",
					DisplayName:         name,
					Description:         "Application Service Class used by acceptance test",
					Tags:                []string{},
					Entries: []appTypes.Entry{
						{
							Type:        "API",
							AccessLabel: "acc-label",
							GatewayUrl:  "http://promotions-gateway.production.svc.cluster.local/",
						},
					},
				},
			},
		},
	}
}

func fixApplicationMapping(name string) *mappingTypes.ApplicationMapping {
	return &mappingTypes.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
