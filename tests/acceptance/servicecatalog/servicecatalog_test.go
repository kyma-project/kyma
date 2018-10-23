package servicecatalog_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	corev1 "github.com/kubernetes/client-go/kubernetes/typed/core/v1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
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

func TestServiceCatalogContainsREBServiceClasses(t *testing.T) {
	// given
	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)
	reClient, err := versioned.NewForConfig(k8sConfig)
	require.NoError(t, err)
	k8sClient, err := corev1.NewForConfig(k8sConfig)
	require.NoError(t, err)
	releaseNS := os.Getenv(releaseNamespaceEnvName)
	fixRE := fixRemoteEnvironment()

	broker := brokerURL{
		namespace: fmt.Sprintf("test-acc-ns-broker-%s", rand.String(4)),
		prefix:    "reb-ns-for-",
	}
	var brokerServices []v2.Service

	t.Logf("Creating RemoteEnvironment")
	re, err := reClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Create(fixRE)
	require.NoError(t, err)

	t.Logf("Creating Namespace %s", broker.namespace)
	_, err = k8sClient.Namespaces().Create(fixNamespace(broker.namespace))
	require.NoError(t, err)

	defer func() {
		err = k8sClient.Namespaces().Delete(broker.namespace, &metav1.DeleteOptions{})
		assert.NoError(t, err)
		err = reClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Delete(re.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: new(int64), // zero
		})
		assert.NoError(t, err)
	}()

	// when
	_, err = reClient.ApplicationconnectorV1alpha1().EnvironmentMappings(broker.namespace).Create(fixEnvironmentMapping())
	require.NoError(t, err)

	// then
	repeat.FuncAtMost(t, func() error {
		brokerServices, err = getCatalogForBroker(broker.buildURL(releaseNS))
		if err != nil {
			return errors.Wrap(err, "while getting catalog")
		}
		for _, svc := range brokerServices {
			if svc.ID == fixRE.Spec.Services[0].ID {
				return nil
			}
		}
		return fmt.Errorf("%s catalog response should contains service with ID: %s", broker, fixRE.Spec.Services[0].ID)
	}, time.Second*90)
	awaitCatalogContainsServiceClasses(t, broker.namespace, time.Minute, brokerServices)
}

type brokerURL struct {
	prefix    string
	namespace string
}

func (b *brokerURL) buildURL(releaseNS string) string {
	return fmt.Sprintf("http://%s%s.%s.svc.cluster.local", b.prefix, b.namespace, releaseNS)
}

func fixRemoteEnvironment() *v1alpha1.RemoteEnvironment {
	return &v1alpha1.RemoteEnvironment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RemoteEnvironment",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-remote-env",
		},
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Description: "Remote Environment used by acceptance test",
			AccessLabel: "fix-access",
			Services: []v1alpha1.Service{
				{
					ID:   "id-00000-1234-test",
					Name: "provider-4951",
					Labels: map[string]string{
						"connected-app": "test-remote-env",
					},
					ProviderDisplayName: "provider",
					DisplayName:         "test-remote-env",
					Description:         "Remote Environment Service Class used by acceptance test",
					Tags:                []string{},
					Entries: []v1alpha1.Entry{
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

func fixEnvironmentMapping() *v1alpha1.EnvironmentMapping {
	return &v1alpha1.EnvironmentMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EnvironmentMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-remote-env",
		},
	}
}
