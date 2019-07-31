package test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/runtimeagent"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO - mock service to which you call

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
//
//type testPhase struct {
//	appId           string
//	appName         string
//	appDescription  string
//	appLabels       map[string][]string
//	apisInputs      []*applications.APIDefinitionInput
//	eventAPIsInputs []*applications.EventAPIDefinitionInput
//	// TODO - docs
//	operation operation
//	result    compass.Application
//}
//
//type testPhase2 struct {
//	setup               func(this testPhase2, t *testing.T)
//	assertFunctionality func(this testPhase2, t *testing.T)
//	assertResources     func(this testPhase2, t *testing.T)
//}

type testCase struct {
	description string

	initialPhaseInput  func() *applications.ApplicationInput
	initialPhaseResult compass.Application

	secondPhaseSetup  func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase)
	secondPhaseAssert func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase)
}

type initialPhase struct {
	applicationInput func() *applications.ApplicationInput
}

//type secondPhase struct {
//	applicationOperation resourceOperation
//	apiOperations        []resourceOperation
//	eventAPIOperations   []resourceOperation
//	documentOperations   []resourceOperation
//}

//type secondPhase struct {
//	setup
//}

type resourceOperation struct {
	input     interface{}
	operation operation
}

type operation int

const (
	doNothingOperation operation = iota
	createOperation
	// Update Application APIs, EventAPIs and Documents
	updateOperation
	deleteOperation
	// Replace Application calling updateApplication mutation
	replaceOperation
)

// TODO
const (
	validPassword     = ""
	validUsername     = ""
	validClientId     = ""
	validClientSecret = ""

	oauthURLPath = ""
)

// TODO - check how it can be done in e2e?
// TODO - can I levrage steps here?

func TestCompassRuntimeAgentSynchronization_TestCases(t *testing.T) {

	noAuthAPIInput := applications.NewAPI("no-auth-api", "no auth api", testSuite.GetMockServiceURL())
	basicAuthAPIInput := applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).
		WithAuth(applications.NewAuth().WithBasicAuth(validUsername, validPassword))
	oauthAPIInput := applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).
		WithAuth(applications.NewAuth().WithOAuth(validClientId, validClientSecret, testSuite.GetMockServiceURL()+oauthURLPath))

	testCases := []*testCase{
		{
			description: "test case 1",
			initialPhaseInput: func() *applications.ApplicationInput {
				return applications.NewApplication("test-app1", "testApp1", map[string][]string{}).
					WithAPIs(
						[]*applications.APIDefinitionInput{
							noAuthAPIInput,
							basicAuthAPIInput,
							oauthAPIInput,
						})
			},
			secondPhaseSetup: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				//application := this.initialPhaseResult
				//
				//newNoAuthAPI := applications.NewAPI("new-no-auth-api", "", testSuite.GetMockServiceURL())
				//newOauthAPI := applications.NewAPI("new-oauth-api", "", testSuite.GetMockServiceURL()).
				//	WithAuth(applications.NewAuth().WithOAuth(validClientId, validClientSecret, oauthURLPath))
				//newBasicAuthAPI := applications.NewAPI("new-basic-auth-api", "", testSuite.GetMockServiceURL()).
				//	WithAuth(applications.NewAuth().WithBasicAuth(validUsername, validPassword))
				//
				//api, err := testSuite.CompassClient.CreateAPI(application.ID, newNoAuthAPI.ToCompassInput())

			},
			secondPhaseAssert: func(t *testing.T, testSuite *runtimeagent.TestSuite, this *testCase) {
				// when
				//application := this.initialPhaseResult

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

		// TODO - assert with input?

		testCase.initialPhaseResult = response
		logrus.Infof("Initial test case setup finished for %s test case", testCase.description)
	}

	// Wait for agent to apply config
	logrus.Info("Waiting for Runtime Agent to apply initial configuration...")
	time.Sleep(15 * time.Second)

	// Assert initial phase
	for _, testCase := range testCases {
		logrus.Infof("Asserting initial phase for test case: %s", testCase.description)

		logrus.Infof("Checking K8s resources")
		testSuite.K8sResourceChecker.AssertResourcesForApp(t, testCase.initialPhaseResult)

		logrus.Infof("Checking API Access")
		testSuite.APIAccessChecker.AssertAPIAccess(t, testCase.initialPhaseResult.APIs.Data)
		// TODO - how to do api check if expected status will be different than 200? Separate test case?
	}

	// Setup second phase
	for _, testCase := range testCases {
		logrus.Infof("Running second phase setup for test case: %s", testCase.description)
		testCase.secondPhaseSetup(t, testSuite, testCase)
		logrus.Infof("Initial test case setup finished for %s test case", testCase.description)
	}

	// Wait for agent to apply config
	logrus.Info("Waiting for Runtime Agent to apply initial configuration...")
	time.Sleep(15 * time.Second)

	// Assert second phase
	for _, testCase := range testCases {
		logrus.Infof("Asserting second phase for test case: %s", testCase.description)
		testCase.secondPhaseAssert(t, testSuite, testCase)
	}
}

