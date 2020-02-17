package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
)

//go:generate mockery -name=gqlKymaVersionConverter -output=automock -outpkg=automock -case=underscore
type gqlKymaVersionConverter interface {
	ToKymaVersion(in *v1.Deployment) gqlschema.VersionInfo
}

type kymaVersionResolver struct {
	deploymentLister     deploymentLister
	kymaVersionConverter gqlKymaVersionConverter
}

func newKymaVersionResolver(deploymentLister deploymentLister) *kymaVersionResolver {
	return &kymaVersionResolver{
		deploymentLister:     deploymentLister,
		kymaVersionConverter: &kymaVersionConverter{},
	}
}

func (r *kymaVersionResolver) KymaVersionQuery(ctx context.Context) (gqlschema.VersionInfo, error) {
	name := "kyma-installer"
	namespace := "kyma-installer"

	deployment, err := r.deploymentLister.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting the %s in namespace `%s`", pretty.Deployment, namespace))
		return gqlschema.VersionInfo{}, gqlerror.New(err, pretty.Deployment, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	if deployment == nil {
		err := errors.Errorf("Deployment with Kyma Version not found")
		glog.Error(err)
		return gqlschema.VersionInfo{}, gqlerror.New(err, pretty.Deployment, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	version := r.kymaVersionConverter.ToKymaVersion(deployment)

	return version, nil
}
