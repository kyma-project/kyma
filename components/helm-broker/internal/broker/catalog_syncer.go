package broker

import (
	"context"

	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

//go:generate mockery -name=syncer -output=automock -outpkg=automock -case=underscore
type syncer interface {
	Execute() error
}

type catalogSyncerService struct {
	underlying catalogGetter
	syncer     syncer
}

func newCatalogSyncerService(underlying catalogGetter, syncer syncer) *catalogSyncerService {
	return &catalogSyncerService{
		syncer:     syncer,
		underlying: underlying,
	}
}

// GetCatalog triggers syncer and execute underlying catalogService
func (cs *catalogSyncerService) GetCatalog(ctx context.Context, osbCtx OsbContext) (*osb.CatalogResponse, error) {
	// Trigger sync with bundle repositories
	// If the sync is too long, the Service Catalog call timeout is exceeded and the Service Catalog
	// will try again
	if err := cs.syncer.Execute(); err != nil {
		return nil, errors.Wrap(err, "while executing bundles sync")
	}
	return cs.underlying.GetCatalog(ctx, osbCtx)
}
