package test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/runtimeagent"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	description string
	log         *testkit.Logger

	initialPhaseInput  func() *applications.ApplicationRegisterInput
	initialPhaseAssert func(t *testing.T, log *testkit.Logger, testSuite *runtimeagent.TestSuite, initialPhaseResult TestApplicationData)
	initialPhaseResult TestApplicationData

	secondPhaseSetup   func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase)
	secondPhaseCleanup CleanupFunc
	secondPhaseAssert  func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase)
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
	csrfAPIPackageName      = "csrf-api-pkg"

	jsonAPIName = "some json api"
	yamlAPIName = "some yaml api"
	xmlAPIName  = "some xml api"

	jsonEventAPIName  = "json event API"
	yamlEventAPIName  = "yaml event API"
	emptyEventAPIName = "empty  event API"
)

func TestCompassRuntimeAgentSynchronization(t *testing.T) {
	var emptySpec graphql.CLOB = ""
	var apiSpecData graphql.CLOB = "defaultContent"

	oauthTokenURL := fmt.Sprintf("%s/%s/%s/%s", testSuite.GetMockServiceURL(), mock.OAuthToken, validClientId, validClientSecret)
	basicAuth := applications.NewAuth().WithBasicAuth(validUsername, validPassword)
	oauth := applications.NewAuth().WithOAuth(validClientId, validClientSecret, oauthTokenURL)

	basicAuthWithCSRF := applications.NewAuth().WithBasicAuth(validUsername, validPassword).
		WithCSRF(testSuite.GetMockServiceURL() + mock.CSRFToken.String() + "/valid-csrf-token")

	jsonAPIInput := applications.NewAPI(jsonAPIName, "json api", testSuite.GetMockServiceURL()).WithJsonApiSpec(&apiSpecData)
	yamlAPIInput := applications.NewAPI(yamlAPIName, "yaml api", testSuite.GetMockServiceURL()).WithYamlApiSpec(&apiSpecData)
	xmlAPIInput := applications.NewAPI(xmlAPIName, "xml api", testSuite.GetMockServiceURL()).WithXMLApiSpec(&emptySpec)

	jsonEventAPIInput := applications.NewEventDefinition(jsonEventAPIName, "some json event API").WithJsonEventSpec(&apiSpecData)
	yamlEventAPIInput := applications.NewEventDefinition(yamlEventAPIName, "some ymal event API").WithYamlEventSpec(&apiSpecData)
	emptySpecEventAPIInput := applications.NewEventDefinition(emptyEventAPIName, "some empty event API").WithJsonEventSpec(&emptySpec)

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

	basicAuthWithCSRFPkgInput := applications.NewAPIPackage(csrfAPIPackageName, "").
		WithAPIDefinitions([]*applications.APIDefinitionInput{
			jsonAPIInput,
		}).
		WithAuth(basicAuthWithCSRF)

	// Define test cases
	testCases := []*TestCase{
		{
			description: "Test case 1: Create all types of API packages and remove them",
			log:         testkit.NewLogger(t, map[string]string{"ApplicationName": "test-app-1"}),
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-1", "provider 1", "testApp1", map[string]interface{}{}).
					WithAPIPackages(noAuthAPIPkgInput, basicAuthAPIPkgInput, oauthAPIPkgInput, basicAuthWithCSRFPkgInput)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
				// when removing all APIs individually
				application := this.initialPhaseResult.Application
				apiPackages := application.Packages.Data
				require.Equal(t, 4, len(apiPackages))

				// remove Packages
				for _, pkg := range application.Packages.Data {
					id, err := testSuite.CompassClient.DeleteAPIPackage(pkg.ID)
					require.NoError(t, err)
					require.Equal(t, pkg.ID, id)
				}

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
					// assert APIs deleted
					for _, pkg := range apiPackages {
						testSuite.K8sResourceChecker.AssertAPIPackageDeleted(t, this.log, pkg, application.Name)
					}
				}
			},
		},
		{
			description: "Test case 2: Add and delete APIs in package",
			log:         testkit.NewLogger(t, map[string]string{"ApplicationName": "test-app-2"}),
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				noAuthAPIPkgInput := applications.NewAPIPackage(noAuthAPIPackageName, "no auth pkg description").
					WithAPIDefinitions([]*applications.APIDefinitionInput{jsonAPIInput}).
					WithEventDefinitions([]*applications.EventDefinitionInput{jsonEventAPIInput})

				basicAuthAPIPkgInput := applications.NewAPIPackage(basicAuthAPIPackageName, "basic auth pkg description").
					WithAPIDefinitions([]*applications.APIDefinitionInput{jsonAPIInput}).
					WithEventDefinitions([]*applications.EventDefinitionInput{jsonEventAPIInput}).
					WithAuth(basicAuth)

				oauthAPIPkgInput := applications.NewAPIPackage(oAuthAPIPackageName, "oauth pkg description").
					WithAPIDefinitions([]*applications.APIDefinitionInput{jsonAPIInput}).
					WithEventDefinitions([]*applications.EventDefinitionInput{jsonEventAPIInput}).
					WithAuth(oauth)

				return applications.NewApplication("test-app-2", "provider 2", "", map[string]interface{}{}).
					WithAPIPackages(noAuthAPIPkgInput, oauthAPIPkgInput, basicAuthAPIPkgInput)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
				// when
				application := this.initialPhaseResult.Application
				apiPackages := application.Packages.Data

				require.Equal(t, 3, len(apiPackages))

				// create new API
				newAPI := xmlAPIInput

				for i, pkg := range apiPackages {
					addedAPI, err := testSuite.CompassClient.AddAPIDefinitionToPackage(pkg.ID, *newAPI.ToCompassInput())
					require.NoError(t, err)
					apiPackages[i].APIDefinitions.Data = append(apiPackages[i].APIDefinitions.Data, addedAPI)
				}

				// delete existing API
				for i, pkg := range apiPackages {
					apiToDelete, found := getAPIByName(pkg.APIDefinitions.Data, jsonAPIName)
					require.True(t, found)

					_, err := testSuite.CompassClient.DeleteAPIDefinition(apiToDelete.ID)
					require.NoError(t, err)
					apiPackages[i].APIDefinitions.Data = deleteAPI(apiPackages[i].APIDefinitions.Data, apiToDelete.ID)
				}

				// create new EventAPI
				newEventAPI := emptySpecEventAPIInput

				for i, pkg := range apiPackages {
					addedEventAPI, err := testSuite.CompassClient.AddEventAPIToPackage(pkg.ID, *newEventAPI.ToCompassInput())
					require.NoError(t, err)
					apiPackages[i].EventDefinitions.Data = append(apiPackages[i].EventDefinitions.Data, addedEventAPI)
				}

				// delete existing EventAPI
				for i, pkg := range apiPackages {
					eventAPIToDelete, found := getEventAPIByName(pkg.EventDefinitions.Data, jsonEventAPIName)
					require.True(t, found)

					_, err := testSuite.CompassClient.DeleteEventAPI(eventAPIToDelete.ID)
					require.NoError(t, err)
					apiPackages[i].EventDefinitions.Data = deleteEventAPI(apiPackages[i].EventDefinitions.Data, eventAPIToDelete.ID)
				}

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
					// override packages in Application
					application.Packages.Data = apiPackages

					testAppData := TestApplicationData{
						Application: application,
						Certificate: this.initialPhaseResult.Certificate,
					}
					assertK8sResourcesAndAPIAccess(t, this.log, testSuite, testAppData)
				}
			},
		},
		{
			description: "Test case 3: Change auth in all API packages",
			log:         testkit.NewLogger(t, map[string]string{"ApplicationName": "test-app-3"}),
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-3", "provider 3", "", map[string]interface{}{}).
					WithAPIPackages(noAuthAPIPkgInput, oauthAPIPkgInput, basicAuthAPIPkgInput)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
				// when
				application := this.initialPhaseResult.Application
				apiPackages := application.Packages.Data
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
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
					// assert updated APIs
					for _, pkg := range updatedPackages {
						testSuite.K8sResourceChecker.AssertAPIPackageResources(t, this.log, pkg, application.Name)
					}

					testSuite.ProxyAPIAccessChecker.AssertAPIAccess(t, this.log, application.ID, updatedPackages)
				}
			},
		},
		{
			description: "Test case 4: Should add API package with CSRF to application",
			log:         testkit.NewLogger(t, map[string]string{"ApplicationName": "test-app-4"}),
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-4", "provider 4", "testApp4", map[string]interface{}{})
			},
			initialPhaseAssert: func(t *testing.T, log *testkit.Logger, testSuite *runtimeagent.TestSuite, initialPhaseResult TestApplicationData) {
				testSuite.K8sResourceChecker.AssertResourcesForApp(t, log, initialPhaseResult.Application)
			},
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
				// when
				application := this.initialPhaseResult.Application

				addedPackage, err := testSuite.CompassClient.AddAPIPackage(application.ID, *basicAuthWithCSRFPkgInput.ToCompassInput())
				require.NoError(t, err)

				// then
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
					packages := []*graphql.PackageExt{&addedPackage}
					testSuite.K8sResourceChecker.AssertAPIPackageResources(t, this.log, &addedPackage, application.Name)
					testSuite.ProxyAPIAccessChecker.AssertAPIAccess(t, this.log, application.ID, packages)
				}
			},
		},
		{
			description: "Test case 5: Should allow multiple certificates only for specific Application",
			log:         testkit.NewLogger(t, map[string]string{"ApplicationName": "test-app-5"}),
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				return applications.NewApplication("test-app-5", "provider 5", "", map[string]interface{}{}).
					WithAPIPackages(noAuthAPIPkgInput)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
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
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
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
		{
			description: "Test case 6: Create API packages with Headers and Query Params",
			log:         testkit.NewLogger(t, map[string]string{"ApplicationName": "test-app-6"}),
			initialPhaseInput: func() *applications.ApplicationRegisterInput {
				headersAPIPkgInput := applications.NewAPIPackage("headers api package", "so much headers").
					WithAPIDefinitions(
						[]*applications.APIDefinitionInput{
							jsonAPIInput,
						}).
					WithAuth(applications.NewAuth().WithHeaders(map[string][]string{"someheader": {"some-value"}}))

				queryAPIPkgInput := applications.NewAPIPackage("query api", "query").
					WithAPIDefinitions(
						[]*applications.APIDefinitionInput{
							jsonAPIInput,
						}).
					WithAuth(applications.NewAuth().WithQueryParams(map[string][]string{"somequery": {"some-value"}}))

				return applications.NewApplication("test-app-6", "provider 6", "testApp6", map[string]interface{}{}).
					WithAPIPackages(headersAPIPkgInput, queryAPIPkgInput)
			},
			initialPhaseAssert: assertK8sResourcesAndAPIAccess,
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {
				// no need to do any second phase
				this.secondPhaseAssert = func(t *testing.T, testSuite *runtimeagent.TestSuite, this *TestCase) {}
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

		defer func(testCase *TestCase) {
			testCase.log.Log("Cleaning up Application...")
			removedId, err := testSuite.CompassClient.DeleteApplication(createdApplication.ID)
			assert.NoError(t, err)
			assert.Equal(t, createdApplication.ID, removedId)
		}(testCase)

		createdApplications = append(createdApplications, &createdApplication)
		testCase.log.AddField("ApplicationId", createdApplication.ID)

		testCase.log.Log("Generating certificate for Application...")
		certificate := testSuite.GenerateCertificateForApplication(t, createdApplication)

		testCase.log.Log("Creating Application Mapping...")
		err = testSuite.CreateApplicationMapping(createdApplication.Name)
		require.NoError(t, err)

		defer func(testCase *TestCase) {
			testCase.log.Log("Cleaning up Application Mapping...")
			err := testSuite.DeleteApplicationMapping(createdApplication.Name)
			assert.NoError(t, err)
		}(testCase)

		testCase.initialPhaseResult = TestApplicationData{
			Application: createdApplication,
			Certificate: certificate,
		}
		t.Logf("Initial test case setup finished for %s test case. Created Application: %s", testCase.description, createdApplication.Name)
	}

	// Wait for agent to apply config
	waitForAgentToApplyConfig(t, testSuite)

	// Assert initial phase
	for _, testCase := range testCases {
		testCase.log.Log(fmt.Sprintf("Asserting initial phase for test case: %s", testCase.description))
		testCase.initialPhaseAssert(t, testCase.log, testSuite, testCase.initialPhaseResult)
	}

	// Setup second phase
	for _, testCase := range testCases {
		testCase.log.Log(fmt.Sprintf("Running second phase setup for test case: %s", testCase.description))
		testCase.secondPhaseSetup(t, testSuite, testCase)
		if testCase.secondPhaseCleanup != nil {
			defer func(testCase *TestCase) {
				testCase.secondPhaseCleanup(t)
			}(testCase)
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
		testCase.log.Log(fmt.Sprintf("Asserting second phase for test case: %s", testCase.description))
		testCase.secondPhaseAssert(t, testSuite, testCase)
	}
}

func assertK8sResourcesAndAPIAccess(t *testing.T, log *testkit.Logger, testSuite *runtimeagent.TestSuite, testData TestApplicationData) {
	log.Log("Waiting for Application to be deployed...")
	testSuite.WaitForApplicationToBeDeployed(t, testData.Application.Name)

	log.Log("Checking K8s resources")
	testSuite.K8sResourceChecker.AssertResourcesForApp(t, log, testData.Application)

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

func getEventAPIByName(apis []*graphql.EventAPIDefinitionExt, name string) (*graphql.EventAPIDefinitionExt, bool) {
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

func deleteAPI(apis []*graphql.APIDefinitionExt, id string) []*graphql.APIDefinitionExt {
	index := -1
	for i, api := range apis {
		if api.ID == id {
			index = i
			break
		}
	}

	if index != -1 {
		return append(apis[:index], apis[index+1:]...)
	}

	return apis
}

func deleteEventAPI(apis []*graphql.EventAPIDefinitionExt, id string) []*graphql.EventAPIDefinitionExt {
	index := -1
	for i, api := range apis {
		if api.ID == id {
			index = i
			break
		}
	}

	if index != -1 {
		return append(apis[:index], apis[index+1:]...)
	}

	return apis
}
