package director

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql/mocks"
	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

const (
	tenant    = "tenant"
	runtimeId = "runtimeId"

	expectedAppsAndLabelsForRuntimeQuery = `query {
		runtime(id: "runtimeId") {
			labels
		}
		applicationsForRuntime(runtimeID: "runtimeId") {
			data {
		id
		name
		providerName
		description
		labels
		auths {id}
		bundles {data {
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

	expectedSetEventsURLLabelQuery = `mutation {
		setRuntimeLabel(runtimeID: "runtimeId", key: "runtime_eventServiceUrl", value: "https://gateway.kyma.local") {
			key
			value
		}
	}`
	expectedSetConsoleURLLabelQuery = `mutation {
		setRuntimeLabel(runtimeID: "runtimeId", key: "runtime_consoleUrl", value: "https://console.kyma.local") {
			key
			value
		}
	}`
)

var runtimeConfig = config.RuntimeConfig{
	RuntimeId: runtimeId,
	Tenant:    tenant,
}

func TestConfigClient_FetchConfiguration(t *testing.T) {
	expectedRequest := gcli.NewRequest(expectedAppsAndLabelsForRuntimeQuery)
	expectedRequest.Header.Set(TenantHeader, tenant)

	setExpectedFetchConfigFunc := func(appsResponse *ApplicationPage, runtimeResponse *Runtime) func(args mock.Arguments) {
		return func(args mock.Arguments) {
			response, ok := args[1].(*ApplicationsAndLabelsForRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, response.ApplicationsPage)
			assert.Empty(t, response.Runtime)
			response.ApplicationsPage = appsResponse
			response.Runtime = runtimeResponse
		}
	}

	t.Run("should fetch configuration", func(t *testing.T) {
		// given
		expectedResponseApplications := &ApplicationPage{
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
					Auths: []*graphql.AppSystemAuth{{"asd", nil}},
				},
			},
			PageInfo:   &graphql.PageInfo{},
			TotalCount: 3,
		}

		expectedResponseRuntime := &Runtime{
			Labels: graphql.Labels{
				eventsURLLabelKey:  "eventsURL",
				consoleURLLabelKey: "consoleURL",
			},
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

		expectedLabels := graphql.Labels{
			eventsURLLabelKey:  "eventsURL",
			consoleURLLabelKey: "consoleURL",
		}

		client := &mocks.Client{}
		client.
			On("Do", expectedRequest, &ApplicationsAndLabelsForRuntimeResponse{}).
			Return(nil).
			Run(setExpectedFetchConfigFunc(expectedResponseApplications, expectedResponseRuntime)).
			Once()

		configClient := NewConfigurationClient(client, runtimeConfig)

		// when
		applicationsResponse, labelsResponse, err := configClient.FetchConfiguration()

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedApps, applicationsResponse)
		assert.Equal(t, expectedLabels, labelsResponse)
	})

	t.Run("should return empty array if no Apps for Runtime", func(t *testing.T) {
		// given
		expectedResponseApps := &ApplicationPage{
			Data:       nil,
			PageInfo:   &graphql.PageInfo{},
			TotalCount: 0,
		}

		expectedResponseRuntime := &Runtime{
			Labels: nil,
		}

		client := &mocks.Client{}
		client.
			On("Do", expectedRequest, &ApplicationsAndLabelsForRuntimeResponse{}).
			Return(nil).
			Run(setExpectedFetchConfigFunc(expectedResponseApps, expectedResponseRuntime)).
			Once()

		configClient := NewConfigurationClient(client, runtimeConfig)

		// when
		applicationsResponse, _, err := configClient.FetchConfiguration()

		// then
		require.NoError(t, err)
		assert.Empty(t, applicationsResponse)
	})

	t.Run("should return empty array if no Labels for Runtime", func(t *testing.T) {
		// given
		expectedResponseRuntime := &Runtime{
			Labels: nil,
		}

		expectedResponseApps := &ApplicationPage{
			Data:       nil,
			PageInfo:   &graphql.PageInfo{},
			TotalCount: 0,
		}

		client := &mocks.Client{}
		client.
			On("Do", expectedRequest, &ApplicationsAndLabelsForRuntimeResponse{}).
			Return(nil).
			Run(setExpectedFetchConfigFunc(expectedResponseApps, expectedResponseRuntime)).
			Once()

		configClient := NewConfigurationClient(client, runtimeConfig)

		// when
		_, labelsResponse, err := configClient.FetchConfiguration()

		// then
		require.NoError(t, err)
		assert.Empty(t, labelsResponse)
	})

	t.Run("should return error when result is nil", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.
			On("Do", expectedRequest, &ApplicationsAndLabelsForRuntimeResponse{}).
			Return(nil).
			Run(setExpectedFetchConfigFunc(nil, nil)).
			Once()

		configClient := NewConfigurationClient(client, runtimeConfig)

		// when
		applicationsResponse, labelsResponse, err := configClient.FetchConfiguration()

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil response")
		assert.Empty(t, labelsResponse)
		assert.Empty(t, applicationsResponse)
	})

	t.Run("should return error when failed to fetch Applications and Labels for Runtime", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.
			On("Do", expectedRequest, &ApplicationsAndLabelsForRuntimeResponse{}).
			Return(errors.New("error")).
			Once()

		configClient := NewConfigurationClient(client, runtimeConfig)

		// when
		applicationsResponse, labelsResponse, err := configClient.FetchConfiguration()

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to fetch Applications and Labels")
		assert.Nil(t, applicationsResponse)
		assert.Nil(t, labelsResponse)
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

	setExpectedRuntimeLabelFunc := func(expectedResponses *graphql.Label) func(args mock.Arguments) {
		return func(args mock.Arguments) {
			response, ok := args[1].(*SetRuntimeLabelResponse)
			require.True(t, ok)
			assert.Empty(t, response.Result)
			response.Result = expectedResponses
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
		currentLabels := graphql.Labels{}

		client := &mocks.Client{}
		client.
			On("Do", expectedSetEventsURLRequest, &SetRuntimeLabelResponse{}).
			Return(nil).
			Run(setExpectedRuntimeLabelFunc(eventsURLLabel)).
			Once()
		client.
			On("Do", expectedSetConsoleURLRequest, &SetRuntimeLabelResponse{}).
			Return(nil).
			Run(setExpectedRuntimeLabelFunc(consoleURLLabel)).
			Once()

		configClient := NewConfigurationClient(client, runtimeConfig)

		// when
		updatedLabels, err := configClient.SetURLsLabels(runtimeURLsConfig, currentLabels)

		// then
		require.NoError(t, err)
		assert.Equal(t, 2, len(updatedLabels))
		assert.Equal(t, runtimeURLsConfig.EventsURL, updatedLabels[eventsURLLabelKey])
		assert.Equal(t, runtimeURLsConfig.ConsoleURL, updatedLabels[consoleURLLabelKey])
	})

	t.Run("should not set URLs as labels if there are already set and they're the same", func(t *testing.T) {
		currentLabels := graphql.Labels{}
		currentLabels[eventsURLLabelKey] = runtimeURLsConfig.EventsURL
		currentLabels[consoleURLLabelKey] = runtimeURLsConfig.ConsoleURL

		configClient := NewConfigurationClient(&mocks.Client{}, runtimeConfig)

		// when
		updatedLabels, err := configClient.SetURLsLabels(runtimeURLsConfig, currentLabels)

		// then
		require.NoError(t, err)
		assert.Equal(t, 0, len(updatedLabels))
	})

	t.Run("should override URLs if there are already set but are different", func(t *testing.T) {
		currentLabels := graphql.Labels{}
		currentLabels[eventsURLLabelKey] = runtimeURLsConfig.EventsURL + " something different"
		currentLabels[consoleURLLabelKey] = runtimeURLsConfig.ConsoleURL + " something different"

		client := &mocks.Client{}
		client.
			On("Do", expectedSetEventsURLRequest, &SetRuntimeLabelResponse{}).
			Return(nil).
			Run(setExpectedRuntimeLabelFunc(eventsURLLabel)).
			Once()
		client.
			On("Do", expectedSetConsoleURLRequest, &SetRuntimeLabelResponse{}).
			Return(nil).
			Run(setExpectedRuntimeLabelFunc(consoleURLLabel)).
			Once()

		configClient := NewConfigurationClient(client, runtimeConfig)

		// when
		updatedLabels, err := configClient.SetURLsLabels(runtimeURLsConfig, currentLabels)

		// then
		require.NoError(t, err)
		assert.Equal(t, 2, len(updatedLabels))
		assert.Equal(t, runtimeURLsConfig.EventsURL, updatedLabels[eventsURLLabelKey])
		assert.Equal(t, runtimeURLsConfig.ConsoleURL, updatedLabels[consoleURLLabelKey])
	})

	t.Run("should set only missing URLs as labels", func(t *testing.T) {
		currentLabels := graphql.Labels{}
		currentLabels[eventsURLLabelKey] = runtimeURLsConfig.EventsURL

		client := &mocks.Client{}
		client.
			On("Do", expectedSetConsoleURLRequest, &SetRuntimeLabelResponse{}).
			Return(nil).
			Run(setExpectedRuntimeLabelFunc(consoleURLLabel)).
			Once()

		configClient := NewConfigurationClient(client, runtimeConfig)

		// when
		updatedLabels, err := configClient.SetURLsLabels(runtimeURLsConfig, currentLabels)

		// then
		require.NoError(t, err)
		assert.Equal(t, 1, len(updatedLabels))
		assert.Equal(t, runtimeURLsConfig.ConsoleURL, updatedLabels[consoleURLLabelKey])
	})

	t.Run("should return error if setting label returned nil response", func(t *testing.T) {
		currentLabels := graphql.Labels{}

		client := &mocks.Client{}
		client.
			On("Do", expectedSetEventsURLRequest, &SetRuntimeLabelResponse{}).
			Return(nil).
			Run(setExpectedRuntimeLabelFunc(eventsURLLabel)).
			Once()

		client.
			On("Do", expectedSetConsoleURLRequest, &SetRuntimeLabelResponse{}).
			Return(nil).
			Run(setExpectedRuntimeLabelFunc(nil)).
			Once()

		configClient := NewConfigurationClient(client, runtimeConfig)

		// when
		updatedLabels, err := configClient.SetURLsLabels(runtimeURLsConfig, currentLabels)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil response")
		assert.Nil(t, updatedLabels)
	})
}
