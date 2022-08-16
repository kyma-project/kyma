package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/oauth"
	gql "github.com/machinebox/graphql"
)

const (
	AuthorizationHeader = "Authorization"
	TenantHeader        = "Tenant"
)

//go:generate mockery -name=DirectorClient
type DirectorClient interface {
	RegisterApplication(config *gqlschema.RuntimeInput, tenant string) (string, error)
	UnregisterApplication(id, tenant string) (graphql.RuntimeExt, error)
	RequestOneTimeTokenForApplication() error
	RuntimeExists(gardenerClusterName, tenant string) (bool, error)
}

type directorClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer.Graphqlizer
	token         oauth.Token
	oauthClient   oauth.Client
}

func NewDirectorClient(gqlClient gql.Client, oauthClient oauth.Client) DirectorClient {
	return &directorClient{
		gqlClient:     gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer.Graphqlizer{},
		token:         oauth.Token{},
	}
}
