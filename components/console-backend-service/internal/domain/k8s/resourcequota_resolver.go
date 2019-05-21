package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=resourceQuotaLister -output=automock -outpkg=automock -case=underscore
type resourceQuotaLister interface {
	ListResourceQuotas(namespace string) ([]*v1.ResourceQuota, error)
	CreateResourceQuota(namespace string, name string, memoryLimits string, memoryRequests string) (*v1.ResourceQuota, error)
}

//go:generate mockery -name=gqlResourceQuotaConverter -output=automock -outpkg=automock -case=underscore
type gqlResourceQuotaConverter interface {
	ToGQLs(in []*v1.ResourceQuota) []gqlschema.ResourceQuota
	ToGQL(in *v1.ResourceQuota) *gqlschema.ResourceQuota
}

func newResourceQuotaResolver(resourceQuotaLister resourceQuotaLister) *resourceQuotaResolver {
	return &resourceQuotaResolver{
		converter: &resourceQuotaConverter{},
		rqLister:  resourceQuotaLister,
	}
}

type resourceQuotaResolver struct {
	rqLister  resourceQuotaLister
	converter gqlResourceQuotaConverter
}

func (r *resourceQuotaResolver) ResourceQuotasQuery(ctx context.Context, namespace string) ([]gqlschema.ResourceQuota, error) {
	items, err := r.rqLister.ListResourceQuotas(namespace)
	if err != nil {
		glog.Error(
			errors.Wrapf(err, "while listing %s [namespace: %s]", pretty.ResourceQuotas, namespace))
		return nil, gqlerror.New(err, pretty.ResourceQuotas, gqlerror.WithNamespace(namespace))
	}

	converted := r.converter.ToGQLs(items)

	return converted, nil
}

func (r *resourceQuotaResolver) CreateResourceQuota(ctx context.Context, namespace string, name string, memoryLimits string, memoryRequests string) (*gqlschema.ResourceQuota, error) {
	item, err := r.rqLister.CreateResourceQuota(namespace, name, memoryLimits, memoryRequests)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s [namespace: %s]", pretty.ResourceQuotas, namespace))
		return nil, gqlerror.New(err, pretty.ResourceQuotas, gqlerror.WithNamespace(namespace))
	}

	converted := r.converter.ToGQL(item)

	return converted, nil
}
