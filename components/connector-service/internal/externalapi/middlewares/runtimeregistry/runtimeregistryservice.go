package runtimeregistry

import (
	"fmt"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/graphql"
	"os"
	"time"
)

const (
	timeout  = 30 * time.Second
	query    = `{ "query": "mutation { updateRuntimeState(identifier: \"%s\", state: \"%s\") { status } }" }`
	filename = "config.json"

	EstablishedState = "Established"
)

type RuntimeRegistryService interface {
	ReportState(state RuntimeState) error
}

type RuntimeState struct {
	Identifier string
	State      string
}

type runtimeRegistryService struct {
	graphQL        graphql.GraphQLService
	configFilePath string
}

func NewRuntimeRegistryService(graphQL graphql.GraphQLService, configFilePath string) RuntimeRegistryService {
	return &runtimeRegistryService{graphQL: graphQL, configFilePath: configFilePath}
}

func (rrs runtimeRegistryService) ReportState(state RuntimeState) error {
	file, e := os.Open(rrs.configFilePath + filename)

	if e != nil {
		return apperrors.Internal("Error while reading config file: %s", e)
	}

	config, e := rrs.graphQL.ReadConfig(file)

	if e != nil {
		return e
	}

	query := prepareQuery(state)

	response, e := rrs.graphQL.SendRequest(query, config, timeout)

	if e != nil {
		return e
	}

	statusCode := response.StatusCode

	if statusCode != 200 {
		return apperrors.Internal("Unexpected status code during runtime State update: %d", statusCode)
	}

	return nil
}

func prepareQuery(runtimeState RuntimeState) string {
	return fmt.Sprintf(query, runtimeState.Identifier, runtimeState.State)
}
