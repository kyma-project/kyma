package compass

import (
	"crypto/tls"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization"
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
	FetchConfiguration(directorURL string, credentials certificates.Credentials) ([]synchronization.Application, error)
}

type GraphQLClientConstructor func(certificate tls.Certificate, graphqlEndpoint string, enableLogging bool) (gql.Client, error)

func NewConfigurationClient(tenant, runtimeId string, gqlClientConstructor GraphQLClientConstructor) ConfigClient {
	return &configClient{
		tenant:               tenant,
		runtimeId:            runtimeId,
		gqlClientConstructor: gqlClientConstructor,
	}
}

type configClient struct {
	tenant               string
	runtimeId            string
	gqlClientConstructor GraphQLClientConstructor
}

// TODO - move to queries

func (cc *configClient) FetchConfiguration(directorURL string, credentials certificates.Credentials) ([]synchronization.Application, error) {
	client, err := cc.gqlClientConstructor(credentials.AsTLSCertificate(), directorURL, true)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	applicationPage := ApplicationPage{}
	response := ApplicationsForRuntimeResponse{Result: &applicationPage}

	// TODO: use proper query when it is ready
	//result: applicationsForRuntime(runtimeId: %s) { ... }

	applicationsQuery := ApplicationsQuery()

	req := graphql.NewRequest(applicationsQuery)
	req.Header.Add(HeaderTenant, cc.tenant)

	err = client.Do(req, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch Applications")
	}

	// TODO: After implementation of paging modify the fetching logic

	applications := make([]synchronization.Application, len(applicationPage.Data))
	for i, app := range applicationPage.Data {
		applications[i] = app.ToApplication()
	}

	return applications, nil
}
