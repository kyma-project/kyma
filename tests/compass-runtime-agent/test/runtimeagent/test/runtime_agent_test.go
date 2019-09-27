package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/runtimeagent"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testCase struct {
	description string

	initialPhaseInput  func() *applications.ApplicationInput
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
	basicAuthAPIInput := applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).WithAuth(basicAuth)
	oauthAPIInput := applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).WithAuth(oauth).WithJsonApiSpec(&emptySpec)

	// Define test cases
	testCases := []*testCase{
		{
			description: "Test case 1: Create all types of APIs and remove them",
			initialPhaseInput: func() *applications.ApplicationInput {
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
							applications.NewEventAPI("events-api-with-empty-string-spec", "description").WithJsonEventApiSpec(&emptySpec),
							applications.NewEventAPI("no-description-events-api", ""),
						},
					)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when removing all APIs individually
				application := this.initialPhaseResult
				assert.Equal(t, 3, len(application.APIs.Data))
				assert.Equal(t, 3, len(application.EventAPIs.Data))

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
							applications.NewEventAPI("no-description-events-api", ""),
						},
					)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				application := this.initialPhaseResult

				// updating whole application
				updatedInput := applications.NewApplication("test-app-2-updated", "", map[string]interface{}{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							noAuthAPIInput,
							basicAuthAPIInput,
							oauthAPIInput,
						}).
					WithEventAPIs(
						[]*applications.EventAPIDefinitionInput{
							applications.NewEventAPI("events-api", "description").WithJsonEventApiSpec(&apiSpecData),
						},
					)

				updatedApp, err := testSuite.CompassClient.UpdateApplication(application.ID, updatedInput.ToCompassInput())
				require.NoError(t, err)
				assert.Equal(t, 3, len(updatedApp.APIs.Data))

				apiIds := getAPIsIds(updatedApp)
				t.Logf("Updated APIs for %s Application", updatedApp.ID)
				logIds(t, apiIds)
				t.Logf("Adding denier labels for %s Application", updatedApp.ID)
				testSuite.AddDenierLabels(t, updatedApp.ID, apiIds...)

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
				return applications.NewApplication("test-app-3", "", map[string]interface{}{}).
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
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				application := this.initialPhaseResult

				var updatedAPIs []*graphql.APIDefinition

				// update no auth API to OAuth
				noAuthAPI, found := getAPIByName(application.APIs.Data, "no-auth-api")
				require.True(t, found)
				updatedInput := applications.NewAPI("no-auth-to-oauth", "", noAuthAPI.TargetURL).WithAuth(oauth)
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

				// update Basic Auth API to no auth
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
		// TODO: CSRF Tokens does not work properly with Director (https://github.com/kyma-incubator/compass/issues/207)
		// TODO: Issue is closed
		//{
		//	description: "Test case 4:Fetch new CSRF token and retry if token expired",
		//	initialPhaseInput: func() *applications.ApplicationInput {
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
			initialPhaseInput: func() *applications.ApplicationInput {
				return applications.NewApplication("test-app-5", "", map[string]interface{}{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							applications.NewAPI("no-auth-api", "no auth api", testSuite.GetMockServiceURL()),
							applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).WithAuth(basicAuth),
							applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).WithAuth(oauth),
						})
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				application := this.initialPhaseResult

				apiIds := getAPIsIds(application)
				t.Logf("Removing denier labels for %s Application, For APIs: ", application.ID)
				logIds(t, apiIds)
				testSuite.RemoveDenierLabels(t, application.ID, apiIds...)

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// assert deniers block requests and the response has status 403
					for _, api := range application.APIs.Data {
						path := testSuite.APIAccessChecker.GetPathBasedOnAuth(t, api.DefaultAuth)
						response := testSuite.APIAccessChecker.CallAccessService(t, application.ID, api.ID, path)
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

		apiIds := getAPIsIds(response)
		t.Logf("APIs for %s Application", response.ID)
		logIds(t, apiIds)

		t.Logf("Adding denier labels for %s Application...", response.ID)
		testSuite.AddDenierLabels(t, response.ID, apiIds...)

		createdApplications = append(createdApplications, &response)

		testCase.initialPhaseResult = response
		t.Logf("Initial test case setup finished for %s test case. Created Application: %s", testCase.description, response.ID)
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

	// Assert second phase
	for _, testCase := range testCases {
		t.Logf("Asserting second phase for test case: %s", testCase.description)
		testCase.secondPhaseAssert(t, testSuite, testCase)
	}
}

func TestCompassRuntimeAgentNotManagedApplications(t *testing.T) {
	t.Run("should not delete an Application if it has no CompassMetadata in Spec", func(t *testing.T) {
		// given
		compassManagedApplicationName := "compass-managed-app"
		compassManagedApplicationTemplate := createSimpleApplicationTemplate(compassManagedApplicationName)

		t.Logf("Creating Application %s managed by Compass Runtime Agent", compassManagedApplicationName)
		compassManagedApplication, err := testSuite.AppClient.Create(&compassManagedApplicationTemplate)
		require.NoError(t, err)

		compassNotManagedApplicationName := "compass-not-managed-app"
		compassNotManagedApplicationTemplate := createSimpleApplicationTemplate(compassNotManagedApplicationName)
		compassNotManagedApplicationTemplate.Spec.CompassMetadata = nil

		t.Logf("Creating Application %s not managed by Compass Runtime Agent", compassNotManagedApplicationName)
		compassNotManagedApplication, err := testSuite.AppClient.Create(&compassNotManagedApplicationTemplate)
		require.NoError(t, err)
		defer func() {
			t.Logf("Deleting Application %s not managed by Compass Runtime Agent", compassNotManagedApplicationName)
			err := testSuite.AppClient.Delete(compassNotManagedApplication.Name, &metav1.DeleteOptions{})
			require.NoError(t, err)

			testSuite.K8sResourceChecker.AssertAppResourcesDeleted(t, compassNotManagedApplication.Name)
		}()

		// when
		waitForAgentToApplyConfig(t, testSuite)

		// then
		t.Logf("Asserting that Application managed by Compass Runtime Agent is deleted if does not exist in Director config")
		testSuite.K8sResourceChecker.AssertAppResourcesDeleted(t, compassManagedApplication.Name)

		t.Logf("Asserting that Application not managed by Compass Runtime Agent still exists even if does not exist in Director config")
		returnedCompassNotManagedApplication, err := testSuite.AppClient.Get(compassNotManagedApplication.Name, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, &compassNotManagedApplication, returnedCompassNotManagedApplication)
	})
}

func createSimpleApplicationTemplate(id string) v1alpha1.Application {
	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
		Spec: v1alpha1.ApplicationSpec{
			Description:     "Description",
			Services:        []v1alpha1.Service{},
			CompassMetadata: &v1alpha1.CompassMetadata{Authentication: v1alpha1.Authentication{ClientIds: []string{id}}},
		},
	}
}

func assertK8sResourcesAndAPIAccess(t *testing.T, testSuite *runtimeagent.TestSuite, application compass.Application) {
	t.Logf("Checking K8s resources")
	testSuite.K8sResourceChecker.AssertResourcesForApp(t, application)

	t.Logf("Checking API Access")
	testSuite.APIAccessChecker.AssertAPIAccess(t, application.APIs.Data...)
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
	ids := make([]string, len(application.APIs.Data))
	for i, api := range application.APIs.Data {
		ids[i] = api.ID
	}

	return ids
}
