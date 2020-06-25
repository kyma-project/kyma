package servicecatalogaddons

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=usageKindServices -output=automock -outpkg=automock -case=underscore
type usageKindServices interface {
	List(params pager.PagingParams) ([]*v1alpha1.UsageKind, error)
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

func (rsv *usageKindResolver) ListUsageKinds(ctx context.Context, first *int, offset *int) ([]*gqlschema.UsageKind, error) {
	res, err := rsv.svc.List(pager.PagingParams{First: first, Offset: offset})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.UsageKinds))
		return nil, gqlerror.New(err, pretty.UsageKinds)
	}
	return rsv.converter.ToGQLs(res), nil
}
