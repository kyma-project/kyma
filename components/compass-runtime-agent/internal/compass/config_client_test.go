package compass

import (
	"crypto/tls"

	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"

	"github.com/stretchr/testify/assert"

	gql "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	"github.com/machinebox/graphql"

	"testing"
)

const (
	tenant    = "tenant"
	runtimeId = "runtimeId"

	directorURL = "https://director.com/graphql"

	expectedQuery = `query {
	result: applicationsForRuntime(runtimeID: runtimeId) {
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
		auth {runtimeID
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

type mockGQLClient struct {
	t               *testing.T
	expectedRequest *graphql.Request
	shouldFail      bool
}

func (c *mockGQLClient) Do(req *graphql.Request, res interface{}) error {
	assert.Equal(c.t, c.expectedRequest, req)

	if !c.shouldFail {
		appForRuntimesResp, ok := res.(*ApplicationsForRuntimeResponse)
		if !ok {
			return errors.New("invalid response type expected")
		}

		appForRuntimesResp.Result.Data = []*Application{
			{Name: "App"},
		}

		return nil
	}

	return errors.New("error")
}

func (c *mockGQLClient) DisableLogging() {}

func newMockClientConstructor(t *testing.T, shouldFail bool) GraphQLClientConstructor {
	return func(certificate tls.Certificate, graphqlEndpoint string, enableLogging bool) (client gql.Client, e error) {
		expectedReq := graphql.NewRequest(expectedQuery)
		expectedReq.Header.Set("Tenant", tenant)

		return &mockGQLClient{
			expectedRequest: expectedReq,
			t:               t,
			shouldFail:      shouldFail,
		}, nil
	}
}

func failingGQLClientConstructor(_ tls.Certificate, _ string, _ bool) (client gql.Client, e error) {
	return nil, errors.New("error")
}

func TestConfigClient_FetchConfiguration(t *testing.T) {

	for _, testCase := range []struct {
		description       string
		expectedApps      []kymamodel.Application
		clientConstructor GraphQLClientConstructor
		shouldFail        bool
	}{
		{
			description: "fetch applications",
			expectedApps: []kymamodel.Application{
				{Name: "App"},
			},
			clientConstructor: newMockClientConstructor(t, false),
			shouldFail:        false,
		},
		{
			description:       "return error when failed to fetch config",
			expectedApps:      nil,
			clientConstructor: newMockClientConstructor(t, true),
			shouldFail:        true,
		},
		{
			description:       "return error when failed to create graphql client",
			expectedApps:      nil,
			clientConstructor: failingGQLClientConstructor,
			shouldFail:        true,
		},
	} {
		t.Run("should "+testCase.description, func(t *testing.T) {
			// given
			configClient := NewConfigurationClient(tenant, runtimeId, testCase.clientConstructor)

			// when
			apps, err := configClient.FetchConfiguration(directorURL, certificates.Credentials{})

			// then
			assert.Equal(t, testCase.expectedApps, apps)

			isError := err != nil
			assert.Equal(t, testCase.shouldFail, isError)
		})
	}

}
