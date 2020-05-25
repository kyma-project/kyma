package ui

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/resource"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/gqlerror"
	gqlschema "github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/pkg/errors"
)

type backendModuleResolver struct {
	service   *resource.Service
	converter BackendModuleConverter
}

func newBackendModuleResolver(sf *resource.ServiceFactory) *backendModuleResolver {
	return &backendModuleResolver{
		service:   sf.ForResource(v1alpha1.SchemeGroupVersion.WithResource("backendmodules")),
		converter: BackendModuleConverter{},
	}
}

func (r *backendModuleResolver) BackendModulesQuery(ctx context.Context) ([]*gqlschema.BackendModule, error) {
	items := BackendModuleList{}
	var err error

	err = r.service.List(&items)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.BackendModules))
		return nil, gqlerror.New(err, pretty.BackendModules)
	}

	result, err := r.converter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.BackendModules))
		return nil, gqlerror.New(err, pretty.BackendModules)
	}

	return result, nil
}
