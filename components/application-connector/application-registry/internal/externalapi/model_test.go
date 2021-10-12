package externalapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

func TestServiceDetailsToServiceDefinition(t *testing.T) {

	documentation := &Documentation{
		DisplayName: "doc",
		Description: "doc description",
		Type:        "docu",
		Tags:        []string{"tag1", "tag2"},
		Docs: []DocsObject{{
			Title:  "docu title",
			Type:   "docu type",
			Source: "docu source",
		}},
	}

	marshalledDocs, err := json.Marshal(documentation)
	require.NoError(t, err)

	for _, testCase := range []struct {
		description               string
		serviceDetails            ServiceDetails
		expectedServiceDefinition model.ServiceDefinition
	}{
		{
			description: "should convert full service details with Oauth to service definition",
			serviceDetails: ServiceDetails{
				Provider:         "provider",
				Name:             "service",
				Description:      "description",
				ShortDescription: "short description",
				Identifier:       "id1",
				Labels:           &map[string]string{"app": "test"},
				Api: &API{
					TargetUrl: "http://test",
					Credentials: &CredentialsWithCSRF{
						OauthWithCSRF: &OauthWithCSRF{
							Oauth: Oauth{
								URL:          "test",
								ClientID:     "testId",
								ClientSecret: "testSecret",
							},
							CSRFInfo: &CSRFInfo{TokenEndpointURL: "token/endpoint"},
						},
					},
					Spec:             []byte("spec"),
					SpecificationUrl: "spec-url",
					ApiType:          "OData",
					RequestParameters: &RequestParameters{
						Headers:         &map[string][]string{"TestHeader": {"header1", "header2"}},
						QueryParameters: &map[string][]string{"TestQuery": {"query1", "query2"}},
					},
					SpecificationCredentials: &Credentials{
						Oauth: &Oauth{
							URL:          "test-spec-cred",
							ClientID:     "testId-spec-cred",
							ClientSecret: "testSecret-spec-cred",
						},
					},
					SpecificationRequestParameters: &RequestParameters{
						Headers:         &map[string][]string{"TestSpecHeader": {"specHeader1", "specHeader2"}},
						QueryParameters: &map[string][]string{"TestSpecQuery": {"specQuery1", "specQuery2"}},
					},
				},
				Events:        &Events{Spec: []byte("events spec")},
				Documentation: documentation,
			},
			expectedServiceDefinition: model.ServiceDefinition{
				Name:             "service",
				Identifier:       "id1",
				Provider:         "provider",
				Description:      "description",
				ShortDescription: "short description",
				Labels:           &map[string]string{"app": "test"},
				Api: &model.API{
					TargetUrl: "http://test",
					Credentials: &model.CredentialsWithCSRF{
						Oauth: &model.Oauth{
							URL:          "test",
							ClientID:     "testId",
							ClientSecret: "testSecret",
						},
						CSRFInfo: &model.CSRFInfo{TokenEndpointURL: "token/endpoint"},
					},
					Spec:             []byte("spec"),
					SpecificationUrl: "spec-url",
					ApiType:          "OData",
					RequestParameters: &model.RequestParameters{
						Headers:         &map[string][]string{"TestHeader": {"header1", "header2"}},
						QueryParameters: &map[string][]string{"TestQuery": {"query1", "query2"}},
					},
					SpecificationCredentials: &model.Credentials{
						Oauth: &model.Oauth{
							URL:          "test-spec-cred",
							ClientID:     "testId-spec-cred",
							ClientSecret: "testSecret-spec-cred",
						},
					},
					SpecificationRequestParameters: &model.RequestParameters{
						Headers:         &map[string][]string{"TestSpecHeader": {"specHeader1", "specHeader2"}},
						QueryParameters: &map[string][]string{"TestSpecQuery": {"specQuery1", "specQuery2"}},
					},
				},
				Events:        &model.Events{Spec: []byte("events spec")},
				Documentation: marshalledDocs,
			},
		},
		{
			description: "should convert service details with Basic Auth to service definition",
			serviceDetails: ServiceDetails{
				Provider:    "provider",
				Name:        "service",
				Description: "description",
				Api: &API{
					TargetUrl: "http://test",
					Credentials: &CredentialsWithCSRF{
						BasicWithCSRF: &BasicAuthWithCSRF{
							BasicAuth: BasicAuth{
								Username: "testuser",
								Password: "testpassword",
							},
							CSRFInfo: &CSRFInfo{TokenEndpointURL: "token/endpoint"},
						},
					},
					SpecificationUrl: "spec-url",
					SpecificationCredentials: &Credentials{
						Basic: &BasicAuth{
							Username: "testuser-spec",
							Password: "testpassword-spec",
						},
					},
				},
			},
			expectedServiceDefinition: model.ServiceDefinition{
				Name:        "service",
				Provider:    "provider",
				Description: "description",
				Api: &model.API{
					TargetUrl: "http://test",
					Credentials: &model.CredentialsWithCSRF{
						Basic: &model.Basic{
							Username: "testuser",
							Password: "testpassword",
						},
						CSRFInfo: &model.CSRFInfo{TokenEndpointURL: "token/endpoint"},
					},
					SpecificationUrl: "spec-url",
					SpecificationCredentials: &model.Credentials{
						Basic: &model.Basic{
							Username: "testuser-spec",
							Password: "testpassword-spec",
						},
					},
				},
			},
		},
		{
			description: "should convert service details with Headers and Query Params provided in old way",
			serviceDetails: ServiceDetails{
				Provider:    "provider",
				Name:        "service",
				Description: "description",
				Api: &API{
					TargetUrl:       "http://test",
					Headers:         &map[string][]string{"TestHeader": {"header1", "header2"}},
					QueryParameters: &map[string][]string{"TestQuery": {"query1", "query2"}},
				},
			},
			expectedServiceDefinition: model.ServiceDefinition{
				Name:        "service",
				Provider:    "provider",
				Description: "description",
				Api: &model.API{
					TargetUrl: "http://test",
					RequestParameters: &model.RequestParameters{
						Headers:         &map[string][]string{"TestHeader": {"header1", "header2"}},
						QueryParameters: &map[string][]string{"TestQuery": {"query1", "query2"}},
					},
				},
			},
		},
		{
			description: "should use requestParameters if both requestParameters and headers provided",
			serviceDetails: ServiceDetails{
				Provider:    "provider",
				Name:        "service",
				Description: "description",
				Api: &API{
					TargetUrl: "http://test",
					RequestParameters: &RequestParameters{
						Headers:         &map[string][]string{"TestHeader": {"header1", "header2"}},
						QueryParameters: &map[string][]string{"TestQuery": {"query1", "query2"}},
					},
					Headers:         &map[string][]string{"DifferentHeader": {"h1", "h2"}},
					QueryParameters: &map[string][]string{"DifferentQuery": {"q1", "q2"}},
				},
			},
			expectedServiceDefinition: model.ServiceDefinition{
				Name:        "service",
				Provider:    "provider",
				Description: "description",
				Api: &model.API{
					TargetUrl: "http://test",
					RequestParameters: &model.RequestParameters{
						Headers:         &map[string][]string{"TestHeader": {"header1", "header2"}},
						QueryParameters: &map[string][]string{"TestQuery": {"query1", "query2"}},
					},
				},
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			serviceDeff, err := serviceDetailsToServiceDefinition(testCase.serviceDetails)
			require.NoError(t, err)

			assert.EqualValues(t, testCase.expectedServiceDefinition, serviceDeff)
		})
	}

}

