package director

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"
	"github.com/stretchr/testify/require"

	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gql "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"

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
		providerName
		description
		labels
		auths {id}
		packages {data {
		id
		name
		description
		instanceAuthRequestInputSchema
		apiDefinitions {data {
				id
		name
		description
		spec {data
		format
		type}
		targetURL
		group
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
		defaultInstanceAuth {
		credential {
		... on BasicCredentialData {
		username
		password
	}
		... on OAuthCredentialData {
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
		}
		}
		}
		
	}
	pageInfo {startCursor
		endCursor
		hasNextPage}
	totalCount
	}
	
	}
	pageInfo {startCursor
		endCursor
		hasNextPage}
	totalCount
	
		}
	}`

	expectedGetLabelsQuery = `query {
		result: runtime(id: "runtimeId") {
			labels
		}
	}`

	expectedSetEventsURLLabelQuery = `mutation {
		result: setRuntimeLabel(runtimeID: "runtimeId", key: "runtime_eventServiceUrl", value: "https://gateway.kyma.local") {
			key
			value
		}
	}`
	expectedSetConsoleURLLabelQuery = `mutation {
		result: setRuntimeLabel(runtimeID: "runtimeId", key: "runtime_consoleUrl", value: "https://console.kyma.local") {
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

		gqlClient := gql.NewQueryAssertClient(t, false, gql.ResponseMock{
			ModifyResponseFunc: func(t *testing.T, r interface{}) {
				cfg, ok := r.(*ApplicationsForRuntimeResponse)
				require.True(t, ok)
				assert.Empty(t, cfg.Result)
				cfg.Result = expectedResponse
			},
			ExpectedReq: expectedRequest,
		})

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

		gqlClient := gql.NewQueryAssertClient(t, false, gql.ResponseMock{
			ModifyResponseFunc: func(t *testing.T, r interface{}) {
				cfg, ok := r.(*ApplicationsForRuntimeResponse)
				require.True(t, ok)
				assert.Empty(t, cfg.Result)
				cfg.Result = expectedResponse
			},
			ExpectedReq: expectedRequest,
		})

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		applicationsResponse, err := configClient.FetchConfiguration()

		// then
		require.NoError(t, err)
		assert.Empty(t, applicationsResponse)
	})

	t.Run("should return error when result is nil", func(t *testing.T) {
		// given
		gqlClient := gql.NewQueryAssertClient(t, false, gql.ResponseMock{
			ModifyResponseFunc: func(t *testing.T, r interface{}) {
				cfg, ok := r.(*ApplicationsForRuntimeResponse)
				require.True(t, ok)
				assert.Empty(t, cfg.Result)
				cfg.Result = nil
			},
			ExpectedReq: expectedRequest,
		})

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		applicationsResponse, err := configClient.FetchConfiguration()

		// then
		require.Error(t, err)
		assert.Empty(t, applicationsResponse)
	})

	t.Run("should return error when failed to fetch Applications", func(t *testing.T) {
		// given
		gqlClient := gql.NewQueryAssertClient(t, true, gql.ResponseMock{
			ModifyResponseFunc: func(t *testing.T, r interface{}) {
				cfg, ok := r.(*ApplicationsForRuntimeResponse)
				require.True(t, ok)
				assert.Empty(t, cfg.Result)
			},
			ExpectedReq: expectedRequest,
		})

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		applicationsResponse, err := configClient.FetchConfiguration()

		// then
		require.Error(t, err)
		assert.Nil(t, applicationsResponse)
	})
}

