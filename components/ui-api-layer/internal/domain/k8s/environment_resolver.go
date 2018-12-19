package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

//go:generate mockery -name=envLister -output=automock -outpkg=automock -case=underscore
type envLister interface {
	List() ([]gqlschema.Environment, error)
	ListForApplication(appName string) ([]gqlschema.Environment, error)
}

type environmentResolver struct {
	envLister envLister
}

func newEnvironmentResolver(envLister envLister) *environmentResolver {
	return &environmentResolver{
		envLister: envLister,
	}
}

func (r *environmentResolver) EnvironmentsQuery(ctx context.Context, applicationName *string) ([]gqlschema.Environment, error) {
	var err error
	var envs []gqlschema.Environment

	if applicationName == nil {
		envs, err = r.envLister.List()
	} else {
		envs, err = r.envLister.ListForApplication(*applicationName)
	}

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.Environments))
		return nil, gqlerror.New(err, pretty.Environments)
	}

	return envs, nil
}
