package director

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	gql "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

//go:generate mockery -name=ConfigClient
type ConfigClient interface {
	FetchConfiguration(directorURL, runtimeId string, credentials certificates.ClientCredentials) ([]kymamodel.Application, error)
}

func NewConfigurationClient(gqlClient gql.Client) ConfigClient {
	return &configClient{
		gqlClient: gqlClient,
	}
}

type configClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	runtimeId     string
}

func (cc *configClient) FetchConfiguration(directorURL, runtimeId string, credentials certificates.ClientCredentials) ([]kymamodel.Application, error) {
	//client, err := cc.gqlClientConstructor(credentials.AsTLSCertificate(), directorURL, true, cc.insecureConfigurationFetch)
	//if err != nil {
	//	return nil, errors.Wrap(err, "Failed to create GraphQL client")
	//}
	//
	//applicationPage := ApplicationPage{}
	//response := ApplicationsForRuntimeResponse{Result: &applicationPage}
	//
	//applicationsQuery := cc.queryProvider.applicationsForRuntimeQuery(cc.runtimeId)
	//req := graphql.NewRequest(applicationsQuery)
	//
	//err = client.Do(req, &response)
	//if err != nil {
	//	return nil, errors.Wrap(err, "Failed to fetch Applications")
	//}
	//
	//// TODO: After implementation of paging modify the fetching logic
	//
	//applications := make([]kymamodel.Application, len(applicationPage.Data))
	//for i, app := range applicationPage.Data {
	//	applications[i] = app.ToApplication()
	//}
	//
	//return applications, nil
	return []kymamodel.Application{}, nil
}
