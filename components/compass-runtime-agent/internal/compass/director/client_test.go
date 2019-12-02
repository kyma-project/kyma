package director

import (
	"github.com/stretchr/testify/require"
	"kyma-project.io/compass-runtime-agent/internal/config"

	kymamodel "kyma-project.io/compass-runtime-agent/internal/kyma/model"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gql "kyma-project.io/compass-runtime-agent/internal/graphql"

	gcli "github.com/machinebox/graphql"

	"testing"
)

const (
	tenant    = "tenant"
	runtimeId = "runtimeId"

	expectedAppsForRuntimeQuery = `query {
	result: applicationsForRuntime(runtimeID: "runtimeId") {
		data {
		id
		name
		description
		labels
		apiDefinitions {data {
				id
		name
		description
		spec {data
		format
		type}
		targetURL
		group
		auth(runtimeID: "runtimeId") {runtimeID
		auth {credential {
				... on BasicCredentialData {
					username
					password
				}
				...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				}
			}
			additionalHeaders
			additionalQueryParams
			requestAuth { 
			  csrf {
				tokenEndpointURL
				credential {
				  ... on BasicCredentialData {
				  	username
					password
				  }
				  ...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				  }
			    }
				additionalHeaders
				additionalQueryParams
			}
			}
		}}
		defaultAuth {credential {
				... on BasicCredentialData {
					username
					password
				}
				...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				}
			}
			additionalHeaders
			additionalQueryParams
			requestAuth { 
			  csrf {
				tokenEndpointURL
				credential {
				  ... on BasicCredentialData {
				  	username
					password
				  }
				  ...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				  }
			    }
				additionalHeaders
				additionalQueryParams
			}
			}
		}
		version {value
		deprecated
		deprecatedSince
		forRemoval}
	}
	pageInfo {startCursor
		endCursor
		hasNextPage}
	totalCount
	}
		eventDefinitions {data {
		
			id
			applicationID
			name
			description
			group 
			spec {data
		type
		format}
			version {value
		deprecated
		deprecatedSince
		forRemoval}
		
	}
	pageInfo {startCursor
		endCursor
		hasNextPage}
	totalCount
	}
		documents {data {
		
		id
		applicationID
		title
		displayName
		description
		format
		kind
		data
	}
	pageInfo {startCursor
		endCursor
		hasNextPage}
	totalCount
	}
		auths {id}
	
	}
	pageInfo {startCursor
		endCursor
		hasNextPage}
	totalCount
	
	}
}`

	expectedSetEventsURLLabelQuery = `mutation {
		result: setRuntimeLabel(runtimeID: "runtimeId", key: "runtime/event_service_url", value: "https://gateway.kyma.local") {
			key
			value
		}
	}`
	expectedSetConsoleURLLabelQuery = `mutation {
		result: setRuntimeLabel(runtimeID: "runtimeId", key: "runtime/console_url", value: "https://console.kyma.local") {
			key
			value
		}
	}`
)

var (
	runtimeConfig = config.RuntimeConfig{
		RuntimeId: runtimeId,
		Tenant:    tenant,
	}
)

func TestConfigClient_FetchConfiguration(t *testing.T) {

	expectedRequest := gcli.NewRequest(expectedAppsForRuntimeQuery)
	expectedRequest.Header.Set(TenantHeader, tenant)

	t.Run("should fetch configuration", func(t *testing.T) {
		// given
		expectedResponse := &ApplicationPage{
			Data: []*Application{
				{
					ID:   "abcd-efgh",
					Name: "App1",
				},
				{
					ID:   "ijkl-mnop",
					Name: "App2",
				},
				{
					ID:    "asda-oqiu",
					Name:  "App3",
					Auths: []*graphql.SystemAuth{&graphql.SystemAuth{"asd", nil}},
				},
			},
			PageInfo:   &graphql.PageInfo{},
			TotalCount: 3,
		}

		expectedApps := []kymamodel.Application{
			{
				Name:           "App1",
				ID:             "abcd-efgh",
				SystemAuthsIDs: make([]string, 0),
			},
			{
				ID:             "ijkl-mnop",
				Name:           "App2",
				SystemAuthsIDs: make([]string, 0),
			},
			{
				ID:             "asda-oqiu",
				Name:           "App3",
				SystemAuthsIDs: []string{"asd"},
			},
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*ApplicationsForRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		applicationsResponse, err := configClient.FetchConfiguration()

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedApps, applicationsResponse)
	})

	t.Run("should return empty array if no Apps for Runtime", func(t *testing.T) {
		// given
		expectedResponse := &ApplicationPage{
			Data:       nil,
			PageInfo:   &graphql.PageInfo{},
			TotalCount: 0,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*ApplicationsForRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		applicationsResponse, err := configClient.FetchConfiguration()

		// then
		require.NoError(t, err)
		assert.Empty(t, applicationsResponse)
	})

	t.Run("should return error when result is nil", func(t *testing.T) {
		// given
		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*ApplicationsForRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		}, expectedRequest)

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		applicationsResponse, err := configClient.FetchConfiguration()

		// then
		require.Error(t, err)
		assert.Empty(t, applicationsResponse)
	})

	t.Run("should return error when failed to fetch Applications", func(t *testing.T) {
		// given
		gqlClient := gql.NewQueryAssertClient(t, true, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*ApplicationsForRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
		}, expectedRequest)

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		applicationsResponse, err := configClient.FetchConfiguration()

		// then
		require.Error(t, err)
		assert.Nil(t, applicationsResponse)
	})
}