//func TestCompassRuntimeAgentSynchronization_Operations(t *testing.T) {
//
//	noAuthAPIInput := applications.NewAPI("no-auth-1", "no auth api 1", testSuite.GetMockServiceURL())
//	basicAuthAPIInput := applications.NewAPI("basic-auth-1", "basic auth api 1", testSuite.GetMockServiceURL()).
//		WithAuth(applications.NewAuth().WithBasicAuth(validUsername, validPassword))
//	oauthAPIInput := applications.NewAPI("oauth-auth-1", "oauth api 1", testSuite.GetMockServiceURL()).
//		WithAuth(applications.NewAuth().WithOAuth(validClientId, validClientSecret, testSuite.GetMockServiceURL()+oauthURLPath))
//
//	testCases := []testCase{
//		{
//			description: "test case 1",
//			initialPhase: initialPhase{
//				applicationInput: func() *applications.ApplicationInput {
//					application := applications.NewApplication("test-app1", "testApp1", map[string][]string{})
//					return application.WithAPIs([]*applications.APIDefinitionInput{noAuthAPIInput, basicAuthAPIInput, oauthAPIInput})
//				},
//			},
//			//initialPhaseCheck: func(t *testing.T, testSuite runtimeagent.TestSuite, application compass.Application) {
//			//	// Assert resource created for all apis (or even pass application input?)
//			//	// Assert you are able to access all apis
//			//},
//			secondPhase: secondPhase{
//				applicationOperation: resourceOperation{operation: doNothingOperation, input: nil},
//				apiOperations: []resourceOperation{
//					{operation: doNothingOperation},
//					{operation: deleteOperation},
//					{operation: deleteOperation},
//				},
//				eventAPIOperations: nil,
//				documentOperations: nil,
//			},
//		},
//	}
//
//	// Setup first phase - when Applications crated
//	for _, testCase := range testCases {
//		logrus.Infof("Running initial phase setup for test case: %s", testCase.description)
//		appInput := testCase.initialPhase.applicationInput()
//
//		logrus.Info("Creating Application...")
//		response, err := testSuite.CompassClient.CreateApplication(appInput.ToCompassInput())
//		require.NoError(t, err)
//
//		// TODO - assert with input?
//
//		testCase.initialPhaseResult = response
//		logrus.Infof("Initial test case setup finished for %s test case", testCase.description)
//	}
//
//	// Wait for agent to apply config
//	logrus.Info("Waiting for Runtime Agent to apply initial configuration...")
//	time.Sleep(15 * time.Second)
//
//	// Then
//	for _, testCase := range testCases {
//		logrus.Infof("Asserting initial phase for test case: %s", testCase.description)
//
//		logrus.Infof("Checking K8s resources")
//		testSuite.K8sResourceChecker.AssertResourcesForApp(t, testCase.initialPhaseResult)
//
//		logrus.Infof("Checking API Access")
//		testSuite.APIAccessChecker.AssertAPIAccess(t, testCase.initialPhaseResult.APIs.Data)
//		// TODO - how to do api check if expected status will be different than 200? Separate test case?
//	}
//
//	// Setup second phase - when Applications updated/deleted
//	for _, testCase := range testCases {
//		logrus.Infof("Running second phase setup for test case: %s", testCase.description)
//
//		var err error
//		applicationResult := compass.Application{}
//
//		// TODO - consider improving second phase stuff (include operation only if app update)
//		switch testCase.secondPhase.applicationOperation.operation {
//		case deleteOperation:
//			id, err := testSuite.CompassClient.DeleteApplication(testCase.initialPhaseResult.ID)
//			assert.NoError(t, err)
//			assert.Equal(t, testCase.initialPhaseResult.ID, id)
//			continue
//		case createOperation:
//			appInput, ok := testCase.secondPhase.applicationOperation.input.(*applications.ApplicationInput)
//			require.True(t, ok)
//
//			logrus.Info("Creating Application...")
//			applicationResult, err = testSuite.CompassClient.CreateApplication(appInput.ToCompassInput())
//			require.NoError(t, err)
//			continue
//		case replaceOperation:
//			appInput, ok := testCase.secondPhase.applicationOperation.input.(*applications.ApplicationInput)
//			require.True(t, ok)
//
//			logrus.Info("Updating Application...")
//			applicationResult, err = testSuite.CompassClient.CreateApplication(appInput.ToCompassInput())
//			require.NoError(t, err)
//			continue
//		case updateOperation:
//			for i, apiOperation := range testCase.secondPhase.apiOperations {
//
//
//				switch apiOperation.operation {
//				case createOperation:
//				// create API
//				case updateOperation:
//					//update API
//				case deleteOperation:
//					// delete API
//				}
//			}
//		}
//
//		logrus.Infof("Second phase test case setup finished for %s test case", testCase.description)
//	}
//}

