package servicecatalog

import (
	"context"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
)

type usageKindServices interface {
	List(params pager.PagingParams) ([]*v1alpha1.UsageKind, error)
	ListUsageKindResources(usageKind string, environment string) ([]gqlschema.UsageKindResource, error)
}

type usageKindResolver struct {
	converter gqlUsageKindConverter
	svc       usageKindServices
}

func newUsageKindResolver(svc usageKindServices) *usageKindResolver {
	return &usageKindResolver{
		svc:       svc,
		converter: &usageKindConverter{},
	}
}

func (rsv *usageKindResolver) ListUsageKinds(ctx context.Context, first *int, offset *int) ([]gqlschema.UsageKind, error) {
	res, err := rsv.svc.List(pager.PagingParams{First: first, Offset: offset})
	if err != nil {
		return nil, errors.Wrap(err, "while listing ServiceBindingTargets")
	}
	return rsv.converter.ToGQLs(res), nil
}

func (rsv *usageKindResolver) ListServiceUsageKindResources(ctx context.Context, usageKind string, environment string) ([]gqlschema.UsageKindResource, error) {
	return rsv.svc.ListUsageKindResources(usageKind, environment)
}
