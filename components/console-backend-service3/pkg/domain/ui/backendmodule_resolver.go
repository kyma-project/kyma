package ui

import (
	"context"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/gqlerror"
	model "github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/resource"
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

func (r *backendModuleResolver) BackendModulesQuery(ctx context.Context) ([]*model.BackendModule, error) {
	items := BackendModuleList{}
	var err error

	err = r.service.List(&items)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.BackendModules))
		return nil, gqlerror.New(err, pretty.BackendModules)
	}

	result := r.converter.ToGQLs(items)

	return result, nil
}

func (r *backendModuleResolver) BackendModulesEvents(ctx context.Context) (<-chan *model.BackendModuleEvent, error) {
	channel := make(chan *model.BackendModuleEvent, 1)

	listener := resource.NewListener(NewBackendModuleEventHandler(channel))
	r.service.AddListener(listener)
	go func() {
		defer close(channel)
		defer r.service.DeleteListener(listener)
		<-ctx.Done()
	}()

	return channel, nil
}