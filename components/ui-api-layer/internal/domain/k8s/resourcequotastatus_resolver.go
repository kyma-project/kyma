package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

//go:generate mockery -name=resourceQuotaStatusChecker -output=automock -outpkg=automock -case=underscore
type resourceQuotaStatusChecker interface {
	CheckResourceQuotaStatus(environment string) (gqlschema.ResourceQuotasStatus, error)
}

func newResourceQuotaStatusResolver(checker resourceQuotaStatusChecker) *resourceQuotaStatusResolver {
	return &resourceQuotaStatusResolver{
		statusChecker: checker,
	}
}

type resourceQuotaStatusResolver struct {
	statusChecker resourceQuotaStatusChecker
}

func (r *resourceQuotaStatusResolver) ResourceQuotasStatus(ctx context.Context, environment string) (gqlschema.ResourceQuotasStatus, error) {
	exceeded, err := r.statusChecker.CheckResourceQuotaStatus(environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s [environment: %s]", pretty.ResourceQuotaStatus, environment))
		return gqlschema.ResourceQuotasStatus{}, gqlerror.New(err, pretty.ResourceQuotaStatus, gqlerror.WithEnvironment(environment))
	}

	return exceeded, nil
}
