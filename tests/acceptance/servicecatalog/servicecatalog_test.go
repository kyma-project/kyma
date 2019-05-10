package servicecatalog_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	mappingClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"

	scc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	corev1 "github.com/kubernetes/client-go/kubernetes/typed/core/v1"
	"github.com/kyma-project/kyma/tests/acceptance/pkg/repeat"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apimerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	restclient "k8s.io/client-go/rest"
)

const (
	helmBrokerURLEnvName    = "HELM_BROKER_URL"
	releaseNamespaceEnvName = "RELEASE_NAMESPACE"
)

func TestBrokerHasIstioRbacAuthorizationRules(t *testing.T) {
	for testName, brokerURL := range map[string]string{
		"Helm Broker": os.Getenv(helmBrokerURLEnvName),
		// "<next broker>": os.Getenv("<broker url env name>"),
	} {
		t.Run(testName, func(t *testing.T) {
			repeat.FuncAtMost(t, func() error {
				isForbidden, err := isCatalogForbidden(brokerURL)
				if err != nil {
					return errors.Wrap(err, "while getting catalog")
				}
				if !isForbidden {
					return fmt.Errorf("%s catalog response must be forbidden", testName)
				}
				return nil
			}, 5*time.Second)
		})
	}
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
	repeat.FuncAtMost(t, func() error {
		_, err := k8sClient.Namespaces().Get(namespace, metav1.GetOptions{})
		if apimerr.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("%s is not removed", namespace)
	}, time.Second*90)

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

type brokerURL struct {
	prefix    string
	namespace string
}

func (b *brokerURL) buildURL(releaseNS string) string {
	return fmt.Sprintf("http://%s%s.%s.svc.cluster.local", b.prefix, b.namespace, releaseNS)
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
