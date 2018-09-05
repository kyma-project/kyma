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

func newResourceQuotaStatusResolver(resourceQuotaStatusService *resourceQuotaStatusService) *resourceQuotaStatusResolver {
	return &resourceQuotaStatusResolver{
		statusSvc: resourceQuotaStatusService,
	}
}

type resourceQuotaStatusResolver struct {
	statusSvc *resourceQuotaStatusService
}

func (r *resourceQuotaStatusResolver) ResourceQuotasStatus(ctx context.Context, environment string) (gqlschema.ResourceQuotasStatus, error) {
	resourcesToCheck := []v1.ResourceName{
		v1.ResourceRequestsMemory,
		v1.ResourceLimitsMemory,
		v1.ResourceRequestsCPU,
		v1.ResourceLimitsCPU,
		v1.ResourcePods,
	}
	exceeded, err := r.statusSvc.CheckResourceQuotaStatus(environment, resourcesToCheck)
	if err != nil {
		glog.Error(
			errors.Wrapf(err, "while getting %s [environment: %s]", pretty.ResourceQuotaStatus, environment))
		return gqlschema.ResourceQuotasStatus{}, gqlerror.New(err, pretty.ResourceQuotaStatus, gqlerror.WithEnvironment(environment))
	}

	return exceeded, nil
}
