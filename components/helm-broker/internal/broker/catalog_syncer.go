package broker

import (
	"context"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type catalogSyncerService struct {
	underlying catalogGetter
}

func newCatalogSyncerService(underlying catalogGetter) *catalogSyncerService {
	return &catalogSyncerService{
		underlying: underlying,
	}
}

// GetCatalog triggers syncer and execute underlying catalogService
func (cs *catalogSyncerService) GetCatalog(ctx context.Context, osbCtx OsbContext) (*osb.CatalogResponse, error) {
	return cs.underlying.GetCatalog(ctx, osbCtx)
}
