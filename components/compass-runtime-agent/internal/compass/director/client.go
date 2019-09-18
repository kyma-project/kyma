package director

import (
	"crypto/tls"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"

	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/machinebox/graphql"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	gql "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ConfigClient
type ConfigClient interface {
	FetchConfiguration(directorURL, runtimeId string, credentials certificates.ClientCredentials) ([]kymamodel.Application, error)
}

type GraphQLClientConstructor func(certificate tls.Certificate, graphqlEndpoint string, enableLogging bool, insecureConfigFetch bool) (gql.Client, error)

func NewConfigurationClient(gqlClientConstructor GraphQLClientConstructor, insecureConfigFetch bool) ConfigClient {
	return &configClient{
		gqlClientConstructor:       gqlClientConstructor,
		insecureConfigurationFetch: insecureConfigFetch,
	}
}

type configClient struct {
	gqlClientConstructor       GraphQLClientConstructor
	insecureConfigurationFetch bool
	queryProvider              queryProvider
	runtimeId                  string
}

func (cc *configClient) FetchConfiguration(directorURL, runtimeId string, credentials certificates.ClientCredentials) ([]kymamodel.Application, error) {
	client, err := cc.gqlClientConstructor(credentials.AsTLSCertificate(), directorURL, true, cc.insecureConfigurationFetch)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	applicationPage := compass.ApplicationPage{}
	response := compass.ApplicationsForRuntimeResponse{Result: &applicationPage}

	// TODO: will this query stay the same? Meaning will the Id be required or will it be determined based on certificate?
	applicationsQuery := cc.queryProvider.applicationsForRuntimeQuery(cc.runtimeId)
	req := graphql.NewRequest(applicationsQuery)

	err = client.Do(req, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch Applications")
	}

	// TODO: After implementation of paging modify the fetching logic

	applications := make([]kymamodel.Application, len(applicationPage.Data))
	for i, app := range applicationPage.Data {
		applications[i] = app.ToApplication()
	}

	return applications, nil
}
