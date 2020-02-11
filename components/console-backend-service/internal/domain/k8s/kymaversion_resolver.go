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

type kymaVersionResolver struct {
	kymaVersionSvc    kymaVersionSvc
	kymaVersionConverter *kymaVersionConverter
}

func newKymaVersionResolver(kymaVersionSvc kymaVersionSvc) *kymaVersionResolver {
	return &kymaVersionResolver{
		kymaVersionSvc: kymaVersionSvc,
	}
}

func (r *kymaVersionResolver) KymaVersionQuery(ctx context.Context) (string, error) {
	namespace := "kyma-installer"

	deployment, err := r.kymaVersionSvc.FindDeployment()
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting the %s in namespace `%s`", pretty.Deployment, namespace))
		return "", gqlerror.New(err, pretty.Deployment, gqlerror.WithNamespace(namespace))
	}

	deploymentImage := deployment.Spec.Template.Spec.Containers[0].Image
	version := r.kymaVersionConverter.ToKymaVersion(deploymentImage)

	return version, nil
}