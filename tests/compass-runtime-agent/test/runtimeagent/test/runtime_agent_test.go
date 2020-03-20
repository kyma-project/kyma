package test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/runtimeagent"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	description string
	log         *testkit.Logger

	initialPhaseInput  func() *applications.ApplicationRegisterInput
	initialPhaseAssert func(t *testing.T, log *testkit.Logger, testSuite *runtimeagent.TestSuite, initialPhaseResult TestApplicationData)
	initialPhaseResult TestApplicationData

	secondPhaseSetup   func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase)
	secondPhaseCleanup CleanupFunc
	secondPhaseAssert  func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase)
}

type CleanupFunc func(t *testing.T)

func (cf CleanupFunc) Append(cleanupFunc CleanupFunc) CleanupFunc {
	return func(t *testing.T) {
		cf(t)
		cleanupFunc(t)
	}
}

type TestApplicationData struct {
	Application compass.Application
	Certificate tls.Certificate
}

const (
	validPassword     = "password"
	validUsername     = "username"
	validClientId     = "clientId"
	validClientSecret = "clientSecret"

	noAuthAPIPackageName    = "no-auth-api-pkg"
	basicAuthAPIPackageName = "basic-auth-api-pkg"
	oAuthAPIPackageName     = "oauth-api-pkg"
)

