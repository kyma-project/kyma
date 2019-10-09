package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"kyma-project.io/compass-runtime-agent/internal/config"
	gql "kyma-project.io/compass-runtime-agent/internal/graphql"
	kymamodel "kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

const (
	TenantHeader = "Tenant"

	eventsURLLabelKey  = "runtime/event_service_url"
	consoleURLLabelKey = "runtime/console_url"
)

type RuntimeURLsConfig struct {
	EventsURL  string `envconfig:"default=https://gateway.kyma.local"`
	ConsoleURL string `envconfig:"default=https://console.kyma.local"`
}

//go:generate mockery -name=ConfigClient
type ConfigClient interface {
	FetchConfiguration() ([]kymamodel.Application, error)
	SetURLsLabels(urlsCfg RuntimeURLsConfig) (graphql.Labels, error)
}

func NewConfigurationClient(gqlClient gql.Client, runtimeConfig config.RuntimeConfig) ConfigClient {
	return &configClient{
		gqlClient:     gqlClient,
		queryProvider: queryProvider{},
		runtimeConfig: runtimeConfig,
	}
}

type configClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	runtimeConfig config.RuntimeConfig
}

func (cc *configClient) FetchConfiguration() ([]kymamodel.Application, error) {
	response := ApplicationsForRuntimeResponse{
		Result: &ApplicationPage{},
	}

	applicationsQuery := cc.queryProvider.applicationsForRuntimeQuery(cc.runtimeConfig.RuntimeId)
	req := gcli.NewRequest(applicationsQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

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

func (cc *configClient) SetURLsLabels(urlsCfg RuntimeURLsConfig) (graphql.Labels, error) {
	eventsURLLabel, err := cc.setURLLabel(eventsURLLabelKey, urlsCfg.EventsURL)
	if err != nil {
		return nil, err
	}

	consoleURLLabel, err := cc.setURLLabel(consoleURLLabelKey, urlsCfg.ConsoleURL)
	if err != nil {
		return nil, err
	}

	return graphql.Labels{
		eventsURLLabel.Key:  eventsURLLabel.Value,
		consoleURLLabel.Key: consoleURLLabel.Value,
	}, nil
}

func (cc *configClient) setURLLabel(key, value string) (*graphql.Label, error) {
	response := SetRuntimeLabelResponse{}

	setLabelQuery := cc.queryProvider.setRuntimeLabelMutation(cc.runtimeConfig.RuntimeId, key, value)
	req := gcli.NewRequest(setLabelQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

	err := cc.gqlClient.Do(req, &response)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to set %s Runtime label to value %s", key, value)
	}

	if response.Result == nil {
		return nil, errors.Errorf("Failed to set %s Runtime label to value %s. Received nil response.", key, value)
	}

	return response.Result, nil
}
