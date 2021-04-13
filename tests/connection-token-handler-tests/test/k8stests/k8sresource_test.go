package k8stests

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/connection-token-handler-tests/test/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	waitTime = 5 * time.Second
	retries  = 5
)

func TestTokenRequests(t *testing.T) {
	appName := "test-name"

	client, err := testkit.NewK8sResourcesClient()

	require.NoError(t, err)

	t.Run("should create token request CR with token", func(t *testing.T) {
		//when
		tokenRequest, e := client.CreateTokenRequest(addSuffix(appName))
		require.NoError(t, e)

		//then
		tokenRequest = waitForToken(t, client, tokenRequest)

		assert.NotEmpty(t, tokenRequest.Status.Token)
	})
}

func waitForToken(t *testing.T, client testkit.K8sResourcesClient, tokenRequest *v1alpha1.TokenRequest) *v1alpha1.TokenRequest {
	isReady := isTokenReady(tokenRequest)

	var e error

	for i := 0; !isReady && i < retries; i++ {
		time.Sleep(waitTime)
		tokenRequest, e = client.GetTokenRequest(tokenRequest.Name, v1.GetOptions{})
		require.NoError(t, e)
		isReady = isTokenReady(tokenRequest)
	}

	require.True(t, isReady)

	return tokenRequest
}

func isTokenReady(tokenRequest *v1alpha1.TokenRequest) bool {
	return &tokenRequest.Status != nil && tokenRequest.Status.State == "OK"
}

func addSuffix(appName string) string {
	return appName + "-" + rand.String(5)
}