func TestCompassRuntimeAgentSynchronization(t *testing.T) {
	var emptySpec graphql.CLOB = ""
	var apiSpecData graphql.CLOB = "defaultContent"

	oauthTokenURL := fmt.Sprintf("%s/%s/%s/%s", testSuite.GetMockServiceURL(), mock.OAuthToken, validClientId, validClientSecret)
	basicAuth := applications.NewAuth().WithBasicAuth(validUsername, validPassword)
	oauth := applications.NewAuth().WithOAuth(validClientId, validClientSecret, oauthTokenURL)

	jsonAPIInput := applications.NewAPI("some json api", "json api", testSuite.GetMockServiceURL()).WithJsonApiSpec(&apiSpecData)
	yamlAPIInput := applications.NewAPI("some yaml api", "yaml api", testSuite.GetMockServiceURL()).WithYamlApiSpec(&apiSpecData)
	xmlAPIInput := applications.NewAPI("some xml api", "xml api", testSuite.GetMockServiceURL()).WithXMLApiSpec(&emptySpec)

	jsonEventAPIInput := applications.NewEventDefinition("json event API", "some json event API").WithJsonEventSpec(&apiSpecData)
	yamlEventAPIInput := applications.NewEventDefinition("yaml event API", "some ymal event API").WithYamlEventSpec(&apiSpecData)
	emptySpecEventAPIInput := applications.NewEventDefinition("empty  event API", "some empty event API").WithJsonEventSpec(&emptySpec)

	noAuthAPIPkgInput := applications.NewAPIPackage(noAuthAPIPackageName, "no auth pkg description").
		WithAPIDefinitions(
			[]*applications.APIDefinitionInput{
				jsonAPIInput, yamlAPIInput, xmlAPIInput,
			}).
		WithEventDefinitions(
			[]*applications.EventDefinitionInput{
				jsonEventAPIInput, yamlEventAPIInput, emptySpecEventAPIInput,
			})

	basicAuthAPIPkgInput := applications.NewAPIPackage(basicAuthAPIPackageName, "basic auth pkg description").
		WithAPIDefinitions(
			[]*applications.APIDefinitionInput{
				jsonAPIInput, yamlAPIInput, xmlAPIInput,
			}).
		WithEventDefinitions(
			[]*applications.EventDefinitionInput{
				jsonEventAPIInput, yamlEventAPIInput, emptySpecEventAPIInput,
			}).
		WithAuth(basicAuth)

	oauthAPIPkgInput := applications.NewAPIPackage(oAuthAPIPackageName, "oauth pkg description").
		WithAPIDefinitions(
			[]*applications.APIDefinitionInput{
				jsonAPIInput, yamlAPIInput, xmlAPIInput,
			}).
		WithEventDefinitions(
			[]*applications.EventDefinitionInput{
				jsonEventAPIInput, yamlEventAPIInput, emptySpecEventAPIInput,
			}).
		WithAuth(oauth)

	// Define test cases
	testCases := []*testCase{
		{
			description: "Test case 1: Create all types of API packages and remove them",
			log:         testkit.NewLogger(t, map[string]string{"ApplicationName": "test-app-1"}),
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-1", "provider 1", "testApp1", map[string]interface{}{}).
					WithAPIPackages(noAuthAPIPkgInput, basicAuthAPIPkgInput, oauthAPIPkgInput)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when removing all APIs individually
				application := this.initialPhaseResult.Application
				apiPackages := application.Packages.Data
				// TODO: assert packages?
				require.Equal(t, 3, len(apiPackages))

				// remove Packages
				for _, pkg := range application.Packages.Data {
					id, err := testSuite.CompassClient.DeleteAPIPackage(pkg.ID)
					require.NoError(t, err)
					require.Equal(t, pkg.ID, id)
				}

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					k8sChecker := testSuite.K8sResourceChecker.NewChecker(t, this.log)

					// assert APIs deleted
					for _, pkg := range apiPackages {
						k8sChecker.AssertAPIPackageDeleted(pkg, application.Name)
					}
				}
			},
		},
		//{
		//	description: "Test case 2: Update Application overriding all APIs",
		//	initialPhaseInput: func() *applications.ApplicationRegisterInput {
		//		return applications.NewApplication("test-app-2", "provider 2", "", map[string]interface{}{}).
		//			WithAPIPackages(noAuthAPIPkgInput, oauthAPIPkgInput, basicAuthAPIPkgInput)
		//	},
		//	initialPhaseAssert: assertK8sResourcesAndAPIAccess,
		//	secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
		//		// when
		//		application := this.initialPhaseResult.Application
		//		require.Equal(t, 1, len(application.Packages.Data))
		//
		//		apiPackage := application.Packages.Data[0]
		//		assert.Equal(t, "api-package-2", apiPackage.Name)
		//
		//		apiDefinitions := apiPackage.APIDefinitions.Data
		//		eventAPIDefinitions := apiPackage.EventDefinitions.Data
		//
		//		assert.Equal(t, 3, len(apiDefinitions))
		//		assert.Equal(t, 2, len(eventAPIDefinitions))
		//
		//		// remove existing APIs
		//		for _, api := range apiDefinitions {
		//			id, err := testSuite.CompassClient.DeleteAPI(api.ID)
		//			require.NoError(t, err)
		//			require.Equal(t, api.ID, id)
		//		}
		//
		//		// remove existing EventAPIs
		//		for _, eventAPI := range eventAPIDefinitions {
		//			id, err := testSuite.CompassClient.DeleteEventAPI(eventAPI.ID)
		//			require.NoError(t, err)
		//			require.Equal(t, eventAPI.ID, id)
		//		}
		//
		//		// create new APIs
		//		apiInputs := []*applications.APIDefinitionInput{
		//			noAuthAPIInput,
		//			basicAuthAPIInput,
		//			oauthAPIInput,
		//		}
		//		for _, v := range apiInputs {
		//			_, err := testSuite.CompassClient.CreateAPI(application.ID, *v.ToCompassInput())
		//			require.NoError(t, err)
		//		}
		//
		//		// create new EventAPIs
		//		eventAPIInputs := []*applications.EventDefinitionInput{
		//			applications.NewEventDefinition("events-api", "description").WithJsonEventSpec(&apiSpecData),
		//		}
		//		for _, v := range eventAPIInputs {
		//			_, err := testSuite.CompassClient.CreateEventAPI(application.ID, *v.ToCompassInput())
		//			require.NoError(t, err)
		//		}
		//
		//		// updating whole application
		//		updatedInput := applications.NewApplicationUpdateInput("test-app-2-updated", "update-provider", "")
		//
		//		updatedApp, err := testSuite.CompassClient.UpdateApplication(application.ID, updatedInput.ToCompassInput())
		//		require.NoError(t, err)
		//
		//		updatedAPIPackage := updatedApp.Packages.Data[0]
		//		assert.Equal(t, "api-package-2", updatedAPIPackage.Name)
		//
		//		updatedAPIDefinitions := updatedAPIPackage.APIDefinitions.Data
		//		updatedEventAPIDefinitions := updatedAPIPackage.EventDefinitions.Data
		//
		//		assert.Equal(t, 3, len(updatedAPIDefinitions))
		//		assert.Equal(t, 1, len(updatedEventAPIDefinitions))
		//
		//		//apiIds := getAPIsIds(updatedApp)
		//		//t.Logf("Updated APIs for %s Application", updatedApp.Name)
		//		//logIds(t, apiIds)
		//
		//		// then
		//		this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
		//			// assert previous APIs deleted
		//			for _, api := range updatedAPIDefinitions {
		//				testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, updatedApp.Name, api.ID)
		//			}
		//			// assert previous EventAPIs deleted
		//			for _, eventAPI := range updatedEventAPIDefinitions {
		//				testSuite.K8sResourceChecker.AssertAPIResourcesDeleted(t, updatedApp.Name, eventAPI.ID)
		//			}
		//
		//			// assert updated Application
		//			testSuite.K8sResourceChecker.AssertResourcesForApp(t, updatedApp)
		//			testSuite.ProxyAPIAccessChecker.AssertAPIAccess(t, updatedApp.Name, updatedAPIDefinitions...)
		//		}
		//	},
		//},
		{
			description: "Test case 3: Change auth in all API packages",
			log:         testkit.NewLogger(t, map[string]string{"ApplicationName": "test-app-3"}),
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-3", "provider 3", "", map[string]interface{}{}).
					WithAPIPackages(noAuthAPIPkgInput, oauthAPIPkgInput, basicAuthAPIPkgInput)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				application := this.initialPhaseResult.Application
				apiPackages := application.Packages.Data
				// TODO: assert packages?
				require.Equal(t, 3, len(apiPackages))

				var updatedPackages []*graphql.PackageExt

				// update no auth Package to OAuth
				noAuthPackage, found := getPackageByName(apiPackages, noAuthAPIPackageName)
				require.True(t, found)
				updateNoAuthInput := applications.NewAPIPackageUpdateInput("no auth to oauth package", "", oauth.ToCompassInput())
				newOauthPackage, err := testSuite.CompassClient.UpdateAPIPackage(noAuthPackage.ID, updateNoAuthInput.ToCompassInput())
				require.NoError(t, err)
				updatedPackages = append(updatedPackages, &newOauthPackage)

				// update OAuth Package to Basic Auth
				oauthPackage, found := getPackageByName(apiPackages, oAuthAPIPackageName)
				require.True(t, found)
				updateOauthInput := applications.NewAPIPackageUpdateInput("oauth to basic package", "", basicAuth.ToCompassInput())
				newBasicAuthPackage, err := testSuite.CompassClient.UpdateAPIPackage(oauthPackage.ID, updateOauthInput.ToCompassInput())
				require.NoError(t, err)
				updatedPackages = append(updatedPackages, &newBasicAuthPackage)

				// update Basic Auth Package to no auth
				basicAuthPackage, found := getPackageByName(apiPackages, basicAuthAPIPackageName)
				require.True(t, found)
				updateBasicAuthInput := applications.NewAPIPackageUpdateInput("basic auth to no auth package", "", &graphql.AuthInput{})
				newNoAuthPackage, err := testSuite.CompassClient.UpdateAPIPackage(basicAuthPackage.ID, updateBasicAuthInput.ToCompassInput())
				require.NoError(t, err)
				updatedPackages = append(updatedPackages, &newNoAuthPackage)

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// assert updated APIs
					k8sChecker := testSuite.K8sResourceChecker.NewChecker(t, this.log)

					for _, pkg := range updatedPackages {
						k8sChecker.AssertAPIPackageResources(pkg, application.Name)
					}

					testSuite.ProxyAPIAccessChecker.AssertAPIAccess(t, this.log, application.ID, updatedPackages)
				}
			},
		},
		//{
		//	description: "Test case 4:Fetch new CSRF token and retry if token expired",
		//	initialPhaseInput: func() *applications.ApplicationRegisterInput {
		//		csrfAuth := applications.NewAuth().WithBasicAuth(validUsername, validPassword).
		//			WithCSRF(testSuite.GetMockServiceURL() + mock.CSRFToken.String() + "/valid-csrf-token")
		//
		//		apiPackage := applications.NewAPIPackage("api-package-4", "awesome description 4").
		//			WithAPIDefinitions([]*applications.APIDefinitionInput{jsonAPIInput}).
		//			WithAuth(csrfAuth)
		//
		//		app := applications.NewApplication("test-app-4", "provider 4", "testApp4", map[string]interface{}{}).
		//			WithAPIPackages(apiPackage)
		//
		//		return app
		//	},
		//	initialPhaseAssert: func(t *testing.T, log *testkit.Logger, testSuite *runtimeagent.TestSuite, initialPhaseResult TestApplicationData) {
		//		k8sChecker := testSuite.K8sResourceChecker.NewChecker(t, log)
		//
		//		k8sChecker.AssertResourcesForApp(t, initialPhaseResult.Application)
		//
		//		require.Equal(t, 1, len(initialPhaseResult.Application.Packages.Data))
		//		apiPackage := initialPhaseResult.Application.Packages.Data[0]
		//
		//		// call CSRF API
		//		testSuite.ProxyAPIAccessChecker.
		//
		//		csrfAPI, found := getAPIByName(apiPackage.APIDefinitions.Data, "csrf-api")
		//		require.True(t, found)
		//		response := testSuite.ProxyAPIAccessChecker.CallAccessService(t, initialPhaseResult.Application.Name, csrfAPI.ID, mock.CSERTarget.String()+"/valid-csrf-token")
		//		util.RequireStatus(t, http.StatusOK, response)
		//	},
		//	secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
		//		apiPackage := this.initialPhaseResult.Application.Packages.Data[0]
		//
		//		csrfAPI, found := getAPIByName(apiPackage.APIDefinitions.Data, "csrf-api")
		//		require.True(t, found)
		//
		//		// when updating CSRF API (new token is expected, old token is cached)
		//		modifiedCSRFAuth := applications.NewAuth().WithBasicAuth(validUsername, validPassword).
		//			WithCSRF(testSuite.GetMockServiceURL() + mock.CSRFToken.String() + "/new-csrf-token")
		//		modifiedCSRFAPIInput := applications.NewAPI("csrf-api", "csrf", testSuite.GetMockServiceURL()).WithAuth(modifiedCSRFAuth)
		//		modifiedCSRFAPI, err := testSuite.CompassClient.UpdateAPI(csrfAPI.ID, *modifiedCSRFAPIInput.ToCompassInput())
		//		require.NoError(t, err)
		//
		//		// then
		//		this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
		//			// assert Token URL updated
		//			testSuite.K8sResourceChecker.AssertAPIResources(t, this.initialPhaseResult.Application.Name, modifiedCSRFAPI)
		//
		//			// assert call is made with new token
		//			response := testSuite.ProxyAPIAccessChecker.CallAccessService(t, this.initialPhaseResult.Application.Name, csrfAPI.ID, mock.CSERTarget.String()+"/new-csrf-token")
		//			util.RequireStatus(t, http.StatusOK, response)
		//		}
		//	},
		//},
		{
			description: "Test case 6: Should allow multiple certificates only for specific Application",
			log:         testkit.NewLogger(t, map[string]string{"ApplicationName": "test-app-5"}),
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-5", "provider 5", "", map[string]interface{}{}).
					WithAPIPackages(noAuthAPIPkgInput)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				initialApplication := this.initialPhaseResult.Application

				secondApplicationInput := applications.NewApplication("test-app-5-second", "provider 5 second", "", map[string]interface{}{})

				secondApplication, err := testSuite.CompassClient.CreateApplication(secondApplicationInput.ToCompassInput())
				require.NoError(t, err)
				this.secondPhaseCleanup = func(t *testing.T) {
					this.log.Log(fmt.Sprintf("Running second phase cleanup for %s test case.", this.description))
					removedId, err := testSuite.CompassClient.DeleteApplication(secondApplication.ID)
					require.NoError(t, err)
					assert.Equal(t, secondApplication.ID, removedId)
				}

				secondAppCertificate := testSuite.GenerateCertificateForApplication(t, secondApplication)

				initialAppAdditionalCertificate := testSuite.GenerateCertificateForApplication(t, initialApplication)

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
					// access Initial app with both certs
					testSuite.EventsAPIAccessChecker.AssertEventAPIAccess(t, initialApplication, this.initialPhaseResult.Certificate)
					testSuite.EventsAPIAccessChecker.AssertEventAPIAccess(t, initialApplication, initialAppAdditionalCertificate)

					// access second app with proper cert
					testSuite.EventsAPIAccessChecker.AssertEventAPIAccess(t, secondApplication, secondAppCertificate)

					// should get 403 when trying to send events with wrong certs
					resp, err := testSuite.EventsAPIAccessChecker.SendEvent(t, initialApplication, secondAppCertificate)
					require.NoError(t, err)
					assertForbidden(t, resp)

					resp, err = testSuite.EventsAPIAccessChecker.SendEvent(t, secondApplication, this.initialPhaseResult.Certificate)
					require.NoError(t, err)
					assertForbidden(t, resp)

					resp, err = testSuite.EventsAPIAccessChecker.SendEvent(t, secondApplication, initialAppAdditionalCertificate)
					require.NoError(t, err)
					assertForbidden(t, resp)

					// should get error when trying to send event without certificate
					_, err = testSuite.EventsAPIAccessChecker.SendEvent(t, initialApplication, tls.Certificate{})
					assert.Error(t, err)
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
		testCase.log.Log(fmt.Sprintf("Running initial phase setup for test case: %s", testCase.description))
		appInput := testCase.initialPhaseInput()

		testCase.log.Log("Creating Application...")
		createdApplication, err := testSuite.CompassClient.CreateApplication(appInput.ToCompassInput())
		require.NoError(t, err)

		defer func() {
			testCase.log.Log("Cleaning up Application...")
			removedId, err := testSuite.CompassClient.DeleteApplication(createdApplication.ID)
			assert.NoError(t, err)
			assert.Equal(t, createdApplication.ID, removedId)
		}()

		createdApplications = append(createdApplications, &createdApplication)
		testCase.log.AddField("ApplicationId", createdApplication.ID)

		testCase.log.Log("Generating certificate for Application...")
		certificate := testSuite.GenerateCertificateForApplication(t, createdApplication)

		testCase.log.Log("Creating Application Mapping...")
		err = testSuite.CreateApplicationMapping(createdApplication.Name)
		require.NoError(t, err)

		defer func() {
			testCase.log.Log("Cleaning up Application Mapping...")
			err := testSuite.DeleteApplicationMapping(createdApplication.Name)
			assert.NoError(t, err)
		}()

		testCase.initialPhaseResult = TestApplicationData{
			Application: createdApplication,
			Certificate: certificate,
		}
		t.Logf("Initial test case setup finished for %s test case. Created Application: %s", testCase.description, createdApplication.Name)
	}

	// TODO: consider creating dummy SI to not to wait for Gateway every time

	// Wait for agent to apply config
	waitForAgentToApplyConfig(t, testSuite)

	// Assert initial phase
	for _, testCase := range testCases {
		t.Logf("Asserting initial phase for test case: %s", testCase.description)
		testCase.initialPhaseAssert(t, testCase.log, testSuite, testCase.initialPhaseResult)
	}

	// Setup second phase
	for _, testCase := range testCases {
		t.Logf("Running second phase setup for test case: %s", testCase.description)
		testCase.secondPhaseSetup(t, testSuite, testCase)
		if testCase.secondPhaseCleanup != nil {
			defer testCase.secondPhaseCleanup(t)
		}
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

// TODO: align APIs of both checkers

func assertK8sResourcesAndAPIAccess(t *testing.T, log *testkit.Logger, testSuite *runtimeagent.TestSuite, testData TestApplicationData) {
	log.Log("Waiting for Application to be deployed...")
	testSuite.WaitForApplicationToBeDeployed(t, testData.Application.Name)

	log.Log("Checking K8s resources")
	k8sChecker := testSuite.K8sResourceChecker.NewChecker(t, log)
	k8sChecker.AssertResourcesForApp(t, testData.Application)

	log.Log("Checking API Access")
	testSuite.ProxyAPIAccessChecker.AssertAPIAccess(t, log, testData.Application.ID, testData.Application.Packages.Data)

	log.Log("Checking sending Events")
	testSuite.EventsAPIAccessChecker.AssertEventAPIAccess(t, testData.Application, testData.Certificate)
}

func assertForbidden(t *testing.T, response *http.Response) {
	defer func() {
		err := response.Body.Close()
		assert.NoError(t, err)
	}()

	util.RequireStatus(t, http.StatusForbidden, response)
}

func waitForAgentToApplyConfig(t *testing.T, testSuite *runtimeagent.TestSuite) {
	t.Log("Waiting for Runtime Agent to apply configuration...")
	testSuite.WaitForConfigurationApplication()
}

func getAPIByName(apis []*graphql.APIDefinitionExt, name string) (*graphql.APIDefinitionExt, bool) {
	for _, api := range apis {
		if api.Name == name {
			return api, true
		}
	}

	return nil, false
}

func getPackageByName(packages []*graphql.PackageExt, name string) (*graphql.PackageExt, bool) {
	for _, pkg := range packages {
		if pkg.Name == name {
			return pkg, true
		}
	}

	return nil, false
}

//func getAPIsIds(application compass.Application) []string {
//	ids := make([]string, len(application.APIDefinitions.Data))
//	for i, api := range application.APIDefinitions.Data {
//		ids[i] = api.ID
//	}
//
//	return ids
//}
