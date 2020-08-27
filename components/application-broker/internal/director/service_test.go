package director

import (
	"context"
	"testing"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/application-broker/third_party/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_EnsureAPIPackageCredentialsDeletedIfApplicationNotExists(t *testing.T) {
	// given
	cli := NewQGLClient(&emptyGQLClient{})
	svc := NewService(cli, ServiceConfig{})

	// when
	err := svc.EnsureAPIPackageCredentialsDeleted(context.Background(), "aid0", "pid0", "iid0")

	// then
	require.NoError(t, err)
}

func TestMapPackageInstanceAuthToModelOauth(t *testing.T) {
	// given
	svc := NewService(nil, ServiceConfig{})

	oauth := &schema.OAuthCredentialData{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		URL:          "http://token.url",
	}
	fixPkgAuth := fixPackageInstanceAuth(oauth)

	// when
	got, err := svc.mapPackageInstanceAuthToModel(fixPkgAuth)

	// then
	require.NoError(t, err)

	assert.Equal(t, fixPkgAuth.ID, got.ID)
	assert.EqualValues(t, fixPkgAuth.Auth.AdditionalHeaders, got.Config.RequestParameters.Headers)
	assert.EqualValues(t, fixPkgAuth.Auth.AdditionalQueryParams, got.Config.RequestParameters.QueryParameters)
	assert.EqualValues(t, fixPkgAuth.Auth.RequestAuth.Csrf.TokenEndpointURL, got.Config.CSRFConfig.TokenURL)

	assert.Equal(t, oauth.ClientID, got.Config.Credentials.ToCredentials().OAuth.ClientID)
	assert.Equal(t, oauth.ClientSecret, got.Config.Credentials.ToCredentials().OAuth.ClientSecret)
	assert.Equal(t, oauth.URL, got.Config.Credentials.ToCredentials().OAuth.URL)
}

func TestMapPackageInstanceAuthToModelBasic(t *testing.T) {
	// given
	svc := NewService(nil, ServiceConfig{})

	basic := &schema.BasicCredentialData{
		Username: "username",
		Password: "password",
	}

	fixPkgAuth := fixPackageInstanceAuth(basic)

	// when
	got, err := svc.mapPackageInstanceAuthToModel(fixPkgAuth)

	// then
	require.NoError(t, err)

	assert.Equal(t, fixPkgAuth.ID, got.ID)
	assert.EqualValues(t, fixPkgAuth.Auth.AdditionalHeaders, got.Config.RequestParameters.Headers)
	assert.EqualValues(t, fixPkgAuth.Auth.AdditionalQueryParams, got.Config.RequestParameters.QueryParameters)
	assert.EqualValues(t, fixPkgAuth.Auth.RequestAuth.Csrf.TokenEndpointURL, got.Config.CSRFConfig.TokenURL)

	assert.Equal(t, basic.Username, got.Config.Credentials.ToCredentials().BasicAuth.Username)
	assert.Equal(t, basic.Password, got.Config.Credentials.ToCredentials().BasicAuth.Password)
}

func fixPackageInstanceAuth(creds schema.CredentialData) schema.PackageInstanceAuth {
	return schema.PackageInstanceAuth{
		ID: "123",
		Auth: &schema.Auth{
			Credential: creds,
			AdditionalHeaders: &schema.HttpHeaders{
				"headers": []string{"header1", "header2"},
			},
			AdditionalQueryParams: &schema.QueryParams{
				"query-params": []string{"query1", "query2"},
			},
			RequestAuth: &schema.CredentialRequestAuth{
				Csrf: &schema.CSRFTokenCredentialRequestAuth{
					TokenEndpointURL: "http://csrf.auth.token.com",
				},
			},
		},
	}
}

type emptyGQLClient struct {
}

func (c *emptyGQLClient) Run(ctx context.Context, req *graphql.Request, resp interface{}) error {
	return nil
}
