package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=resourceQuotaStatusChecker -output=automock -outpkg=automock -case=underscore
type resourceQuotaStatusChecker interface {
	CheckResourceQuotaStatus(namespace string) (*gqlschema.ResourceQuotasStatus, error)
}

func newResourceQuotaStatusResolver(checker resourceQuotaStatusChecker) *resourceQuotaStatusResolver {
	return &resourceQuotaStatusResolver{
		statusChecker: checker,
	}
}

type resourceQuotaStatusResolver struct {
	statusChecker resourceQuotaStatusChecker
}

func (r *resourceQuotaStatusResolver) ResourceQuotasStatus(ctx context.Context, namespace string) (*gqlschema.ResourceQuotasStatus, error) {
	exceeded, err := r.statusChecker.CheckResourceQuotaStatus(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s [namespace: %s]", pretty.ResourceQuotaStatus, namespace))
		return nil, gqlerror.New(err, pretty.ResourceQuotaStatus, gqlerror.WithNamespace(namespace))
	}

	return exceeded, nil
}
