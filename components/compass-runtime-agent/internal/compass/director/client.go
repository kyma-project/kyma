package director

import (
	gql "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ConfigClient
type ConfigClient interface {
	FetchConfiguration(directorURL, runtimeId string) ([]kymamodel.Application, error)
}

func NewConfigurationClient(gqlClient gql.Client) ConfigClient {
	return &configClient{
		gqlClient: gqlClient,
	}
}

type configClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
}

func (cc *configClient) FetchConfiguration(directorURL, runtimeId string) ([]kymamodel.Application, error) {
	response := ApplicationsForRuntimeResponse{}

	applicationsQuery := cc.queryProvider.applicationsForRuntimeQuery(runtimeId)
	req := graphql.NewRequest(applicationsQuery)

	err := cc.gqlClient.Do(req, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch Applications")
	}

	// TODO: After implementation of paging modify the fetching logic

	applications := make([]kymamodel.Application, len(response.Result.Data))
	for i, app := range response.Result.Data {
		applications[i] = app.ToApplication()
	}

	return applications, nil
}
