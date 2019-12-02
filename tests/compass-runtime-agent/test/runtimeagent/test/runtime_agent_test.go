package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/runtimeagent"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	description string

	initialPhaseInput  func() *applications.ApplicationRegisterInput
	initialPhaseAssert func(t *testing.T, testSuite *runtimeagent.TestSuite, application compass.Application)
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
	var emptySpec graphql.CLOB = ""
	var apiSpecData graphql.CLOB = "defaultContent"

	oauthTokenURL := fmt.Sprintf("%s/%s/%s/%s", testSuite.GetMockServiceURL(), mock.OAuthToken, validClientId, validClientSecret)

	basicAuth := applications.NewAuth().WithBasicAuth(validUsername, validPassword)
	oauth := applications.NewAuth().WithOAuth(validClientId, validClientSecret, oauthTokenURL)

	noAuthAPIInput := applications.NewAPI("no-auth-api", "no auth api", testSuite.GetMockServiceURL()).WithJsonApiSpec(&apiSpecData)
	basicAuthAPIInput := applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).WithAuth(basicAuth).WithYamlApiSpec(&apiSpecData)
	oauthAPIInput := applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).WithAuth(oauth).WithXMLApiSpec(&emptySpec)

	// Define test cases
	testCases := []*testCase{
		{
			description: "Test case 1: Create all types of APIs and remove them",
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-1", "testApp1", map[string]interface{}{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							noAuthAPIInput,
							basicAuthAPIInput,
							oauthAPIInput,
						}).
					WithEventAPIs(
						[]*applications.EventAPIDefinitionInput{
							applications.NewEventAPI("events-api", "description").WithJsonEventApiSpec(&apiSpecData),
							applications.NewEventAPI("events-api-with-empty-string-spec", "description").WithYamlEventApiSpec(&emptySpec),
							applications.NewEventAPI("no-description-events-api", "").WithYamlEventApiSpec(&apiSpecData),
						},
					)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when removing all APIs individually
				application := this.initialPhaseResult
				assert.Equal(t, 3, len(application.APIDefinitions.Data))
				assert.Equal(t, 3, len(application.EventDefinitions.Data))

				// remove APIs
				for _, api := range application.APIDefinitions.Data {
					id, err := testSuite.CompassClient.DeleteAPI(api.ID)
					require.NoError(t, err)
					require.Equal(t, api.ID, id)
				}

				// remove EventAPIs
				for _, eventAPI := range application.EventDefinitions.Data {
					id, err := testSuite.CompassClient.DeleteEventAPI(eventAPI.ID)
					require.NoError(t, err)
					require.Equal(t, eventAPI.ID, id)
				}

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// assert APIs deleted
					for _, api := range application.APIDefinitions.Data {
						testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, application.Name, api.ID)
					}

					// assert EventAPIs deleted
					for _, eventAPI := range application.EventDefinitions.Data {
						testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, application.Name, eventAPI.ID)
					}
				}
			},
		},
		{
			description: "Test case 2: Update Application overriding all APIs",
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-2", "", map[string]interface{}{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							noAuthAPIInput,
							basicAuthAPIInput,
							oauthAPIInput,
						}).
					WithEventAPIs(
						[]*applications.EventAPIDefinitionInput{
							applications.NewEventAPI("events-api", "description").WithJsonEventApiSpec(&apiSpecData),
							applications.NewEventAPI("no-description-events-api", "").WithJsonEventApiSpec(&emptySpec),
						},
					)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				application := this.initialPhaseResult

				// remove existing APIs
				for _, api := range application.APIDefinitions.Data {
					id, err := testSuite.CompassClient.DeleteAPI(api.ID)
					require.NoError(t, err)
					require.Equal(t, api.ID, id)
				}

				// remove existing EventAPIs
				for _, eventAPI := range application.EventDefinitions.Data {
					id, err := testSuite.CompassClient.DeleteEventAPI(eventAPI.ID)
					require.NoError(t, err)
					require.Equal(t, eventAPI.ID, id)
				}

				// create new APIs
				apiInputs := []*applications.APIDefinitionInput{
					noAuthAPIInput,
					basicAuthAPIInput,
					oauthAPIInput,
				}
				for _, v := range apiInputs {
					_, err := testSuite.CompassClient.CreateAPI(application.ID, *v.ToCompassInput())
					require.NoError(t, err)
				}

				// create new EventAPIs
				eventAPIInputs := []*applications.EventAPIDefinitionInput{
					applications.NewEventAPI("events-api", "description").WithJsonEventApiSpec(&apiSpecData),
				}
				for _, v := range eventAPIInputs {
					_, err := testSuite.CompassClient.CreateEventAPI(application.ID, *v.ToCompassInput())
					require.NoError(t, err)
				}

				// updating whole application
				updatedInput := applications.NewApplicationUpdateInput("test-app-2-updated", "")

				updatedApp, err := testSuite.CompassClient.UpdateApplication(application.ID, updatedInput.ToCompassInput())
				require.NoError(t, err)
				assert.Equal(t, 3, len(updatedApp.APIDefinitions.Data))
				assert.Equal(t, 1, len(updatedApp.EventDefinitions.Data))

				apiIds := getAPIsIds(updatedApp)
				t.Logf("Updated APIs for %s Application", updatedApp.Name)
				logIds(t, apiIds)
				t.Logf("Adding denier labels for %s Application", updatedApp.Name)
				testSuite.AddDenierLabels(t, updatedApp.Name, apiIds...)

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// assert previous APIs deleted
					for _, api := range application.APIDefinitions.Data {
						testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, updatedApp.Name, api.ID)
					}
					// assert previous EventAPIs deleted
					for _, eventAPI := range application.EventDefinitions.Data {
						testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, updatedApp.Name, eventAPI.ID)
					}

					// assert updated Application
					testSuite.K8sResourceChecker.AssertResourcesForApp(t, updatedApp)
					testSuite.APIAccessChecker.AssertAPIAccess(t, updatedApp.Name, updatedApp.APIDefinitions.Data...)
				}
			},
		},
		{
			description: "Test case 3: Change auth in all APIs",
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-3", "", map[string]interface{}{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							applications.NewAPI("no-auth-api", "no auth api", testSuite.GetMockServiceURL()).WithJsonApiSpec(&emptySpec),
							applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).WithAuth(basicAuth).WithXMLApiSpec(&emptySpec),
							applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).WithAuth(oauth).WithYamlApiSpec(&emptySpec),
						}).
					WithEventAPIs(
						[]*applications.EventAPIDefinitionInput{
							applications.NewEventAPI("events-api", "description").WithJsonEventApiSpec(&emptySpec),
							applications.NewEventAPI("no-description-events-api", "").WithJsonEventApiSpec(&emptySpec),
						},
					)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				application := this.initialPhaseResult

				var updatedAPIs []*graphql.APIDefinition

				// update no auth API to OAuth
				noAuthAPI, found := getAPIByName(application.APIDefinitions.Data, "no-auth-api")
				require.True(t, found)
				updatedInput := applications.NewAPI("no-auth-to-oauth", "", noAuthAPI.TargetURL).WithAuth(oauth).WithJsonApiSpec(&emptySpec)
				newOauthAPI, err := testSuite.CompassClient.UpdateAPI(noAuthAPI.ID, *updatedInput.ToCompassInput())
				require.NoError(t, err)
				updatedAPIs = append(updatedAPIs, newOauthAPI)

				// update OAuth API to Basic Auth
				oauthAPI, found := getAPIByName(application.APIDefinitions.Data, "oauth-auth-api")
				require.True(t, found)
				updatedInput = applications.NewAPI("oauth-to-basic", "", oauthAPI.TargetURL).WithAuth(basicAuth).WithJsonApiSpec(&emptySpec)
				newBasicAuthAPI, err := testSuite.CompassClient.UpdateAPI(oauthAPI.ID, *updatedInput.ToCompassInput())
				require.NoError(t, err)
				updatedAPIs = append(updatedAPIs, newBasicAuthAPI)

				// update Basic Auth API to no auth
				basicAuthAPI, found := getAPIByName(application.APIDefinitions.Data, "basic-auth-api")
				require.True(t, found)
				updatedInput = applications.NewAPI("basic-to-no-auth", "", basicAuthAPI.TargetURL).WithJsonApiSpec(&emptySpec)
				newNoAuthAPI, err := testSuite.CompassClient.UpdateAPI(basicAuthAPI.ID, *updatedInput.ToCompassInput())
				require.NoError(t, err)
				updatedAPIs = append(updatedAPIs, newNoAuthAPI)

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// assert updated APIs
					testSuite.K8sResourceChecker.AssertAPIResources(t, application.Name, updatedAPIs...)

					testSuite.APIAccessChecker.AssertAPIAccess(t, application.Name, updatedAPIs...)
				}
			},
		},
		// TODO: CSRF Tokens does not work properly with Director (https://github.com/kyma-incubator/compass/issues/207)
		// TODO: Issue is closed
		//{
		//	description: "Test case 4:Fetch new CSRF token and retry if token expired",
		//	initialPhaseInput: func() *applications.ApplicationRegisterInput {
		//		csrfAuth := applications.NewAuth().WithBasicAuth(validUsername, validPassword).
		//			WithCSRF(testSuite.GetMockServiceURL() + mock.CSRFToken.String() + "/valid-csrf-token")
		//		csrfAPIInput := applications.NewAPI("csrf-api", "csrf", testSuite.GetMockServiceURL()).WithAuth(csrfAuth)
		//
		//		app := applications.NewApplication("test-app-4", "testApp4", map[string][]string{}).
		//			WithAPIs([]*applications.APIDefinitionInput{csrfAPIInput})
		//
		//		return app
		//	},
		//	initialPhaseAssert: func(t *testing.T, testSuite *runtimeagent.TestSuite, application compass.Application) {
		//		testSuite.K8sResourceChecker.AssertResourcesForApp(t, application)
		//
		//		// call CSRF API
		//		csrfAPI, found := getAPIByName(application.APIs.Data, "csrf-api")
		//		require.True(t, found)
		//		response := testSuite.APIAccessChecker.CallAccessService(t, application.ID, csrfAPI.ID, mock.CSERTarget.String()+"/valid-csrf-token")
		//		util.RequireStatus(t, http.StatusOK, response)
		//	},
		//	secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
		//		csrfAPI, found := getAPIByName(this.initialPhaseResult.APIs.Data, "csrf-api")
		//		require.True(t, found)
		//
		//		// when updating CSRF API (new token is expected, old token is cached)
		//		modifiedCSRFAuth := applications.NewAuth().WithCSRF(testSuite.GetMockServiceURL() + mock.CSRFToken.String() + "/new-csrf-token")
		//		modifiedCSRFAPIInput := applications.NewAPI("csrf-api", "csrf", testSuite.GetMockServiceURL()).WithAuth(modifiedCSRFAuth)
		//		modifiedCSRFAPI, err := testSuite.CompassClient.UpdateAPI(csrfAPI.ID, *modifiedCSRFAPIInput.ToCompassInput())
		//		require.NoError(t, err)
		//
		//		// then
		//		this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
		//			// assert Token URL updated
		//			testSuite.K8sResourceChecker.AssertAPIResources(t, this.initialPhaseResult.ID, modifiedCSRFAPI)
		//
		//			// assert call is made with new token
		//			response := testSuite.APIAccessChecker.CallAccessService(t, this.initialPhaseResult.ID, csrfAPI.ID, mock.CSERTarget.String()+"/new-csrf-token")
		//			util.RequireStatus(t, http.StatusOK, response)
		//		}
		//	},
		//},
		{
			description: "Test case 5: Denier should block access without labels",
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-5", "", map[string]interface{}{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							applications.NewAPI("no-auth-api", "no auth api", testSuite.GetMockServiceURL()).WithJsonApiSpec(&emptySpec),
							applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).WithAuth(basicAuth).WithJsonApiSpec(&emptySpec),
							applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).WithAuth(oauth).WithJsonApiSpec(&emptySpec),
						})
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				application := this.initialPhaseResult

				apiIds := getAPIsIds(application)
				t.Logf("Removing denier labels for %s Application, For APIs: ", application.Name)
				logIds(t, apiIds)
				testSuite.RemoveDenierLabels(t, application.Name, apiIds...)

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// assert deniers block requests and the response has status 403
					for _, api := range application.APIDefinitions.Data {
						path := testSuite.APIAccessChecker.GetPathBasedOnAuth(t, api.DefaultAuth)
						response := testSuite.APIAccessChecker.CallAccessService(t, application.Name, api.ID, path)
						util.RequireStatus(t, http.StatusForbidden, response)
					}
				}
			},
		},
	}

	// Setup check if all resources were deleted
	var createdApplications []*compass.Application
	defer func() {
		t.Log("Checking resources are removed")
		waitForAgentToApplyConfig(t, testSuite)
		for _, app := range createdApplications {
			t.Logf("Asserting resources for %s application are deleted", app.Name)
			testSuite.K8sResourceChecker.AssertAppResourcesDeleted(t, app.Name)
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
			t.Logf("Cleaning up %s Application...", response.Name)
			removedId, err := testSuite.CompassClient.DeleteApplication(response.ID)
			require.NoError(t, err)
			assert.Equal(t, response.ID, removedId)
		}()

		apiIds := getAPIsIds(response)
		t.Logf("APIs for %s Application", response.Name)
		logIds(t, apiIds)

		t.Logf("Adding denier labels for %s Application...", response.Name)
		testSuite.AddDenierLabels(t, response.Name, apiIds...)

		createdApplications = append(createdApplications, &response)

		testCase.initialPhaseResult = response
		t.Logf("Initial test case setup finished for %s test case. Created Application: %s", testCase.description, response.Name)
	}

	// Wait for agent to apply config
	waitForAgentToApplyConfig(t, testSuite)

	// Assert initial phase
	for _, testCase := range testCases {
		t.Logf("Asserting initial phase for test case: %s", testCase.description)
		testCase.initialPhaseAssert(t, testSuite, testCase.initialPhaseResult)
	}

	// Setup second phase
	for _, testCase := range testCases {
		t.Logf("Running second phase setup for test case: %s", testCase.description)
		testCase.secondPhaseSetup(t, testSuite, testCase)
		t.Logf("Second test case setup finished for %s test case", testCase.description)
	}

	// Wait for agent to apply config
	waitForAgentToApplyConfig(t, testSuite)

	// Due to Application Gateway caching the reverse proxy, after changing the authentication method we need to wait for cache to invalidate
	t.Log("Wait for proxy to invalidate...")
	testSuite.WaitForProxyInvalidation()

	// Assert second phase
	for _, testCase := range testCases {
		t.Logf("Asserting second phase for test case: %s", testCase.description)
		testCase.secondPhaseAssert(t, testSuite, testCase)
	}
}

func assertK8sResourcesAndAPIAccess(t *testing.T, testSuite *runtimeagent.TestSuite, application compass.Application) {
	t.Logf("Waiting for %s Application to be deployed...", application.Name)
	testSuite.WaitForApplicationToBeDeployed(t, application.Name)

	t.Logf("Checking K8s resources")
	testSuite.K8sResourceChecker.AssertResourcesForApp(t, application)

	t.Logf("Checking API Access")
	testSuite.APIAccessChecker.AssertAPIAccess(t, application.Name, application.APIDefinitions.Data...)
}

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

func logIds(t *testing.T, ids []string) {
	for _, id := range ids {
		t.Log(id)
	}
}

func getAPIsIds(application compass.Application) []string {
	ids := make([]string, len(application.APIDefinitions.Data))
	for i, api := range application.APIDefinitions.Data {
		ids[i] = api.ID
	}

	return ids
}
