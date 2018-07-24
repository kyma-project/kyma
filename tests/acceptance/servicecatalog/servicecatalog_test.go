package servicecatalog_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	reclient "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/typed/remoteenvironment/v1alpha1"
	"github.com/pmorie/go-open-service-broker-client/v2"
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
			brokerServices := getServiceBrokerServices(t, brokerURL)

			// then
			assert.NotEmpty(t, brokerServices)
			// test is run against existing service classes (before the test) - no need to wait too much.
			awaitCatalogContainsServiceClasses(t, 2*time.Second, brokerServices)
		})
	}
}

func TestServiceCatalogContainsRemoteEnvironmentBrokerServiceClass(t *testing.T) {
	// given
	rei := remoteEnvironmentsClient(t)

	// when
	re, err := rei.Create(fixRemoteEnvironment())
	require.NoError(t, err)

	var brokerServices []v2.Service
	defer func() {
		err := rei.Delete(re.Name, &v1.DeleteOptions{
			GracePeriodSeconds: new(int64), // zero
		})
		assert.NoError(t, err)

		deleteClusterServiceClassesForRemoteEnvironment(t, re)
	}()
	brokerServices = getServiceBrokerServices(t, os.Getenv(remoteEnvironmentBrokerURLEnvName))

	// then
	assert.NotEmpty(t, brokerServices)
	// timeout must be greater than the broker relistDuration
	awaitCatalogContainsServiceClasses(t, 12*time.Second, brokerServices)
}

func remoteEnvironmentsClient(t *testing.T) reclient.RemoteEnvironmentInterface {
	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)

	reClient, err := versioned.NewForConfig(k8sConfig)
	require.NoError(t, err)

	rei := reClient.RemoteenvironmentV1alpha1().RemoteEnvironments("default")
	return rei
}

func fixRemoteEnvironment() *v1alpha1.RemoteEnvironment {
	return &v1alpha1.RemoteEnvironment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RemoteEnvironment",
			APIVersion: "remoteenvironment.kyma.cx/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-remote-env",
		},
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Source: v1alpha1.Source{
				Namespace:   "com.ns",
				Type:        "commerce",
				Environment: "production",
			},
			Description: "Remote Environment used by acceptance test",
			Services: []v1alpha1.Service{
				{
					ID:                  "id-00000-1234",
					ProviderDisplayName: "provider",
					DisplayName:         "test-remote-env",
					LongDescription:     "Remote Environment Service Class used by acceptance test",
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
