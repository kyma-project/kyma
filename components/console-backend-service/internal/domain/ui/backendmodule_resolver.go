package ui

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=backendModuleLister -output=automock -outpkg=automock -case=underscore
type backendModuleLister interface {
	List() ([]*v1alpha1.BackendModule, error)
}

//go:generate mockery -name=gqlBackendModuleConverter -output=automock -outpkg=automock -case=underscore
type gqlBackendModuleConverter interface {
	ToGQL(in *v1alpha1.BackendModule) (*gqlschema.BackendModule, error)
	ToGQLs(in []*v1alpha1.BackendModule) ([]*gqlschema.BackendModule, error)
}

type backendModuleResolver struct {
	backendModuleLister    backendModuleLister
	backendModuleConverter gqlBackendModuleConverter
}

func newBackendModuleResolver(backendModuleLister backendModuleLister) *backendModuleResolver {
	return &backendModuleResolver{
		backendModuleLister:    backendModuleLister,
		backendModuleConverter: &backendModuleConverter{},
	}
}

func (r *backendModuleResolver) BackendModulesQuery(ctx context.Context) ([]*gqlschema.BackendModule, error) {
	var items []*v1alpha1.BackendModule
	var err error

	items, err = r.backendModuleLister.List()

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.BackendModules))
		return nil, gqlerror.New(err, pretty.BackendModules)
	}

	serviceInstances, err := r.backendModuleConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.BackendModules))
		return nil, gqlerror.New(err, pretty.BackendModules)
	}

	return serviceInstances, nil
}
