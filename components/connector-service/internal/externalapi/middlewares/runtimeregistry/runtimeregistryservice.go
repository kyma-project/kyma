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
	ReportState(state RuntimeState, configFilePath string) error
}

type RuntimeState struct {
	identifier string
	state      string
}

type runtimeRegistryService struct {
	graphQL graphql.GraphQLService
}

func NewRuntimeRegistryService(graphQL graphql.GraphQLService) RuntimeRegistryService {
	return &runtimeRegistryService{graphQL: graphQL}
}

func (rrs runtimeRegistryService) ReportState(state RuntimeState, configFilePath string) error {
	file, e := os.Open(configFilePath + filename)

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
		return apperrors.Internal("Unexpected status code during runtime state update: %d", statusCode)
	}

	return nil
}

func prepareQuery(runtimeState RuntimeState) string {
	return fmt.Sprintf(query, runtimeState.identifier, runtimeState.state)
}
