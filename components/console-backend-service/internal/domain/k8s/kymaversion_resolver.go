package k8s

import (
	"context"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
)

//go:generate mockery -name=kymaVersionSvc -output=automock -outpkg=automock -case=underscore
type kymaVersionSvc interface {
	FindDeployment() (*v1.Deployment, error)
}

//go:generate mockery -name=gqlKymaVersionConverter -output=automock -outpkg=automock -case=underscore
type gqlKymaVersionConverter interface {
	ToKymaVersion(in *v1.Deployment) string
}

type kymaVersionResolver struct {
	kymaVersionSvc    kymaVersionSvc
	kymaVersionConverter gqlKymaVersionConverter
}

func newKymaVersionResolver(kymaVersionSvc kymaVersionSvc) *kymaVersionResolver {
	return &kymaVersionResolver{
		kymaVersionSvc: kymaVersionSvc,
		kymaVersionConverter: &kymaVersionConverter{},
	}
}

func (r *kymaVersionResolver) KymaVersionQuery(ctx context.Context) (string, error) {
	namespace := "kyma-installer"

	deployment, err := r.kymaVersionSvc.FindDeployment()
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting the %s in namespace `%s`", pretty.Deployment, namespace))
		return "", gqlerror.New(err, pretty.Deployment, gqlerror.WithNamespace(namespace))
	}

	if deployment == nil {
		glog.Error(errors.Errorf("`%s` not found in namespace `%s`", pretty.Deployment, namespace))
		return "", gqlerror.New(err, pretty.Deployment, gqlerror.WithNamespace(namespace))
	}

	version := r.kymaVersionConverter.ToKymaVersion(deployment)

	return version, nil
}