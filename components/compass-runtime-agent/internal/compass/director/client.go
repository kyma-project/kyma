package director

import (
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
	FetchConfiguration() ([]kymamodel.Application, graphql.Labels, error)
	SetURLsLabels(urlsCfg RuntimeURLsConfig, actualLabels graphql.Labels) (graphql.Labels, error)
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

func (cc *directorClient) FetchConfiguration() ([]kymamodel.Application, graphql.Labels, error) {
	response := ApplicationsAndLabelsForRuntimeResponse{}

	appsAndLabelsForRuntimeQuery := cc.queryProvider.applicationsAndLabelsForRuntimeQuery(cc.runtimeConfig.RuntimeId)
	req := gcli.NewRequest(appsAndLabelsForRuntimeQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

	err := cc.gqlClient.Do(req, &response)
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

func (cc *directorClient) SetURLsLabels(urlsCfg RuntimeURLsConfig, currentLabels graphql.Labels) (graphql.Labels, error) {
	targetLabels := []struct {
		key string
		val string
	}{
		{key: eventsURLLabelKey, val: urlsCfg.EventsURL},
		{key: consoleURLLabelKey, val: urlsCfg.ConsoleURL},
	}

	updatedLabels := make(map[string]interface{})
	for _, tl := range targetLabels {
		if val, ok := currentLabels[tl.key]; !ok || val != tl.val {
			l, err := cc.setURLLabel(tl.key, tl.val)
			if err != nil {
				return nil, errors.WithMessagef(err, "Failed to set %s Runtime label to value %s", tl.key, tl.val)
			}

			updatedLabels[l.Key] = l.Value
		}
	}

	return updatedLabels, nil
}

func (cc *directorClient) setURLLabel(key, value string) (*graphql.Label, error) {
	response := SetRuntimeLabelResponse{}

	setLabelQuery := cc.queryProvider.setRuntimeLabelMutation(cc.runtimeConfig.RuntimeId, key, value)
	req := gcli.NewRequest(setLabelQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

	err := cc.gqlClient.Do(req, &response)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to set %s Runtime label to value %s", key, value)
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return nil, errors.Errorf("Failed to set %s Runtime label to value %s. Received nil response.", key, value)
	}

	return response.Result, nil
}
