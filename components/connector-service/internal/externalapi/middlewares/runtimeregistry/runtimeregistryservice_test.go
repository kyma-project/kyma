package runtimeregistry

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/graphql"
	"github.com/kyma-project/kyma/components/connector-service/internal/graphql/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

func TestRuntimeRegistryService_ReportState(t *testing.T) {
	t.Run("should send updated state", func(t *testing.T) {
		//given
		service := &mocks.GraphQLService{}
		registryService := NewRuntimeRegistryService(service)

		response := &http.Response{StatusCode: http.StatusOK}

		service.On("ReadConfig", mock.Anything).Return(graphql.Config{}, nil)
		service.On("SendRequest", mock.Anything, mock.Anything, mock.Anything).Return(response, nil)

		state := RuntimeState{
			identifier: "ff0aad1d-c602-446d-8852-bc35508d0d52",
			state:      EstablishedState,
		}

		//when
		e := registryService.ReportState(state, "testdata/")

		//then
		assert.NoError(t, e)
		service.AssertNumberOfCalls(t, "SendRequest", 1)
	})
}
