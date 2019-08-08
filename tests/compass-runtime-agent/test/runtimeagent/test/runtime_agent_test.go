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

type testCase struct {
	description string

	initialPhaseInput  func() *applications.ApplicationInput
	initialPhaseResult compass.Application

	secondPhaseSetup  func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase)
	secondPhaseAssert func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase)
}

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
			description: "Test case 1: Create all types of APIs and remove them",
			initialPhaseInput: func() *applications.ApplicationInput {
				return applications.NewApplication("test-app-1", "testApp1", map[string][]string{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							noAuthAPIInput,
							basicAuthAPIInput,
							oauthAPIInput,
						}).
					WithEventAPIs(
						[]*applications.EventAPIDefinitionInput{
							applications.NewEventAPI("events-api", "description"),
							applications.NewEventAPI("no-description-events-api", ""),
						},
					)
			},
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when removing all APIs individually
				application := this.initialPhaseResult
				assert.Equal(t, 3, len(application.APIs.Data))
				assert.Equal(t, 2, len(application.EventAPIs.Data))

				// remove APIs
				for _, api := range application.APIs.Data {
					id, err := testSuite.CompassClient.DeleteAPI(api.ID)
					require.NoError(t, err)
					require.Equal(t, api.ID, id)
				}

				// remove EventAPIs
				for _, eventAPI := range application.EventAPIs.Data {
					id, err := testSuite.CompassClient.DeleteEventAPI(eventAPI.ID)
					require.NoError(t, err)
					require.Equal(t, eventAPI.ID, id)
				}

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// assert APIs deleted
					for _, api := range application.APIs.Data {
						testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, application.ID, api.ID)
					}

					// assert EventAPIs deleted
					for _, eventAPI := range application.EventAPIs.Data {
						testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, application.ID, eventAPI.ID)
					}
				}
			},
		},
		{
			description: "Test case 2: Update Application overriding all APIs",
			initialPhaseInput: func() *applications.ApplicationInput {
				return applications.NewApplication("test-app-2", "", map[string][]string{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							noAuthAPIInput,
							basicAuthAPIInput,
							oauthAPIInput,
						}).
					WithEventAPIs(
						[]*applications.EventAPIDefinitionInput{
							applications.NewEventAPI("events-api", "description"),
							applications.NewEventAPI("no-description-events-api", ""),
						},
					)
			},
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when removing all APIs individually
				application := this.initialPhaseResult

				updatedInput := applications.NewApplication("test-app-2-updated", "", map[string][]string{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							noAuthAPIInput,
							basicAuthAPIInput,
							oauthAPIInput,
						}).
					WithEventAPIs(
						[]*applications.EventAPIDefinitionInput{
							applications.NewEventAPI("events-api", "description"),
						},
					)

				updatedApp, err := testSuite.CompassClient.UpdateApplication(application.ID, updatedInput.ToCompassInput())
				require.NoError(t, err)
				assert.Equal(t, 3, len(updatedApp.APIs.Data))

				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// assert previous APIs deleted
					for _, api := range application.APIs.Data {
						testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, application.ID, api.ID)
					}
					// assert previous EventAPIs deleted
					for _, eventAPI := range application.EventAPIs.Data {
						testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, application.ID, eventAPI.ID)
					}

					// assert updated Application
					testSuite.K8sResourceChecker.AssertResourcesForApp(t, updatedApp)
					testSuite.APIAccessChecker.AssertAPIAccess(t, updatedApp.APIs.Data...)
				}
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

// TODO - test cases for deniers (after implemented)

func waitForAgentToApplyConfig(t *testing.T) {
	// TODO - consider some smarter way to wait for it
	logrus.Info("Waiting for Runtime Agent to apply configuration...")
	time.Sleep(35 * time.Second)
}
