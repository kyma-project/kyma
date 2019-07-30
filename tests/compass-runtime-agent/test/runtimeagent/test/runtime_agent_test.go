package test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO - mock service to which you call

// TODO - consider defining testcase object
type testCase struct {
}

func TestCompassRuntimeAgentSynchronization(t *testing.T) {

	tenant := "acc-tests"

	client := compass.NewCompassClient("http://localhost:3000/graphql", tenant, "")

	// TODO - create 3 APIs of each kind
	//// One should remain the same after update
	//// One should be deleted
	//// One should be updated to something else

	noAuthAPIInput := applications.NewAPI("no-auth-1", "no auth api for app 1", "") // TODO - mock service URL

	basicAuthAPIInput := applications.NewAPI("no-auth-1", "no auth api for app 1", ""). // TODO - mock service URL
												WithAuth(applications.NewAuth().WithBasicAuth("", "")) // TODO - user pswd

	oauthAPIInput := applications.NewAPI("no-auth-1", "no auth api for app 1", ""). // TODO - mock service URL
											WithAuth(applications.NewAuth().WithOAuth("", "", "")) // TODO - clientId secret url

	application := applications.NewApplication("test-app1", "testApp1", map[string][]string{})
	application = application.WithAPIs([]*applications.APIDefinitionInput{
		noAuthAPIInput, basicAuthAPIInput, oauthAPIInput,
	})

	response, err := client.CreateApplication(application.ToCompassInput())
	require.NoError(t, err)
	defer func() {
		logrus.Infof("Cleaning up %s Application", response.Id)
		removedId, err := client.DeleteApplication(response.Id)
		require.NoError(t, err)
		assert.Equal(t, response.Id, removedId)
	}()

	assert.NotEmpty(t, response.Id)
	assert.Equal(t, 3, len(response.APIsIds))

	// Create application 1
	//// With BasicAuth, Oauth, NoAuth APIs
	//// With Events

	// TODO: consider checking CompassConnection CR to decide when to start checks
	logrus.Info("Waiting for Runtime Agent to apply configuration...")
	time.Sleep(25 * time.Second)

	// TODO - assertions

	// Verify that resources were created

	// Create application 2 and update application 1

}