func TestServiceDefinitionToServiceDetails(t *testing.T) {

	documentation := &Documentation{
		DisplayName: "doc",
		Description: "doc description",
		Type:        "docu",
		Tags:        []string{"tag1", "tag2"},
		Docs: []DocsObject{{
			Title:  "docu title",
			Type:   "docu type",
			Source: "docu source",
		}},
	}

	marshalledDocs, err := json.Marshal(documentation)
	require.NoError(t, err)

	for _, testCase := range []struct {
		description            string
		serviceDefinition      model.ServiceDefinition
		expectedServiceDetails ServiceDetails
	}{
		{
			description: "should convert full service definition with Oauth to service details",
			serviceDefinition: model.ServiceDefinition{
				Name:             "service",
				Identifier:       "id1",
				Provider:         "provider",
				Description:      "description",
				ShortDescription: "short description",
				Labels:           &map[string]string{"app": "test"},
				Api: &model.API{
					TargetUrl: "http://test",
					Credentials: &model.CredentialsWithCSRF{
						Oauth: &model.Oauth{
							URL:          "test",
							ClientID:     "testId",
							ClientSecret: "testSecret",
						},
						CSRFInfo: &model.CSRFInfo{TokenEndpointURL: "token/endpoint"},
					},
					Spec:             []byte("spec"),
					SpecificationUrl: "spec-url",
					ApiType:          "OData",
					RequestParameters: &model.RequestParameters{
						Headers:         &map[string][]string{"TestHeader": {"header1", "header2"}},
						QueryParameters: &map[string][]string{"TestQuery": {"query1", "query2"}},
					},
				},
				Events:        &model.Events{Spec: []byte("events spec")},
				Documentation: marshalledDocs,
			},
			expectedServiceDetails: ServiceDetails{
				Provider:         "provider",
				Name:             "service",
				Description:      "description",
				ShortDescription: "short description",
				Identifier:       "id1",
				Labels:           &map[string]string{"app": "test"},
				Api: &API{
					TargetUrl: "http://test",
					Credentials: &CredentialsWithCSRF{
						OauthWithCSRF: &OauthWithCSRF{
							Oauth: Oauth{
								URL:          "test",
								ClientID:     "********",
								ClientSecret: "********",
							},
							CSRFInfo: &CSRFInfo{TokenEndpointURL: "token/endpoint"},
						},
					},
					Spec:             []byte("spec"),
					SpecificationUrl: "spec-url",
					ApiType:          "OData",
					RequestParameters: &RequestParameters{
						Headers:         &map[string][]string{"TestHeader": {"header1", "header2"}},
						QueryParameters: &map[string][]string{"TestQuery": {"query1", "query2"}},
					},
					Headers:         &map[string][]string{"TestHeader": {"header1", "header2"}},
					QueryParameters: &map[string][]string{"TestQuery": {"query1", "query2"}},
				},
				Events:        &Events{Spec: []byte("events spec")},
				Documentation: documentation,
			},
		}, {
			description: "should convert service definition with Basic Auth to service details",
			serviceDefinition: model.ServiceDefinition{
				Name:        "service",
				Provider:    "provider",
				Description: "description",
				Api: &model.API{
					TargetUrl: "http://test",
					Credentials: &model.CredentialsWithCSRF{
						Basic: &model.Basic{
							Username: "username",
							Password: "password",
						},
						CSRFInfo: &model.CSRFInfo{TokenEndpointURL: "token/endpoint"},
					},
				},
			},
			expectedServiceDetails: ServiceDetails{
				Provider:    "provider",
				Name:        "service",
				Description: "description",
				Api: &API{
					TargetUrl: "http://test",
					Credentials: &CredentialsWithCSRF{
						BasicWithCSRF: &BasicAuthWithCSRF{
							BasicAuth: BasicAuth{
								Username: "********",
								Password: "********",
							},
							CSRFInfo: &CSRFInfo{TokenEndpointURL: "token/endpoint"},
						},
					},
				},
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			serviceDetails, err := serviceDefinitionToServiceDetails(testCase.serviceDefinition)
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedServiceDetails, serviceDetails)
		})
	}

}
