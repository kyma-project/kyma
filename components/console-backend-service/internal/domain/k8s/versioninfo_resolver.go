package k8s

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/apps/v1"
)

//go:generate mockery -name=gqlVersionInfoConverter -output=automock -outpkg=automock -case=underscore
type gqlVersionInfoConverter interface {
	ToGQL(in *v1.Deployment) *gqlschema.VersionInfo
}

type versionInfoResolver struct {
	deploymentLister     deploymentLister
	versionInfoConverter gqlVersionInfoConverter
}

func newVersionInfoResolver(deploymentLister deploymentLister) *versionInfoResolver {
	return &versionInfoResolver{
		deploymentLister:     deploymentLister,
		versionInfoConverter: &versionInfoConverter{},
	}
}

func (r *versionInfoResolver) VersionInfoQuery(ctx context.Context) (*gqlschema.VersionInfo, error) {
	name := "kyma-installer"
	namespace := "kyma-installer"

	deployment, err := r.deploymentLister.Find(name, namespace)
	if err != nil || deployment == nil {
		return nil, nil
	}

	version := r.versionInfoConverter.ToGQL(deployment)

	return version, nil
}