func TestCompassRuntimeAgentSynchronization_Simple(t *testing.T) {
	// Step 1 - create Apps
	// given
	//// APIs
	noAuthAPIInput := applications.NewAPI("no-auth-api", "no auth api", testSuite.GetMockServiceURL())
	basicAuthAPIInput := applications.NewAPI("basic-auth-api", "basic auth api", testSuite.GetMockServiceURL()).
		WithAuth(applications.NewAuth().WithBasicAuth(validUsername, validPassword))
	oauthAPIInput := applications.NewAPI("oauth-auth-api", "oauth api", testSuite.GetMockServiceURL()).
		WithAuth(applications.NewAuth().WithOAuth(validClientId, validClientSecret, testSuite.GetMockServiceURL()+oauthURLPath))

	//// EventAPIs

	//// Documents

	//// Applications
	appWithAllAuths := applications.NewApplication(
		"test-app1",
		"application with services with all auth methods",
		map[string][]string{},
	).WithAPIs([]*applications.APIDefinitionInput{noAuthAPIInput, basicAuthAPIInput, oauthAPIInput})

	// when
	logrus.Info("Creating Application...")
	response, err := testSuite.CompassClient.CreateApplication(appWithAllAuths.ToCompassInput())
	require.NoError(t, err)
	defer func() {
		logrus.Infof("Cleaning up %s Application...", response.ID)
		removedId, err := testSuite.CompassClient.DeleteApplication(response.ID)
		require.NoError(t, err)
		assert.Equal(t, response.ID, removedId)
	}()

	// then
	// assert

	// Step 2 - updated/delete Apps
	// given
	// when
	// then

}

//
//func TestCompassRuntimeAgentSynchronization(t *testing.T) {
//
//	tenant := "acc-tests"
//
//	client := compass.NewCompassClient("http://localhost:3000/graphql", tenant, "")
//
//	// TODO - create 3 APIs of each kind
//	//// One should remain the same after update
//	//// One should be deleted
//	//// One should be updated to something else
//
//	noAuthAPIInput := applications.NewAPI("no-auth-1", "no auth api for app 1", "") // TODO - mock service URL
//
//	basicAuthAPIInput := applications.NewAPI("no-auth-1", "no auth api for app 1", ""). // TODO - mock service URL
//												WithAuth(applications.NewAuth().WithBasicAuth("", "")) // TODO - user pswd
//
//	oauthAPIInput := applications.NewAPI("no-auth-1", "no auth api for app 1", ""). // TODO - mock service URL
//											WithAuth(applications.NewAuth().WithOAuth("", "", "")) // TODO - clientId secret url
//
//	application := applications.NewApplication("test-app1", "testApp1", map[string][]string{})
//	application = application.WithAPIs([]*applications.APIDefinitionInput{
//		noAuthAPIInput, basicAuthAPIInput, oauthAPIInput,
//	})
//
//	logrus.Info("Creating Application...")
//	response, err := client.CreateApplication(application.ToCompassInput())
//	require.NoError(t, err)
//	defer func() {
//		logrus.Infof("Cleaning up %s Application...", response.ID)
//		removedId, err := client.DeleteApplication(response.ID)
//		require.NoError(t, err)
//		assert.Equal(t, response.ID, removedId)
//	}()
//
//	assert.NotEmpty(t, response.ID)
//	assert.Equal(t, 3, len(response.APIs.Data))
//
//	// Create application 1
//	//// With BasicAuth, Oauth, NoAuth APIs
//	//// With Events
//
//	// TODO: consider checking CompassConnection CR to decide when to start checks
//	logrus.Info("Waiting for Runtime Agent to apply configuration...")
//	time.Sleep(45 * time.Second)
//
//	// TODO - assertions
//
//	// Verify that resources were created
//
//	// Create application 2 and update application 1
//
//	//for i, testCase := range testCases {
//	//	// Perform first operation for all test cases
//	//}
//	//
//	//for i, testCase := range testCases {
//	//	// Perform assertions for first operation
//	//}
//	//
//	//for i, testCase := range testCases {
//	//	// Perform second operation for all test cases
//	//}
//	//
//	//for i, testCase := range testCases {
//	//	// Perform assertions for second operation
//	//}
//}
