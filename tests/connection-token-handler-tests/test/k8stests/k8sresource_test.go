package k8stests

import (
	"github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/connection-token-handler-tests/test/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

const (
	waitTime = 5 * time.Second
	retries  = 5
)

func TestTokenRequests(t *testing.T) {

	appName := "test-name"
	tenant := "test-tenant"
	group := "test-group"

	client, err := testkit.NewK8sResourcesClient()

	require.NoError(t, err)

	t.Run("should create token request CR with token", func(t *testing.T) {
		//when
		tokenRequest, e := client.CreateTokenRequest(appName, group, tenant)
		require.NoError(t, e)

		//then
		checkIfTokenRequestIsProperlyCreated(t, client, tokenRequest)
	})
}
func checkIfTokenRequestIsProperlyCreated(t *testing.T, client testkit.K8sResourcesClient, tokenRequest *v1alpha1.TokenRequest) {
	isReady := &tokenRequest.Status != nil

	var e error

	for i := 0; isReady == false && i < retries; i++ {
		time.Sleep(waitTime)
		tokenRequest, e = client.GetTokenRequest(tokenRequest.Name, v1.GetOptions{})
		require.NoError(t, e)
		isReady = &tokenRequest.Status != nil
	}

	require.True(t, isReady)

	assert.NotEmpty(t, tokenRequest.Status.Token)
	assert.NotEmpty(t, tokenRequest.Context.Tenant)
	assert.NotEmpty(t, tokenRequest.Context.Group)
}
