package servicecatalog_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	reclient "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/acceptance/servicecatalog/wait"
	"github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

const (
	helmBrokerURLEnvName              = "HELM_BROKER_URL"
	remoteEnvironmentBrokerURLEnvName = "REMOTE_ENVIRONMENT_BROKER_URL"
)

func TestServiceCatalogContainsServiceClasses(t *testing.T) {
	for testName, brokerURL := range map[string]string{
		"Helm Broker": os.Getenv(helmBrokerURLEnvName),
		// "<next broker>": os.Getenv("<broker url env name>"),
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			var brokerServices []v2.Service
			wait.ForFuncAtMost(t, func() error {
				brokerServices = getCatalogForBroker(t, brokerURL)
				if len(brokerServices) < 1 {
					return fmt.Errorf("%s catalog response should contains not empty list of the services", brokerURL)
				}
				return nil
			}, 5*time.Second)

			// test is run against existing service classes (before the test) - no need to wait too much.
			awaitCatalogContainsServiceClasses(t, 2*time.Second, brokerServices)
		})
	}
}

func TestServiceCatalogContainsRemoteEnvironmentBrokerServiceClass(t *testing.T) {
	// given
	rei := remoteEnvironmentsClient(t)
	fixRE := fixRemoteEnvironment()
	brokerURL := os.Getenv(remoteEnvironmentBrokerURLEnvName)

	// when
	re, err := rei.Create(fixRE)
	require.NoError(t, err)

	var brokerServices []v2.Service
	defer func() {
		err := rei.Delete(re.Name, &v1.DeleteOptions{
			GracePeriodSeconds: new(int64), // zero
		})
		assert.NoError(t, err)

		deleteClusterServiceClassesForRemoteEnvironment(t, re)
	}()

	// then
	wait.ForFuncAtMost(t, func() error {
		brokerServices = getCatalogForBroker(t, brokerURL)
		for _, svc := range brokerServices {
			if svc.ID == fixRE.Spec.Services[0].ID {
				return nil
			}
		}
		return fmt.Errorf("%s catalog response should contains service with ID: %s", brokerURL, fixRE.Spec.Services[0].ID)
	}, 5*time.Second)
	awaitCatalogContainsServiceClasses(t, 30*time.Second, brokerServices)
}

func remoteEnvironmentsClient(t *testing.T) reclient.RemoteEnvironmentInterface {
	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)

	reClient, err := versioned.NewForConfig(k8sConfig)
	require.NoError(t, err)

	rei := reClient.ApplicationconnectorV1alpha1().RemoteEnvironments()
	return rei
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