func TestConfigClient_ReconcileURLsLabels(t *testing.T) {
	runtimeURLsConfig := RuntimeURLsConfig{
		EventsURL:  "https://gateway.kyma.local",
		ConsoleURL: "https://console.kyma.local",
	}

	expectedGetLabelsRequest := gcli.NewRequest(expectedGetLabelsQuery)
	expectedGetLabelsRequest.Header.Set(TenantHeader, tenant)
	expectedSetEventsURLRequest := gcli.NewRequest(expectedSetEventsURLLabelQuery)
	expectedSetEventsURLRequest.Header.Set(TenantHeader, tenant)
	expectedSetConsoleURLRequest := gcli.NewRequest(expectedSetConsoleURLLabelQuery)
	expectedSetConsoleURLRequest.Header.Set(TenantHeader, tenant)

	newGetExpectedLabelsFunc := func(expectedResponses graphql.Labels) func(t *testing.T, r interface{}) {
		return func(t *testing.T, r interface{}) {
			cfg, ok := r.(*GetRuntimeLabelsResponse)
			assert.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = &Labels{expectedResponses}
		}
	}

	newSetExpectedLabelFunc := func(expectedResponses *graphql.Label) func(t *testing.T, r interface{}) {
		return func(t *testing.T, r interface{}) {
			cfg, ok := r.(*SetRuntimeLabelResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponses
		}
	}

	eventsURLLabel := &graphql.Label{
		Key:   eventsURLLabelKey,
		Value: runtimeURLsConfig.EventsURL,
	}

	consoleURLLabel := &graphql.Label{
		Key:   consoleURLLabelKey,
		Value: runtimeURLsConfig.ConsoleURL,
	}

	t.Run("should set URLs as labels if no labels are set", func(t *testing.T) {
		labelsResponse := graphql.Labels{}

		gqlClient := gql.NewQueryAssertClient(t, false,
			gql.ResponseMock{
				ModifyResponseFunc: newGetExpectedLabelsFunc(labelsResponse),
				ExpectedReq:        expectedGetLabelsRequest,
			}, gql.ResponseMock{
				ModifyResponseFunc: newSetExpectedLabelFunc(eventsURLLabel),
				ExpectedReq:        expectedSetEventsURLRequest,
			}, gql.ResponseMock{
				ModifyResponseFunc: newSetExpectedLabelFunc(consoleURLLabel),
				ExpectedReq:        expectedSetConsoleURLRequest,
			})

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		reconciledLabels, err := configClient.ReconcileLabels(runtimeURLsConfig)

		// then
		require.NoError(t, err)
		assert.Equal(t, 2, len(reconciledLabels))
		assert.Equal(t, runtimeURLsConfig.EventsURL, reconciledLabels[eventsURLLabelKey])
		assert.Equal(t, runtimeURLsConfig.ConsoleURL, reconciledLabels[consoleURLLabelKey])
	})

	t.Run("should not set URLs as labels if there are already set and they're the same", func(t *testing.T) {
		labelsResponse := graphql.Labels{}
		labelsResponse[eventsURLLabelKey] = runtimeURLsConfig.EventsURL
		labelsResponse[consoleURLLabelKey] = runtimeURLsConfig.ConsoleURL

		gqlClient := gql.NewQueryAssertClient(t, false,
			gql.ResponseMock{
				ModifyResponseFunc: newGetExpectedLabelsFunc(labelsResponse),
				ExpectedReq:        expectedGetLabelsRequest,
			})

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		reconciledLabels, err := configClient.ReconcileLabels(runtimeURLsConfig)

		// then
		require.NoError(t, err)
		assert.Equal(t, 0, len(reconciledLabels))
	})

	t.Run("should override URLs if there are already set but are different", func(t *testing.T) {
		labelsResponse := graphql.Labels{}
		labelsResponse[eventsURLLabelKey] = runtimeURLsConfig.EventsURL + " something different"
		labelsResponse[consoleURLLabelKey] = runtimeURLsConfig.ConsoleURL + " something different"

		gqlClient := gql.NewQueryAssertClient(t, false,
			gql.ResponseMock{
				ModifyResponseFunc: newGetExpectedLabelsFunc(labelsResponse),
				ExpectedReq:        expectedGetLabelsRequest,
			}, gql.ResponseMock{
				ModifyResponseFunc: newSetExpectedLabelFunc(eventsURLLabel),
				ExpectedReq:        expectedSetEventsURLRequest,
			}, gql.ResponseMock{
				ModifyResponseFunc: newSetExpectedLabelFunc(consoleURLLabel),
				ExpectedReq:        expectedSetConsoleURLRequest,
			})

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		reconciledLabels, err := configClient.ReconcileLabels(runtimeURLsConfig)

		// then
		require.NoError(t, err)
		assert.Equal(t, 2, len(reconciledLabels))
		assert.Equal(t, runtimeURLsConfig.EventsURL, reconciledLabels[eventsURLLabelKey])
		assert.Equal(t, runtimeURLsConfig.ConsoleURL, reconciledLabels[consoleURLLabelKey])
	})

	t.Run("should set only missing URLs as labels", func(t *testing.T) {
		labelsResponse := graphql.Labels{}
		labelsResponse[eventsURLLabelKey] = runtimeURLsConfig.EventsURL

		gqlClient := gql.NewQueryAssertClient(t, false,
			gql.ResponseMock{
				ModifyResponseFunc: newGetExpectedLabelsFunc(labelsResponse),
				ExpectedReq:        expectedGetLabelsRequest,
			},
			gql.ResponseMock{
				ModifyResponseFunc: newSetExpectedLabelFunc(consoleURLLabel),
				ExpectedReq:        expectedSetConsoleURLRequest,
			})

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		reconciledLabels, err := configClient.ReconcileLabels(runtimeURLsConfig)

		// then
		require.NoError(t, err)
		assert.Equal(t, 1, len(reconciledLabels))
		assert.Equal(t, runtimeURLsConfig.ConsoleURL, reconciledLabels[consoleURLLabelKey])
	})

	t.Run("should return error if getting labels returns nil response", func(t *testing.T) {
		gqlClient := gql.NewQueryAssertClient(t, false,
			gql.ResponseMock{
				ModifyResponseFunc: func(t *testing.T, r interface{}) {
					cfg, ok := r.(*GetRuntimeLabelsResponse)
					assert.True(t, ok)
					assert.Empty(t, cfg.Result)
				},
				ExpectedReq: expectedGetLabelsRequest,
			})

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		labels, err := configClient.ReconcileLabels(runtimeURLsConfig)

		// then
		require.Error(t, err)
		assert.Nil(t, labels)
	})

	t.Run("should return error if setting label returned nil response", func(t *testing.T) {
		labelsResponse := graphql.Labels{}

		gqlClient := gql.NewQueryAssertClient(t, false,
			gql.ResponseMock{
				ModifyResponseFunc: newGetExpectedLabelsFunc(labelsResponse),
				ExpectedReq:        expectedGetLabelsRequest,
			}, gql.ResponseMock{
				ModifyResponseFunc: newSetExpectedLabelFunc(eventsURLLabel),
				ExpectedReq:        expectedSetEventsURLRequest,
			}, gql.ResponseMock{
				ModifyResponseFunc: func(t *testing.T, r interface{}) {
					cfg, ok := r.(*SetRuntimeLabelResponse)
					assert.True(t, ok)
					assert.Empty(t, cfg.Result)
				},
				ExpectedReq: expectedSetConsoleURLRequest,
			})

		configClient := NewConfigurationClient(gqlClient, runtimeConfig)

		// when
		labels, err := configClient.ReconcileLabels(runtimeURLsConfig)

		// then
		require.Error(t, err)
		assert.Nil(t, labels)
	})
}
