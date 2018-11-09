package broker_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker/automock"
	"github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCatalogSyncerSuccess(t *testing.T) {
	// GIVEN
	osbCtx := *broker.NewOSBContext("id", "2.3")
	underlying := &automock.CatalogGetter{}
	underlying.On("GetCatalog", mock.Anything, osbCtx).Return(fixCatalogResponse(), nil).Once()
	defer underlying.AssertExpectations(t)

	syncer := &automock.Syncer{}
	syncer.On("Execute").Once()
	defer syncer.AssertExpectations(t)

	svc := broker.NewCatalogSyncerService(underlying, syncer)

	// WHEN
	resp, err := svc.GetCatalog(context.Background(), osbCtx)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixCatalogResponse(), resp)

}

func fixCatalogResponse() *v2.CatalogResponse {
	return &v2.CatalogResponse{
		Services: []v2.Service{
			{ID: "id-1"},
		},
	}
}
