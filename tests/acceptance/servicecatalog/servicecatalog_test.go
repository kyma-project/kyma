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

	corev1 "github.com/kubernetes/client-go/kubernetes/typed/core/v1"
	"github.com/kyma-project/kyma/tests/acceptance/pkg/repeat"
	"github.com/pkg/errors"
	"github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	restclient "k8s.io/client-go/rest"
)

const (
	helmBrokerURLEnvName    = "HELM_BROKER_URL"
	releaseNamespaceEnvName = "RELEASE_NAMESPACE"
)

func TestServiceCatalogContainsClusterServiceClasses(t *testing.T) {
	for testName, brokerURL := range map[string]string{
		"Helm Broker": os.Getenv(helmBrokerURLEnvName),
		// "<next broker>": os.Getenv("<broker url env name>"),
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			var brokerServices []v2.Service
			repeat.FuncAtMost(t, func() error {
				brokerServices, err := getCatalogForBroker(brokerURL)
				if err != nil {
					return errors.Wrap(err, "while getting catalog")
				}
				if len(brokerServices) < 1 {
					return fmt.Errorf("%s catalog response should contains not empty list of the services", brokerURL)
				}
				return nil
			}, 5*time.Second)

			// test is run against existing service classes (before the test) - no need to wait too much.
			awaitCatalogContainsClusterServiceClasses(t, 2*time.Second, brokerServices)
		})
	}
}

func TestServiceCatalogContainsABServiceClasses(t *testing.T) {
	// given
	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)
	aClient, err := appClient.NewForConfig(k8sConfig)
	require.NoError(t, err)
	mClient, err := mappingClient.NewForConfig(k8sConfig)
	require.NoError(t, err)
	k8sClient, err := corev1.NewForConfig(k8sConfig)
	require.NoError(t, err)
	releaseNS := os.Getenv(releaseNamespaceEnvName)
	fixApp := fixApplication()

	broker := brokerURL{
		namespace: fmt.Sprintf("test-acc-ns-broker-%s", rand.String(4)),
		prefix:    "ab-ns-for-",
	}
	var brokerServices []v2.Service

	t.Log("Creating Application")
	app, err := aClient.ApplicationconnectorV1alpha1().Applications().Create(fixApp)
	require.NoError(t, err)

	t.Logf("Creating Namespace %s", broker.namespace)
	_, err = k8sClient.Namespaces().Create(fixNamespace(broker.namespace))
	require.NoError(t, err)

	defer func() {
		if t.Failed() {
			serviceClassesReport(t, brokerServices, broker.namespace)
		}
		err = k8sClient.Namespaces().Delete(broker.namespace, &metav1.DeleteOptions{})
		assert.NoError(t, err)
		err = aClient.ApplicationconnectorV1alpha1().Applications().Delete(app.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: new(int64), // zero
		})
		assert.NoError(t, err)
	}()

	// when
	t.Log("Creating ApplicationMapping")
	_, err = mClient.ApplicationconnectorV1alpha1().ApplicationMappings(broker.namespace).Create(fixApplicationMapping())
	require.NoError(t, err)

	// then
	repeat.FuncAtMost(t, func() error {
		brokerServices, err = getCatalogForBroker(broker.buildURL(releaseNS))
		if err != nil {
			return errors.Wrap(err, "while getting catalog")
		}
		for _, svc := range brokerServices {
			if svc.ID == fixApp.Spec.Services[0].ID {
				return nil
			}
		}
		return fmt.Errorf("%s catalog response should contains service with ID: %s", broker, fixApp.Spec.Services[0].ID)
	}, time.Second*90)

	awaitCatalogContainsServiceClasses(t, broker.namespace, timeoutPerAssert, brokerServices)
}

type brokerURL struct {
	prefix    string
	namespace string
}

func (b *brokerURL) buildURL(releaseNS string) string {
	return fmt.Sprintf("http://%s%s.%s.svc.cluster.local", b.prefix, b.namespace, releaseNS)
}

func fixApplication() *appTypes.Application {
	return &appTypes.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-acc-app",
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
						"connected-app": "test-acc-app",
					},
					ProviderDisplayName: "provider",
					DisplayName:         "test-acc-app",
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

func fixApplicationMapping() *mappingTypes.ApplicationMapping {
	return &mappingTypes.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-acc-app",
		},
	}
}
