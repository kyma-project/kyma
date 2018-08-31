package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

//go:generate mockery -name=limitRangeLister -output=automock -outpkg=automock -case=underscore
type limitRangeLister interface {
	List(env string) ([]*v1.LimitRange, error)
}

func newLimitRangeResolver(svc *limitRangeService) *limitRangeResolver {
	return &limitRangeResolver{
		lister:    svc,
		converter: limitRangeConverter{},
	}
}

type limitRangeResolver struct {
	converter limitRangeConverter
	lister    limitRangeLister
}

func (lr *limitRangeResolver) LimitRangesQuery(ctx context.Context, env string) ([]gqlschema.LimitRange, error) {
	limitRange, err := lr.lister.List(env)
	if err != nil {
		glog.Error(errors.Wrapf(err, "Cannot list %s from `%s` environment", pretty.LimitRanges, env))
		return nil, gqlerror.New(err, pretty.LimitRanges, gqlerror.WithEnvironment(env))
	}

	return lr.converter.ToGQLs(limitRange), nil
}
