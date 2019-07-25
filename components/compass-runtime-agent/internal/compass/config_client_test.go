package compass

import (
	"crypto/tls"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	gql "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	"github.com/machinebox/graphql"

	"testing"
)

const (
	tenant    = "tenant"
	runtimeId = "runtimeId"

	directorURL = "https://director.com/graphql"
)

type mockGQLClient struct {
	t               *testing.T
	expectedRequest *graphql.Request
	//requestAssert   func(t *testing.T, req *graphql.Request)
}

func (c *mockGQLClient) Do(req *graphql.Request, res interface{}) error {
	assert.Equal(c.t, c.expectedRequest, req)

	appForRuntimesResp, ok := res.(*ApplicationsForRuntimeResponse)
	require.True(c.t, ok)

	appForRuntimesResp.Result.Data = []*Application{
		{
			Name: "App",
		},
	}

	return nil
}

func (c *mockGQLClient) DisableLogging() {
	panic("implement me")
}

func newMockClientConstructor(t *testing.T,
	//assertRequestFunc func(t *testing.T, req *graphql.Request)
	expectedRequest *graphql.Request,
) GraphQLClientConstructor {
	return func(certificate tls.Certificate, graphqlEndpoint string, enableLogging bool) (client gql.Client, e error) {
		return &mockGQLClient{
			t: t,
			//requestAssert: assertRequestFunc,
			expectedRequest: expectedRequest,
		}, nil
	}
}

const (
	expectedQuery = `query {
	result: applications {
		data {
		id
		name
		description
		labels
		status {condition timestamp}
		webhooks {id
		applicationID
		type
		url
		auth {
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
		healthCheckURL
		apis {data {
				id
		name
		description
		spec {data
		format
		type
		fetchRequest {url
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
		}
		mode
		filter
		status {condition timestamp}}}
		targetURL
		group
		auths {runtimeID
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
		format
		fetchRequest {url
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
		}
		mode
		filter
		status {condition timestamp}}}
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
		fetchRequest {url
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
		}
		mode
		filter
		status {condition timestamp}}
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

	//for _, testCase := range []struct {
	//	description         string
	//	expectedRequestFunc func() *graphql.Request
	//	//expectedApps        []*Application // TODO - use it instead of asserts after implementation
	//}{
	//	{
	//		description: "fetch applications",
	//		expectedRequestFunc: func() *graphql.Request {
	//			expectedReq := graphql.NewRequest(expectedQuery)
	//			expectedReq.Header.Set("Tenant", tenant)
	//			return expectedReq
	//		},
	//		//expectedApps: []*Application{},
	//	},
	//} {
	//	t.Run("should "+testCase.description, func(t *testing.T) {
	//		mockClientConstructor := newMockClientConstructor(t, testCase.expectedRequestFunc())
	//
	//		configClient := NewConfigurationClient(tenant, runtimeId, mockClientConstructor)
	//
	//		// when
	//		config, err := configClient.FetchConfiguration(directorURL, certificates.Credentials{})
	//
	//		// then
	//		require.NoError(t, err)
	//
	//		// TODO - assert after implementation
	//		//assert.Equal(t, "App", config[0].Name)
	//		assert.Nil(t, config)
	//	})
	//}

	//t.Run("should fetch applications", func(t *testing.T) {
	//	// given
	//
	//	expectedReq := graphql.NewRequest(expectedQuery)
	//	expectedReq.Header.Set("Tenant", tenant)
	//
	//	mockClientConstructor := newMockClientConstructor(t, expectedReq) //	func(t *testing.T, req *graphql.Request) {
	//	//	// TODO - test specific asserts
	//	//	expectedReq := graphql.NewRequest(expectedQuery) // TODO - pass only request?
	//	//	expectedReq.Header.Set("Tenant", tenant)
	//	//
	//	//	assert.EqualValues(t, expectedReq, req)
	//	//	assert.Equal(t, tenant, req.Header.Get("Tenant"))
	//	//}
	//
	//	configClient := NewConfigurationClient(tenant, runtimeId, mockClientConstructor)
	//
	//	// when
	//	config, err := configClient.FetchConfiguration(directorURL, certificates.Credentials{})
	//
	//	// then
	//	require.NoError(t, err)
	//
	//	// TODO - assert after implementation
	//	//assert.Equal(t, "App", config[0].Name)
	//	assert.Nil(t, config)
	//})

}
