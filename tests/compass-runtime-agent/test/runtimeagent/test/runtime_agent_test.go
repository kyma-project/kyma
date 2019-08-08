package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/runtimeagent"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO - replace logrus with t.log?

// TODO - consider defining testcase object
//type testCase struct {
//	firstPhase  testPhase
//	secondPhase testPhase
//}

//func (tc testCase) ExecuteFirstPhase(t *testing.T, client *compass.Client) {
//	applicationInput := applications.NewApplication(tc.firstPhase.appName, tc.firstPhase.appDescription, tc.firstPhase.appLabels).
//		WithAPIs(tc.firstPhase.apisInputs).
//		WithEventAPIs(tc.firstPhase.eventAPIsInputs)
//
//	logrus.Info("Creating Application...")
//	response, err := client.CreateApplication(applicationInput.ToCompassInput())
//	require.NoError(t, err)
//
//	tc.firstPhase.result = response
//}

type testCase struct {
	description string

	initialPhaseInput  func() *applications.ApplicationInput
	initialPhaseResult compass.Application

	secondPhaseSetup  func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase)
	secondPhaseAssert func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase)
}

// TODO
const (
	validPassword     = "password"
	validUsername     = "username"
	validClientId     = "clientId"
	validClientSecret = "clientSecret"
)

func TestCompassRuntimeAgentSynchronization_TestCases(t *testing.T) {

	oauthTokenURL := fmt.Sprintf("%s/%s/%s/%s", testSuite.GetMockServiceURL(), mock.OAuthToken, validClientId, validClientSecret)

	noAuthAPIInput := applications.NewAPI("no-auth-api", "no auth api", testSuite.GetMockServiceURL())
	basicAuthAPIInput := applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).
		WithAuth(applications.NewAuth().WithBasicAuth(validUsername, validPassword))
	oauthAPIInput := applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).
		WithAuth(applications.NewAuth().WithOAuth(validClientId, validClientSecret, oauthTokenURL))

	// Define test cases
	testCases := []*testCase{
		{
			// TODO - event APIs
			description: "Test case 1: Create all types of APIs and remove them",
			initialPhaseInput: func() *applications.ApplicationInput {
				return applications.NewApplication("test-app-1", "testApp1", map[string][]string{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							noAuthAPIInput,
							basicAuthAPIInput,
							oauthAPIInput,
						})
			},
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// given
				application := this.initialPhaseResult

				createdNoAuthAPI := application.APIs.Data[0]
				createdBasicAuthAPI := application.APIs.Data[1]
				createdOAuthAPI := application.APIs.Data[2]

				// when
				id, err := testSuite.CompassClient.DeleteAPI(createdNoAuthAPI.ID)
				require.NoError(t, err)
				require.Equal(t, createdNoAuthAPI.ID, id)
				id, err = testSuite.CompassClient.DeleteAPI(createdBasicAuthAPI.ID)
				require.NoError(t, err)
				require.Equal(t, createdBasicAuthAPI.ID, id)
				id, err = testSuite.CompassClient.DeleteAPI(createdOAuthAPI.ID)
				require.NoError(t, err)
				require.Equal(t, createdOAuthAPI.ID, id)
			},
			secondPhaseAssert: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// then
				application := this.initialPhaseResult
				originalAPIs := application.APIs.Data

				testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, application.ID, originalAPIs[0].ID)
				testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, application.ID, originalAPIs[1].ID)
				testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, application.ID, originalAPIs[2].ID)
			},
		},
	}

	// Setup initial phase
	for _, testCase := range testCases {
		logrus.Infof("Running initial phase setup for test case: %s", testCase.description)
		appInput := testCase.initialPhaseInput()

		logrus.Info("Creating Application...")
		response, err := testSuite.CompassClient.CreateApplication(appInput.ToCompassInput())
		require.NoError(t, err)

		defer func() {
			logrus.Infof("Cleaning up %s Application...", response.ID)
			removedId, err := testSuite.CompassClient.DeleteApplication(response.ID)
			require.NoError(t, err)
			assert.Equal(t, response.ID, removedId)
		}()

		// TODO - assert with input?

		testCase.initialPhaseResult = response
		logrus.Infof("Initial test case setup finished for %s test case", testCase.description)
	}

	// Wait for agent to apply config
	waitForAgentToApplyConfig(t)

	// Assert initial phase
	for _, testCase := range testCases {
		logrus.Infof("Asserting initial phase for test case: %s", testCase.description)

		logrus.Infof("Checking K8s resources")
		testSuite.K8sResourceChecker.AssertResourcesForApp(t, testCase.initialPhaseResult)

		logrus.Infof("Checking API Access")
		testSuite.APIAccessChecker.AssertAPIAccess(t, testCase.initialPhaseResult.APIs.Data...)
		// TODO - how to do api check if expected status will be different than 200? Separate test case?
	}

	// Setup second phase
	for _, testCase := range testCases {
		logrus.Infof("Running second phase setup for test case: %s", testCase.description)
		testCase.secondPhaseSetup(t, testSuite, testCase)
		logrus.Infof("Initial test case setup finished for %s test case", testCase.description)
	}

	// Wait for agent to apply config
	waitForAgentToApplyConfig(t)

	// Assert second phase
	for _, testCase := range testCases {
		logrus.Infof("Asserting second phase for test case: %s", testCase.description)
		testCase.secondPhaseAssert(t, testSuite, testCase)
	}
}

func waitForAgentToApplyConfig(t *testing.T) {
	// TODO - consider some smarter way to wait for it
	logrus.Info("Waiting for Runtime Agent to apply configuration...")
	time.Sleep(35 * time.Second)
}
