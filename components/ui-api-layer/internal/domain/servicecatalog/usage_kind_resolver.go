package servicecatalog

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

//go:generate mockery -name=usageKindServices -output=automock -outpkg=automock -case=underscore
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
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.UsageKinds))
		return nil, gqlerror.New(err, pretty.UsageKinds)
	}
	return rsv.converter.ToGQLs(res), nil
}

func (rsv *usageKindResolver) ListServiceUsageKindResources(ctx context.Context, usageKind string, environment string) ([]gqlschema.UsageKindResource, error) {
	res, err := rsv.svc.ListUsageKindResources(usageKind, environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for %s `%s` in environment `%s`", pretty.UsageKindResource, pretty.UsageKind, usageKind, environment))
		return nil, gqlerror.New(err, pretty.UsageKindResource, gqlerror.WithEnvironment(environment))
	}

	return res, nil
}
