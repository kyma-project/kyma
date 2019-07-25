package compass

import (
	"crypto/tls"
	"fmt"

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
const applicationQueryFormat = `query {
	result: applications {
		%s
	}
}`

func (cc *configClient) FetchConfiguration(directorURL string, credentials certificates.Credentials) ([]synchronization.Application, error) {
	client, err := cc.gqlClientConstructor(credentials.AsTLSCertificate(), directorURL, true)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	gqlFieldsProvider := gqlFieldsProvider{}

	applicationPage := ApplicationPage{}

	response := ApplicationsForRuntimeResponse{Result: &applicationPage}

	// TODO: use proper query when it is ready
	//result: applicationsForRuntime(runtimeId: %s) { ... }

	applicationsQuery := fmt.Sprintf(applicationQueryFormat, gqlFieldsProvider.Page(gqlFieldsProvider.ForApplication()))

	//fmt.Println(applicationsQuery)

	req := graphql.NewRequest(applicationsQuery)
	req.Header.Add(HeaderTenant, cc.tenant)

	err = client.Do(req, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch Applications")
	}

	fmt.Println("Applications Count: ", len(applicationPage.Data))
	fmt.Println("Page Info start: ", applicationPage.PageInfo.StartCursor)
	fmt.Println("Page Info end: ", applicationPage.PageInfo.EndCursor)
	fmt.Println("Page Info has next: ", applicationPage.PageInfo.HasNextPage)

	for _, app := range applicationPage.Data {
		fmt.Println("App ", app.ID, " apis count: ", len(app.APIs.Data))
	}

	// TODO - fetch all API pages and EventAPI pages for this Application

	// TODO - repeat for all ApplicationPages

	return nil, nil
}