func TestConfigClient_SetURLsLabels(t *testing.T) {

	runtimeURLsConfig := RuntimeURLsConfig{
		EventsURL:  "https://gateway.kyma.local",
		ConsoleURL: "https://console.kyma.local",
	}

	expectedSetEventsURLRequest := gcli.NewRequest(expectedSetEventsURLLabelQuery)
	expectedSetEventsURLRequest.Header.Set(TenantHeader, tenant)
	expectedSetConsoleURLRequest := gcli.NewRequest(expectedSetConsoleURLLabelQuery)
	expectedSetConsoleURLRequest.Header.Set(TenantHeader, tenant)

	newSetExpectedLabelFunc := func(expectedResponses []*graphql.Label) func(t *testing.T, r interface{}) {
		var responseIndex = 0

		return func(t *testing.T, r interface{}) {
			cfg, ok := r.(*SetRuntimeLabelResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponses[responseIndex]
			responseIndex++
		}
	}

	t.Run("should set URLs as labels", func(t *testing.T) {

		expectedResponses := []*graphql.Label{
			{
				Key:   eventsURLLabelKey,
				Value: runtimeURLsConfig.EventsURL,
			},
			{
				Key:   consoleURLLabelKey,
				Value: runtimeURLsConfig.ConsoleURL,
			},
		}

		gqlClient := gql.NewQueryAssertClient(t, false, newSetExpectedLabelFunc(expectedResponses), expectedSetEventsURLRequest, expectedSetConsoleURLRequest)

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		labels, err := configClient.SetURLsLabels(runtimeURLsConfig)

		// then
		require.NoError(t, err)
		assert.Equal(t, runtimeURLsConfig.EventsURL, labels[eventsURLLabelKey])
		assert.Equal(t, runtimeURLsConfig.ConsoleURL, labels[consoleURLLabelKey])
	})

	t.Run("should return error if setting Console URL as label returned nil response", func(t *testing.T) {
		expectedResponses := []*graphql.Label{
			{
				Key:   eventsURLLabelKey,
				Value: runtimeURLsConfig.EventsURL,
			},
			nil,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, newSetExpectedLabelFunc(expectedResponses), expectedSetEventsURLRequest, expectedSetConsoleURLRequest)

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		labels, err := configClient.SetURLsLabels(runtimeURLsConfig)

		// then
		require.Error(t, err)
		assert.Nil(t, labels)
	})

	t.Run("should return error if setting Console URL as label returned nil response", func(t *testing.T) {
		expectedResponses := []*graphql.Label{nil, nil}

		gqlClient := gql.NewQueryAssertClient(t, false, newSetExpectedLabelFunc(expectedResponses), expectedSetEventsURLRequest, expectedSetConsoleURLRequest)

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		labels, err := configClient.SetURLsLabels(runtimeURLsConfig)

		// then
		require.Error(t, err)
		assert.Nil(t, labels)
	})

	t.Run("should return error if failed to set labels", func(t *testing.T) {
		gqlClient := gql.NewQueryAssertClient(t, true, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*SetRuntimeLabelResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
		}, expectedSetEventsURLRequest, expectedSetConsoleURLRequest)

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		labels, err := configClient.SetURLsLabels(runtimeURLsConfig)

		// then
		require.Error(t, err)
		assert.Nil(t, labels)
	})

}
