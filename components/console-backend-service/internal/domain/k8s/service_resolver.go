package k8s

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	v1 "k8s.io/api/core/v1"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type serviceResolver struct {
	gqlServiceConverter
	serviceSvc
}

//go:generate mockery -name=serviceSvc -output=automock -outpkg=automock -case=underscore
type serviceSvc interface {
	Find(name, namespace string) (*v1.Service, error)
	List(namespace string, excludedLabels []string, pagingParams pager.PagingParams) ([]*v1.Service, error)
	Update(name, namespace string, update v1.Service) (*v1.Service, error)
	Delete(name, namespace string) error
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=gqlServiceConverter -output=automock -outpkg=automock -case=underscore
type gqlServiceConverter interface {
	ToGQL(in *v1.Service) (*gqlschema.Service, error)
	ToGQLs(in []*v1.Service) ([]gqlschema.Service, error)
	GQLJSONToService(in gqlschema.JSON) (v1.Service, error)
}

func newServiceResolver(serviceSvc serviceSvc) *serviceResolver {
	return &serviceResolver{
		serviceSvc:          serviceSvc,
		gqlServiceConverter: &serviceConverter{},
	}
}

func (r *serviceResolver) ServicesQuery(ctx context.Context, namespace string, excludedLabels []string, first *int, offset *int) ([]gqlschema.Service, error) {
	services, err := r.serviceSvc.List(namespace, excludedLabels, pager.PagingParams{
		First:  first,
		Offset: offset,
	})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s from namespace %s", pretty.Services, namespace))
		return nil, gqlerror.New(err, pretty.Services, gqlerror.WithNamespace(namespace))
	}
	return r.gqlServiceConverter.ToGQLs(services)
}

func (r *serviceResolver) ServiceQuery(ctx context.Context, name string, namespace string) (*gqlschema.Service, error) {
	service, err := r.serviceSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s from namespace %s", pretty.Service, name, namespace))
		return nil, gqlerror.New(err, pretty.Service, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	return r.gqlServiceConverter.ToGQL(service)
}

func (r *serviceResolver) ServiceEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceEvent, error) {
	channel := make(chan gqlschema.ServiceEvent, 1)
	filter := func(service *v1.Service) bool {
		return service != nil && service.Namespace == namespace
	}

	serviceListener := listener.NewService(channel, filter, r.gqlServiceConverter)

	r.serviceSvc.Subscribe(serviceListener)
	go func() {
		defer close(channel)
		defer r.serviceSvc.Unsubscribe(serviceListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (r *serviceResolver) UpdateServiceMutation(ctx context.Context, name string, namespace string, update gqlschema.JSON) (*gqlschema.Service, error) {
	service, err := r.gqlServiceConverter.GQLJSONToService(update)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace `%s`", pretty.Service, name, namespace))
		return nil, gqlerror.New(err, pretty.Service, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	updated, err := r.serviceSvc.Update(name, namespace, service)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace %s", pretty.Service, name, namespace))
		return nil, gqlerror.New(err, pretty.Service, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return r.gqlServiceConverter.ToGQL(updated)
}

func (r *serviceResolver) DeleteServiceMutation(context context.Context, name string, namespace string) (*gqlschema.Service, error) {
	service, err := r.serviceSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s `%s` in namespace `%s`", pretty.Service, name, namespace))
		return nil, gqlerror.New(err, pretty.Service, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	serviceCopy := service.DeepCopy()
	deletedService, err := r.gqlServiceConverter.ToGQL(serviceCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Service))
		return nil, gqlerror.New(err, pretty.Service, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	err = r.serviceSvc.Delete(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from namespace `%s`", pretty.Service, name, namespace))
		return nil, gqlerror.New(err, pretty.Service, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return deletedService, nil
}
