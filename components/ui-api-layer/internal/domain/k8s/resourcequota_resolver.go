package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

//go:generate mockery -name=resourceQuotaLister -output=automock -outpkg=automock -case=underscore
type resourceQuotaLister interface {
	ListResourceQuotas(environment string) ([]*v1.ResourceQuota, error)
}

func newResourceQuotaResolver(resourceQuotaLister resourceQuotaLister, resourceQuotaStatusService *resourceQuotaStatusService) *resourceQuotaResolver {
	return &resourceQuotaResolver{
		converter: &resourceQuotaConverter{},
		rqLister:  resourceQuotaLister,
		rqSvc:     resourceQuotaStatusService,
	}
}

type resourceQuotaResolver struct {
	rqSvc     *resourceQuotaStatusService
	rqLister  resourceQuotaLister
	converter *resourceQuotaConverter
}

func (r *resourceQuotaResolver) ResourceQuotasQuery(ctx context.Context, environment string) ([]gqlschema.ResourceQuota, error) {
	items, err := r.rqLister.ListResourceQuotas(environment)
	if err != nil {
		glog.Error(
			errors.Wrapf(err, "while listing resource quotas [environment: %s]", environment))
		return nil, errors.New("cannot get resource quotas")
	}

	return r.converter.ToGQLs(items), nil
}

func (r *resourceQuotaResolver) ResourceQuotaStatus(ctx context.Context, environment string) (gqlschema.ResourceQuotaStatus, error) {
	resourcesToCheck := []v1.ResourceName{
		v1.ResourceRequestsMemory,
		v1.ResourceLimitsMemory,
		v1.ResourceRequestsCPU,
		v1.ResourceLimitsCPU,
		v1.ResourcePods,
	}
	exceeded, err := r.rqSvc.CheckResourceQuotaStatus(environment, resourcesToCheck)
	if err != nil {
		glog.Error(
			errors.Wrapf(err, "while getting ResourceQuotaStatus [environment: %s]", environment))
		return gqlschema.ResourceQuotaStatus{}, errors.New("cannot get resource quota status")
	}

	return exceeded, nil
}
