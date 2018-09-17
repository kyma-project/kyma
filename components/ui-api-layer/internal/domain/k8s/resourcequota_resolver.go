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

//go:generate mockery -name=resourceQuotaLister -output=automock -outpkg=automock -case=underscore
type resourceQuotaLister interface {
	ListResourceQuotas(environment string) ([]*v1.ResourceQuota, error)
}

func newResourceQuotaResolver(resourceQuotaLister resourceQuotaLister) *resourceQuotaResolver {
	return &resourceQuotaResolver{
		converter: &resourceQuotaConverter{},
		rqLister:  resourceQuotaLister,
	}
}

type resourceQuotaResolver struct {
	rqLister  resourceQuotaLister
	converter *resourceQuotaConverter
}

func (r *resourceQuotaResolver) ResourceQuotasQuery(ctx context.Context, environment string) ([]gqlschema.ResourceQuota, error) {
	items, err := r.rqLister.ListResourceQuotas(environment)
	if err != nil {
		glog.Error(
			errors.Wrapf(err, "while listing %s [environment: %s]", pretty.ResourceQuotas, environment))
		return nil, gqlerror.New(err, pretty.ResourceQuotas, gqlerror.WithEnvironment(environment))
	}

	return r.converter.ToGQLs(items), nil
}
