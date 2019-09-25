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

	directorURL = "https://director.com/graphql"

	expectedQuery = `query {
	result: applicationsForRuntime(runtimeID: "runtimeId") {
		data {
		id
		name
		description
		labels
		apis {data {
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
		eventAPIs {data {
		
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
	
	}
	pageInfo {startCursor
		endCursor
		hasNextPage}
	totalCount
	
	}
}`
)

func TestConfigClient_FetchConfiguration(t *testing.T) {

	expectedRequest := gcli.NewRequest(expectedQuery)
	expectedRequest.Header.Set(TenantHeader, tenant)

	runtimeConfig := config.RuntimeConfig{
		RuntimeId: runtimeId,
		Tenant:    tenant,
	}

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
			},
			PageInfo:   &graphql.PageInfo{},
			TotalCount: 2,
		}

		expectedApps := []kymamodel.Application{
			{
				Name: "App1",
				ID:   "abcd-efgh",
			},
			{
				ID:   "ijkl-mnop",
				Name: "App2",
			},
		}

		gqlClient := gql.NewQueryAssertClient(t, expectedRequest, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*ApplicationsForRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg)
			cfg.Result = expectedResponse
		})

		configClient := NewConfigurationClient(gqlClient)

		// when
		applicationsResponse, err := configClient.FetchConfiguration(directorURL, runtimeConfig)

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

		gqlClient := gql.NewQueryAssertClient(t, expectedRequest, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*ApplicationsForRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg)
			cfg.Result = expectedResponse
		})

		configClient := NewConfigurationClient(gqlClient)

		// when
		applicationsResponse, err := configClient.FetchConfiguration(directorURL, runtimeConfig)

		// then
		require.NoError(t, err)
		assert.Empty(t, applicationsResponse)
	})

	t.Run("should return error when failed to fetch Applications", func(t *testing.T) {
		// given
		gqlClient := gql.NewQueryAssertClient(t, expectedRequest, true, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*ApplicationsForRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg)
		})

		configClient := NewConfigurationClient(gqlClient)

		// when
		applicationsResponse, err := configClient.FetchConfiguration(directorURL, runtimeConfig)

		// then
		require.Error(t, err)
		assert.Nil(t, applicationsResponse)
	})
}
