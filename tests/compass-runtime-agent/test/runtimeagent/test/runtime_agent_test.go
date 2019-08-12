package test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/runtimeagent"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestCompassRuntimeAgentSynchronization(t *testing.T) {

	oauthTokenURL := fmt.Sprintf("%s/%s/%s/%s", testSuite.GetMockServiceURL(), mock.OAuthToken, validClientId, validClientSecret)

	basicAuth := applications.NewAuth().WithBasicAuth(validUsername, validPassword)
	oauth := applications.NewAuth().WithOAuth(validClientId, validClientSecret, oauthTokenURL)

	noAuthAPIInput := applications.NewAPI("no-auth-api", "no auth api", testSuite.GetMockServiceURL())
	basicAuthAPIInput := applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).WithAuth(basicAuth)
	oauthAPIInput := applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).WithAuth(oauth)

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
				// when
				application := this.initialPhaseResult

				// updating whole application
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

				// then
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
		{
			description: "Test case 3: Change auth in all APIs",
			initialPhaseInput: func() *applications.ApplicationInput {
				return applications.NewApplication("test-app-3", "", map[string][]string{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							applications.NewAPI("no-auth-api", "no auth api", testSuite.GetMockServiceURL()),
							applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).WithAuth(basicAuth),
							applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).WithAuth(oauth),
						}).
					WithEventAPIs(
						[]*applications.EventAPIDefinitionInput{
							applications.NewEventAPI("events-api", "description"),
							applications.NewEventAPI("no-description-events-api", ""),
						},
					)
			},
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				application := this.initialPhaseResult

				var updatedAPIs []*graphql.APIDefinition

				// update no auth API to OAuth
				noAuthAPI, found := getAPIByName(application.APIs.Data, "no-auth-api")
				require.True(t, found)
				updatedInput := applications.NewAPI("basic-to-oauth", "", noAuthAPI.TargetURL).WithAuth(oauth)
				newOauthAPI, err := testSuite.CompassClient.UpdateAPI(noAuthAPI.ID, *updatedInput.ToCompassInput())
				require.NoError(t, err)
				updatedAPIs = append(updatedAPIs, newOauthAPI)

				// update OAuth API to Basic Auth
				oauthAPI, found := getAPIByName(application.APIs.Data, "oauth-auth-api")
				require.True(t, found)
				updatedInput = applications.NewAPI("oauth-to-basic", "", oauthAPI.TargetURL).WithAuth(basicAuth)
				newBasicAuthAPI, err := testSuite.CompassClient.UpdateAPI(oauthAPI.ID, *updatedInput.ToCompassInput())
				require.NoError(t, err)
				updatedAPIs = append(updatedAPIs, newBasicAuthAPI)

				// update OAuth API to Basic Auth
				basicAuthAPI, found := getAPIByName(application.APIs.Data, "basic-auth-api")
				require.True(t, found)
				updatedInput = applications.NewAPI("basic-to-no-auth", "", basicAuthAPI.TargetURL)
				newNoAuthAPI, err := testSuite.CompassClient.UpdateAPI(basicAuthAPI.ID, *updatedInput.ToCompassInput())
				require.NoError(t, err)
				updatedAPIs = append(updatedAPIs, newNoAuthAPI)

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// Due to Application Gateway caching the reverse proxy, after changing the authentication method we need to wait for cache to invalidate
					t.Log("Wait for proxy to invalidate...")
					testSuite.WaitForProxyInvalidation()

					// assert updated APIs
					testSuite.K8sResourceChecker.AssertAPIResources(t, application.ID, updatedAPIs...)
					testSuite.APIAccessChecker.AssertAPIAccess(t, updatedAPIs...)
				}
			},
		},
	}

	// Setup check if all resources were deleted
	var createdApplications []*compass.Application
	defer func() {
		waitForAgentToApplyConfig(t, testSuite)
		for _, app := range createdApplications {
			t.Logf("Asserting resources for %s application are deleted", app.ID)
			testSuite.K8sResourceChecker.AssertAppResourcesDeleted(t, app.ID)
		}
	}()

	// Setup initial phase
	for _, testCase := range testCases {
		t.Logf("Running initial phase setup for test case: %s", testCase.description)
		appInput := testCase.initialPhaseInput()

		t.Log("Creating Application...")
		response, err := testSuite.CompassClient.CreateApplication(appInput.ToCompassInput())
		require.NoError(t, err)

		defer func() {
			t.Logf("Cleaning up %s Application...", response.ID)
			removedId, err := testSuite.CompassClient.DeleteApplication(response.ID)
			require.NoError(t, err)
			assert.Equal(t, response.ID, removedId)
		}()

		createdApplications = append(createdApplications, &response)

		testCase.initialPhaseResult = response
		t.Logf("Initial test case setup finished for %s test case", testCase.description)
	}

	// Wait for agent to apply config
	waitForAgentToApplyConfig(t, testSuite)

	// Assert initial phase
	for _, testCase := range testCases {
		t.Logf("Asserting initial phase for test case: %s", testCase.description)

		t.Logf("Checking K8s resources")
		testSuite.K8sResourceChecker.AssertResourcesForApp(t, testCase.initialPhaseResult)

		t.Logf("Checking API Access")
		testSuite.APIAccessChecker.AssertAPIAccess(t, testCase.initialPhaseResult.APIs.Data...)
	}

	// Setup second phase
	for _, testCase := range testCases {
		t.Logf("Running second phase setup for test case: %s", testCase.description)
		testCase.secondPhaseSetup(t, testSuite, testCase)
		t.Logf("Initial test case setup finished for %s test case", testCase.description)
	}

	// Wait for agent to apply config
	waitForAgentToApplyConfig(t, testSuite)

	// Assert second phase
	for _, testCase := range testCases {
		t.Logf("Asserting second phase for test case: %s", testCase.description)
		testCase.secondPhaseAssert(t, testSuite, testCase)
	}
}

// TODO - test cases for deniers (after Istio resources are implemented)

func waitForAgentToApplyConfig(t *testing.T, testSuite *runtimeagent.TestSuite) {
	t.Log("Waiting for Runtime Agent to apply configuration...")
	testSuite.WaitForConfigurationApplication()
}

func getAPIByName(apis []*graphql.APIDefinition, name string) (*graphql.APIDefinition, bool) {
	for _, api := range apis {
		if api.Name == name {
			return api, true
		}
	}

	return nil, false
}
