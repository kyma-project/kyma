package k8s

import (
	"context"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
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
		return nil, errors.Wrapf(err, "cannot get limit range from ns: %s", env)
	}

	return lr.converter.ToGQLs(limitRange), nil
}
