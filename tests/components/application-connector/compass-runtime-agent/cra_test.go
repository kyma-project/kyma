package compass_runtime_agent

import (
	"context"
	"fmt"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"testing"
	"time"

	app_clientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	compass_conn_clientset "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"
	cra_compass "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestCompassRuntimeAgentFunctionalities(t *testing.T) {

	// According to https://kyma-project.io/docs/kyma/latest/05-technical-reference/00-architecture/ra-01-runtime-agent-workflow/
	// there are the things that should be tested:
	// 1. Runtime Agent fetches the certificate from the Connector to initialize connection with Compass.
	// 2. Runtime Agent stores the certificate and key for the Connector and the Director in the Secret.
	// 3. Runtime Agent synchronizes the Runtime with the Director. It does so by:
	//    - fetching new Applications from the Director and creating them in the Runtime
	//    - removing from the Runtime the Applications that no longer exist in the Director.
	// 4. Runtime Agent labels the Runtime data in the Director with the Event Gateway URL and the Console URL of the Kyma cluster. These URLs are displayed in the Compass UI.
	// 5. Runtime Agent renews the certificate for the Connector and the Director to maintain connection with Compass. This only happens if the remaining validity period for the certificate passes a certain threshold.

	k8sConfig, err := rest.InClusterConfig()
	require.Nil(t, err, "failed to create a k8s in-cluster-config")

	aclientset, err := app_clientset.NewForConfig(k8sConfig)
	require.Nil(t, err, "failed to create an application clientset")
	applications := aclientset.ApplicationconnectorV1alpha1().Applications()

	ccclientset, err := compass_conn_clientset.NewForConfig(k8sConfig)
	require.Nil(t, err, "failed to create a compass connection clientset")
	compassConnections := ccclientset.CompassConnections()

	t.Run("Compass Runtime Agent fetches the certificate from the Connector to initialize connection with Compass", func(t *testing.T) {
		// TODO: wait for the Compass Runtime Agent - Connector connection
		// TODO: check if the Compass Connection exists
		// TODO: check if the connectionState of the Compass Connection is correct
		var compassConnection *cra_compass.CompassConnection
		err := retry(func() error {
			compassConnection, err = compassConnections.Get(context.Background(), "compass-connection", metav1.GetOptions{})
			if err != nil {
				require.True(t, k8serrors.IsNotFound(err), "failed to get a compass connection: %v", err)
				return err
			}
			return nil
		}, 5, 5)
		require.Nil(t, err, "failed to get a compass connection")


		require.Equal(t, cra_compass.Synchronized, compassConnection.Status.State, "Compass Connection is not synchronized")
	})

	t.Run("Compass Runtime Agent stores the certificate and key for the Connector and the Director in the Secret", func(t *testing.T) {
		// TODO: check if the Secret exists
		// TODO: check if the certificate and the key exist in the Secret
	})

	t.Run("Compass Runtime Agent fetches new Applications from the Director and creates them in the Runtime", func(t *testing.T) {
		// TODO: check if the Applications hardcoded in Compass exist
		// TODO: check if the Applications have all the expected APIs
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

func retry(f func() error, waitSeconds, maxRetries int) error {
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