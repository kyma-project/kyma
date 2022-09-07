package director

import (
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gcli "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/third_party/machinebox/graphql"

	gql "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/graphql"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/oauth"
	oauthmocks "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/oauth/mocks"
)

const (
	testAppName          = "Test-application-123"
	applicationTestingID = "test-application-ID-12345"
	testAppScenario      = "Testing-scenario"

	validTokenValue = "12345"
	tenantValue     = "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"
	oneTimeToken    = "54321"
	connectorURL    = "https://kyma.cx/connector/graphql"

	//expectedRegisterApplicationQuery = `mutation {
	//result: registerApplication(in: {
	//	name: "Test-application-123",
	//	labels: { scenarios: ["Testing-scenario"] }
	//}) { id } }`

	expectedRegisterApplicationQuery = `mutation {
	result: registerApplication(in: {
		name: "Test-application-123"
	}) { id } }`

	expectedDeleteApplicationQuery = `mutation {
	result: unregisterApplication(id: "test-application-ID-12345") {
		id
	} }`
)

var (
	futureExpirationTime = time.Now().Add(time.Duration(60) * time.Minute).Unix()
	passedExpirationTime = time.Now().Add(time.Duration(60) * time.Minute * -1).Unix()
)

func TestDirectorClient_ApplicationRegistering(t *testing.T) {
	expectedRequest := gcli.NewRequest(expectedRegisterApplicationQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedRequest.Header.Set(TenantHeader, tenantValue)

	t.Run("Should register application and return new application ID when the Director access token is valid", func(t *testing.T) {
		// given
		expectedResponse := &graphql.Application{
			Name: testAppName,
			BaseEntity: &graphql.BaseEntity{
				ID: applicationTestingID,
			},
		}
		expectedID := applicationTestingID

		gqlClient := gql.NewQueryAssertClient(t, nil, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateApplicationResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		})

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		receivedApplicationID, err := configClient.RegisterApplication(testAppName, testAppScenario)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedID, receivedApplicationID)
	})

	t.Run("Should not register application and return an error when the Director access token is empty", func(t *testing.T) {
		// given
		token := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient, tenantValue)

		// when
		receivedApplicationID, err := configClient.RegisterApplication(testAppName, testAppScenario)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedApplicationID)
	})

	t.Run("Should not register Application and return an error when the Director access token is expired", func(t *testing.T) {
		// given
		expiredToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  passedExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(expiredToken, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient, tenantValue)

		// when
		receivedApplicationID, err := configClient.RegisterApplication(testAppName, testAppScenario)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedApplicationID)
	})

	t.Run("Should not register Application and return error when the client fails to get an access token for Director", func(t *testing.T) {
		// given
		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(oauth.Token{}, errors.New("Failed token error"))

		configClient := NewDirectorClient(nil, mockedOAuthClient, tenantValue)

		// when
		receivedApplicationID, err := configClient.RegisterApplication(testAppName, testAppScenario)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedApplicationID)
	})

	t.Run("Should return error when the result of the call to Director service is nil", func(t *testing.T) {
		// given
		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		gqlClient := gql.NewQueryAssertClient(t, nil, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateApplicationResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		receivedApplicationID, err := configClient.RegisterApplication(testAppName, testAppScenario)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedApplicationID)
	})

	t.Run("Should return error when Director fails to register Runtime ", func(t *testing.T) {
		// given
		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		gqlClient := gql.NewQueryAssertClient(t, errors.New("error"), []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateApplicationResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		receivedRuntimeID, err := configClient.RegisterApplication(testAppName, testAppScenario)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})
}

func TestDirectorClient_ApplicationUnregistering(t *testing.T) {
	expectedRequest := gcli.NewRequest(expectedDeleteApplicationQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedRequest.Header.Set(TenantHeader, tenantValue)

	t.Run("Should unregister runtime of given ID and return no error when the Director access token is valid", func(t *testing.T) {
		// given
		expectedResponse := &graphql.Application{
			Name: testAppName,
			BaseEntity: &graphql.BaseEntity{
				ID: applicationTestingID,
			},
		}
		// expectedID := applicationTestingID

		gqlClient := gql.NewQueryAssertClient(t, nil, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteApplicationResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		})

		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.UnregisterApplication(applicationTestingID)

		// then
		assert.NoError(t, err)
	})

	t.Run("Should not unregister runtime and return an error when the Director access token is empty", func(t *testing.T) {
		// given
		emptyToken := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(emptyToken, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient, tenantValue)

		// when
		err := configClient.UnregisterApplication(applicationTestingID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should not unregister register runtime and return an error when the Director access token is expired", func(t *testing.T) {
		// given
		expiredToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  passedExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(expiredToken, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient, tenantValue)

		// when
		err := configClient.UnregisterApplication(applicationTestingID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should not unregister Runtime and return error when the client fails to get an access token for Director", func(t *testing.T) {
		// given
		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(oauth.Token{}, errors.New("Failed token error"))

		configClient := NewDirectorClient(nil, mockedOAuthClient, tenantValue)

		// when
		err := configClient.UnregisterApplication(applicationTestingID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should return error when the result of the call to Director service is nil", func(t *testing.T) {
		// given
		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		// given
		gqlClient := gql.NewQueryAssertClient(t, nil, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteApplicationResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.UnregisterApplication(applicationTestingID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should return error when Director fails to delete Runtime", func(t *testing.T) {
		// given
		gqlClient := gql.NewQueryAssertClient(t, errors.New("error"), []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteApplicationResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.UnregisterApplication(applicationTestingID)

		// then
		assert.Error(t, err)
	})

	// unusual and strange case
	t.Run("Should return error when Director returns bad ID after Deleting", func(t *testing.T) {
		// given
		expectedResponse := &graphql.Application{
			Name: testAppName,
			BaseEntity: &graphql.BaseEntity{
				ID: "badID",
			},
		}

		gqlClient := gql.NewQueryAssertClient(t, nil, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteApplicationResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		})

		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.UnregisterApplication(applicationTestingID)

		// then
		assert.Error(t, err)
	})
}
