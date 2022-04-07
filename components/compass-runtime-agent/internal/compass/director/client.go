package director

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"
	gql "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	TenantHeader = "Tenant"

	eventsURLLabelKey  = "runtime_eventServiceUrl"
	consoleURLLabelKey = "runtime_consoleUrl"
)

type RuntimeURLsConfig struct {
	EventsURL  string `envconfig:"default=https://gateway.kyma.local"`
	ConsoleURL string `envconfig:"default=https://console.kyma.local"`
}

//go:generate mockery --name=DirectorClient
type DirectorClient interface {
	FetchConfiguration(ctx context.Context) ([]kymamodel.Application, graphql.Labels, error)
	SetURLsLabels(ctx context.Context, urlsCfg RuntimeURLsConfig, actualLabels graphql.Labels) (graphql.Labels, error)
}

func NewConfigurationClient(gqlClient gql.Client, runtimeConfig config.RuntimeConfig) DirectorClient {
	return &directorClient{
		gqlClient:     gqlClient,
		queryProvider: queryProvider{},
		runtimeConfig: runtimeConfig,
	}
}

type directorClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	runtimeConfig config.RuntimeConfig
}

func (cc *directorClient) FetchConfiguration(ctx context.Context) ([]kymamodel.Application, graphql.Labels, error) {
	response := ApplicationsAndLabelsForRuntimeResponse{}

	appsAndLabelsForRuntimeQuery := cc.queryProvider.applicationsAndLabelsForRuntimeQuery(cc.runtimeConfig.RuntimeId)
	req := gcli.NewRequest(appsAndLabelsForRuntimeQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

	err := cc.gqlClient.Do(ctx, req, &response)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to fetch Applications and Labels")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Runtime == nil || response.ApplicationsPage == nil {
		return nil, nil, errors.Errorf("Failed fetch Applications or Labels for Runtime from Director: received nil response.")
	}

	// TODO: After implementation of paging modify the fetching logic

	applications := make([]kymamodel.Application, len(response.ApplicationsPage.Data))
	for i, app := range response.ApplicationsPage.Data {
		applications[i] = app.ToApplication()
	}

	return applications, response.Runtime.Labels, nil
}

func (cc *directorClient) SetURLsLabels(ctx context.Context, urlsCfg RuntimeURLsConfig, currentLabels graphql.Labels) (graphql.Labels, error) {
	targetLabels := map[string]string{
		eventsURLLabelKey:  urlsCfg.EventsURL,
		consoleURLLabelKey: urlsCfg.ConsoleURL,
	}

	updatedLabels := make(map[string]interface{})
	for key, value := range targetLabels {
		if val, ok := currentLabels[key]; !ok || val != value {
			l, err := cc.setURLLabel(ctx, key, value)
			if err != nil {
				return nil, errors.WithMessagef(err, "Failed to set %s Runtime label to value %s", key, val)
			}

			updatedLabels[l.Key] = l.Value
		}
	}

	return updatedLabels, nil
}

func (cc *directorClient) setURLLabel(ctx context.Context, key, value string) (*graphql.Label, error) {
	response := SetRuntimeLabelResponse{}

	setLabelQuery := cc.queryProvider.setRuntimeLabelMutation(cc.runtimeConfig.RuntimeId, key, value)
	req := gcli.NewRequest(setLabelQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

	err := cc.gqlClient.Do(ctx, req, &response)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to set %s Runtime label to value %s", key, value)
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return nil, errors.Errorf("Failed to set %s Runtime label to value %s. Received nil response.", key, value)
	}

	return response.Result, nil
}
