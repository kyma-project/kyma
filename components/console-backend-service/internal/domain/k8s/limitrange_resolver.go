package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
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

func (lr *limitRangeResolver) CreateLimitRange(ctx context.Context, namespace string, name string, limitRange gqlschema.LimitRangeInput) (*gqlschema.LimitRange, error) {
	createdLimitRange, err := lr.lister.Create(namespace, name, limitRange)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s` in namespace `%s`", pretty.LimitRange, name, namespace))
		return nil, gqlerror.New(err, pretty.LimitRange, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return lr.converter.ToGQL(createdLimitRange), err
}
