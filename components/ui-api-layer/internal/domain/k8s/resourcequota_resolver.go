package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

func newResourceQuotaResolver(resourceQuotaLister resourceQuotaLister) *resourceQuotaResolver {
	return &resourceQuotaResolver{
		converter: &resourceQuotaConverter{},
		lister:    resourceQuotaLister,
	}
}

type resourceQuotaResolver struct {
	lister    resourceQuotaLister
	converter *resourceQuotaConverter
}

func (r *resourceQuotaResolver) ResourceQuotasQuery(ctx context.Context, environment string) ([]gqlschema.ResourceQuota, error) {
	items, err := r.lister.List(environment)
	if err != nil {
		glog.Error(
			errors.Wrapf(err, "while listing resource quotas [environment: %s]", environment))
		return nil, errors.New("cannot get resource quotas")
	}

	return r.converter.ToGQLs(items), nil
}
