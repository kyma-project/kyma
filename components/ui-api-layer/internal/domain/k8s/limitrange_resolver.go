package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

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

func (lr *limitRangeResolver) LimitRangesQuery(ctx context.Context, namespace string) ([]gqlschema.LimitRange, error) {
	limitRange, err := lr.lister.List(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "Cannot list %s from `%s` namespace", pretty.LimitRanges, namespace))
		return nil, gqlerror.New(err, pretty.LimitRanges, gqlerror.WithNamespace(namespace))
	}

	return lr.converter.ToGQLs(limitRange), nil
}
