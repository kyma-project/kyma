package testing

import (
	"fmt"
	"net/http"
	"testing"

	eventmeshclient "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	eventmeshtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/auth"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/httpclient"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/stretchr/testify/require"
)

// Test_Client_Update tests the update method for patching webhook auth.
func Test_Client_Update(t *testing.T) {
	// given
	// start mock EventMesh server.
	emMock := NewEventMeshMock()
	emMock.Start()
	defer emMock.Stop()

	// initialize EventMesh client.
	cfg := env.Config{
		BEBAPIURL:     emMock.MessagingURL,
		TokenEndpoint: emMock.TokenURL,
	}
	authenticatedClient := auth.NewAuthenticatedClient(cfg)
	httpClient, err := httpclient.NewHTTPClient(cfg.BEBAPIURL, authenticatedClient)
	require.NoError(t, err)
	emClient := eventmeshclient.NewClient(httpClient)

	// declare given objects
	givenEventMeshSub := &eventmeshtypes.Subscription{
		Name:            "Name1",
		ContentMode:     "ContentMode",
		ExemptHandshake: true,
		Qos:             eventmeshtypes.QosAtLeastOnce,
		WebhookURL:      "www.kyma-project.io",
		WebhookAuth: &eventmeshtypes.WebhookAuth{
			Type:         "abc",
			User:         "test",
			Password:     "test123",
			GrantType:    "test",
			ClientID:     "123456",
			ClientSecret: "qwerty",
			TokenURL:     "www.kyma-project.io",
		},
	}
	givenUpdateWebhook := &eventmeshtypes.WebhookAuth{
		Type:         "abc",
		User:         "test",
		Password:     "test123changed",
		GrantType:    "test",
		ClientID:     "123456changed",
		ClientSecret: "qwertychanged",
		TokenURL:     "www.changed.io",
	}

	emKeyPrefix := "/messaging/events/subscriptions"
	key := fmt.Sprintf("%s/%s", emKeyPrefix, givenEventMeshSub.Name)

	// add eventMesh subscription to mocked server.
	emMock.Subscriptions.PutSubscription(key, givenEventMeshSub)

	// when
	updateResp, err := emClient.Update(givenEventMeshSub.Name, givenUpdateWebhook)

	// then
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, updateResp.StatusCode)
	// verify webhook auth updated on server or not.
	gotSub := emMock.Subscriptions.GetSubscription(key)
	require.NotNil(t, gotSub)
	require.Equal(t, givenUpdateWebhook, gotSub.WebhookAuth)
}
