package director

import (
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"kyma-project.io/compass-runtime-agent/internal/config"
	gql "kyma-project.io/compass-runtime-agent/internal/graphql"
	kymamodel "kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

const (
	TenantHeader = "Tenant"
)

//go:generate mockery -name=ConfigClient
type ConfigClient interface {
	FetchConfiguration(directorURL string, runtimeConfig config.RuntimeConfig) ([]kymamodel.Application, error)
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

func (cc *configClient) FetchConfiguration(directorURL string, runtimeConfig config.RuntimeConfig) ([]kymamodel.Application, error) {
	response := ApplicationsForRuntimeResponse{
		Result: &ApplicationPage{},
	}

	applicationsQuery := cc.queryProvider.applicationsForRuntimeQuery(runtimeConfig.RuntimeId)
	req := graphql.NewRequest(applicationsQuery)
	req.Header.Set(TenantHeader, runtimeConfig.Tenant)

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
