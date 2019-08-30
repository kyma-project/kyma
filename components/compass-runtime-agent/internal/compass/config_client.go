package compass

import (
	"crypto/tls"

	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/machinebox/graphql"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	gql "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	"github.com/pkg/errors"
)

const (
	HeaderTenant = "Tenant"
)

//go:generate mockery -name=ConfigClient
type ConfigClient interface {
	FetchConfiguration(directorURL string, credentials certificates.Credentials) ([]kymamodel.Application, error)
}

type GraphQLClientConstructor func(certificate tls.Certificate, graphqlEndpoint string, enableLogging bool, insecureConfigFetch bool) (gql.Client, error)

func NewConfigurationClient(tenant, runtimeId string, gqlClientConstructor GraphQLClientConstructor, insecureConfigFetch bool) ConfigClient {
	return &configClient{
		tenant:                     tenant,
		runtimeId:                  runtimeId,
		gqlClientConstructor:       gqlClientConstructor,
		insecureConfigurationFetch: insecureConfigFetch,
	}
}

type configClient struct {
	tenant                     string
	runtimeId                  string
	gqlClientConstructor       GraphQLClientConstructor
	insecureConfigurationFetch bool
}

func (cc *configClient) FetchConfiguration(directorURL string, credentials certificates.Credentials) ([]kymamodel.Application, error) {
	client, err := cc.gqlClientConstructor(credentials.AsTLSCertificate(), directorURL, true, cc.insecureConfigurationFetch)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	applicationPage := ApplicationPage{}
	response := ApplicationsForRuntimeResponse{Result: &applicationPage}

	applicationsQuery := ApplicationsForRuntimeQuery(cc.runtimeId)
	req := graphql.NewRequest(applicationsQuery)
	req.Header.Add(HeaderTenant, cc.tenant)

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
