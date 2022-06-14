package compass_runtime_agent

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	applicationv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	app_clientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	cra_compass "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	compass_conn_clientset "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestCompassRuntimeAgentFunctionalities(t *testing.T) {

	// Prerequisites:
	// - Istio and "istio-system/ca-certificates" CA Cert secret
	// - Application Connector with Central Application Gateway and Application CRD installed
	// - Compass Runtime Agent installed
	// - clusterID and clusterTenant provided

	// TODO: Log explicitly that it's Compass fault if the test fails there!

	clusterID := "and-id-that-will-be-read-from-some-env-variable"
	clusterTenant := "a-tenant-that-will-be-read-from-some-env-variable"

	connectionToken, err := getConnectionToken(clusterID, clusterTenant)
	require.NoError(t, err, "failed to get runtime connection token, it's not Compass Runtime Agent's fault")

	k8sConfig, err := rest.InClusterConfig()
	require.NoError(t, err, "failed to create a k8s in-cluster-config, it's not Compass Runtime Agent's fault")

	k8sclientset, err := kubernetes.NewForConfig(k8sConfig)
	require.NoError(t, err, "failed to create a kubernetes clientset, it's not Compass Runtime Agent's fault")
	csSecrets := k8sclientset.CoreV1().Secrets("compass-system")

	_, err = csSecrets.Create(context.Background(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "compass-agent-configuration",
			Namespace: "compass-system",
		},
		StringData: map[string]string{
			"CONNECTOR_URL": connectionToken.ConnectorURL,
			"RUNTIME_ID":    clusterID,
			"TENANT":        clusterTenant,
			"TOKEN":         connectionToken.Token,
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err, "failed to create the configuration Secret, it's not Compass Runtime Agent's fault")

	ccclientset, err := compass_conn_clientset.NewForConfig(k8sConfig)
	require.NoError(t, err, "failed to create a compass connection clientset, it's not Compass Runtime Agent's fault")
	compassConnections := ccclientset.CompassConnections()

	aclientset, err := app_clientset.NewForConfig(k8sConfig)
	require.NoError(t, err, "failed to create an application clientset, it's not Compass Runtime Agent's fault")
	applications := aclientset.ApplicationconnectorV1alpha1().Applications()

	// According to https://kyma-project.io/docs/kyma/latest/05-technical-reference/00-architecture/ra-01-runtime-agent-workflow/
	// there are the things that should be tested:
	// 1. Runtime Agent fetches the certificate from the Connector to initialize connection with Compass.
	// 2. Runtime Agent stores the certificate and key for the Connector and the Director in the Secret.
	// 3. Runtime Agent synchronizes the Runtime with the Director. It does so by:
	//    - fetching new Applications from the Director and creating them in the Runtime
	//    - removing from the Runtime the Applications that no longer exist in the Director.
	// 4. Runtime Agent labels the Runtime data in the Director with the Event Gateway URL and the Console URL of the Kyma cluster. These URLs are displayed in the Compass UI.
	// 5. Runtime Agent renews the certificate for the Connector and the Director to maintain connection with Compass. This only happens if the remaining validity period for the certificate passes a certain threshold.

	t.Run("Compass Runtime Agent fetches the certificate from the Connector to initialize connection with Compass", func(t *testing.T) {
		// TODO: wait for the Compass Runtime Agent - Connector connection
		// TODO: check if the Compass Connection exists
		var compassConnection *cra_compass.CompassConnection
		err := retry(5, 5, func() error {
			compassConnection, err = compassConnections.Get(context.Background(), "compass-connection", metav1.GetOptions{})
			if err != nil {
				require.True(t, k8serrors.IsNotFound(err), "failed to get a Compass Connection: %v", err)
				return err
			}
			return nil
		})
		require.NoError(t, err, "failed to get a Compass Connection")

		// TODO: check if the connectionState of the Compass Connection is correct
		// TODO: Consider reconfiguring runtime by recreating the configuration secret when the state is ConnectionFailed. Provisioner does that
		require.Equal(t, cra_compass.Synchronized, compassConnection.Status.State, "Compass Connection is not synchronized")
	})

	t.Run("Compass Runtime Agent stores the certificate and key for the Connector and the Director in the Secret", func(t *testing.T) {
		// TODO: check if the Secret exists
		// TODO: check if the certificate and the key exist in the Secret
	})

	t.Run("Compass Runtime Agent fetches new Applications from the Director and creates them in the Runtime", func(t *testing.T) {
		var listedApplications *applicationv1alpha1.ApplicationList
		err := retry(5, 5, func() error {
			listedApplications, err = applications.List(context.Background(), metav1.ListOptions{})
			if err != nil {
				require.True(t, k8serrors.IsNotFound(err), "failed to list Applications: %v", err)
				return err
			}
			return nil
		})
		require.NoError(t, err, "failed to list Applications")

		// TODO: check if the Applications hardcoded in Compass exist
		assert.Equal(t, 3, len(listedApplications.Items))

		for _, app := range listedApplications.Items {
			// TODO: check if the Applications have all the expected APIs
		}
	})

	t.Run("Compass Runtime Agent removes from the Runtime the Applications that no longer exist in the Director", func(t *testing.T) {
		// Will it be possible without calling Director to remove an Application?
	})

	t.Run("Compass Runtime Agent labels the Runtime data in the Director with the Event Gateway URL and the Console URL of the Kyma cluster", func(t *testing.T) {
		// Will it be possible without calling Director to obtain the labeled data?
	})

	t.Run("Compass Runtime Agent renews the certificate for the Connector and the Director to maintain connection with Compass", func(t *testing.T) {
		// TODO: set the refreshCredentialsNow field in the Compass Connection
		// TODO: check if the credentials in the Secret have changed
	})
}

func getConnectionToken(id, tenant string) (graphql.OneTimeTokenForRuntimeExt, error) {
	requestOneTimeTokenForRuntimeMutation := fmt.Sprintf(`mutation { result: requestOneTimeTokenForRuntime(id: "%s"){ token connectorURL } }`, id)
	var response struct {
		Result *graphql.OneTimeTokenForRuntimeExt `json:"result"`
	}
	// TODO: executeDirectorGraphQLCall - https://github.com/kyma-project/control-plane/blob/main/components/provisioner/internal/director/directorclient.go#L214
	if err := executeDirectorGraphQLCall(requestOneTimeTokenForRuntimeMutation, tenant, &response); err != nil {
		return graphql.OneTimeTokenForRuntimeExt{}, fmt.Errorf("failed to get one-time token for Runtime %s from Director: %v", id, err)
	}
	if response.Result == nil {
		return graphql.OneTimeTokenForRuntimeExt{}, fmt.Errorf("failed to get one-time token for Runtime %s from Director: received nil response", id)
	}
	return *response.Result, nil
}

func retry(waitSeconds, maxRetries int, f func() error) error {
	tries := 0
	for {
		err := f()
		if err == nil {
			return nil
		}
		tries++
		if tries > maxRetries {
			return fmt.Errorf("all attempts failed, last error: %v", err)
		}
		<-time.After(time.Duration(waitSeconds) * time.Second)
	}
}
