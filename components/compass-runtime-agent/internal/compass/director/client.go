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

type GetRuntimeResponse struct {
	Result *graphql.RuntimeExt `json:"result"`
}

type UpdateRuntimeResponse struct {
	Result *graphql.Runtime `json:"result"`
}

//go:generate mockery --name=DirectorClient
type DirectorClient interface {
	FetchConfiguration(ctx context.Context) ([]kymamodel.Application, graphql.Labels, error)
	SetURLsLabels(ctx context.Context, urlsCfg RuntimeURLsConfig, actualLabels graphql.Labels) (graphql.Labels, error)
	SetRuntimeStatusCondition(ctx context.Context, statusCondition graphql.RuntimeStatusCondition) error
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
	graphqlizer   Graphqlizer
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

func (cc *directorClient) UpdateRuntime(ctx context.Context, id string, directorInput *graphql.RuntimeInput) error {

	if directorInput == nil {
		return errors.New("Cannot update runtime in Director: missing Runtime config")
	}

	runtimeInput, err := cc.graphqlizer.RuntimeUpdateInputToGQL(*directorInput)
	if err != nil {
		return err
	}
	runtimeQuery := cc.queryProvider.updateRuntimeMutation(id, runtimeInput)

	req := gcli.NewRequest(runtimeQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

	var response UpdateRuntimeResponse
	err = cc.gqlClient.Do(ctx, req, &response)
	if err != nil {
		return err
	}
	if response.Result == nil {
		return err
	}
	if response.Result.ID != id {
		return err
	}

	return nil
}

func (cc *directorClient) SetRuntimeStatusCondition(ctx context.Context, statusCondition graphql.RuntimeStatusCondition) error {
	// TODO: Set StatusCondition without getting the Runtime
	//       It'll be possible after this issue implementation:
	//       - https://github.com/kyma-incubator/compass/issues/1186
	runtime, err := cc.getRuntime(ctx)
	if err != nil {
		return err
	}
	runtimeInput := &graphql.RuntimeInput{
		Name:            runtime.Name,
		Description:     runtime.Description,
		StatusCondition: &statusCondition,
		Labels:          runtime.Labels,
	}
	err = cc.UpdateRuntime(ctx, cc.runtimeConfig.RuntimeId, runtimeInput)
	if err != nil {
		return err
	}
	return nil
}

func (cc *directorClient) getRuntime(ctx context.Context) (graphql.RuntimeExt, error) {

	runtimeQuery := cc.queryProvider.getRuntimeQuery(cc.runtimeConfig.RuntimeId)

	req := gcli.NewRequest(runtimeQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

	var response GetRuntimeResponse
	err := cc.gqlClient.Do(ctx, req, &response)
	if err != nil {
		return graphql.RuntimeExt{}, err
	}
	if response.Result == nil {
		return graphql.RuntimeExt{}, err
	}
	if response.Result.ID != cc.runtimeConfig.RuntimeId {
		return graphql.RuntimeExt{}, err
	}

	return *response.Result, nil
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
