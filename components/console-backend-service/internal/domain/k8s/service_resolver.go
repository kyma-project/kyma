package k8s

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type serviceResolver struct {
	gqlServiceConverter
	serviceSvc
}

//go:generate mockery -name=serviceSvc -output=automock -outpkg=automock -case=underscore
type serviceSvc interface {
	Find(name, namespace string) (*v1.Service, error)
	List(namespace string, pagingParams pager.PagingParams) ([]*v1.Service, error)
	Update(name, namespace string, update v1.Service) (*v1.Service, error)
	Delete(name, namespace string) error
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=gqlServiceConverter -output=automock -outpkg=automock -case=underscore
type gqlServiceConverter interface {
	ToGQL(in *v1.Service) *gqlschema.Service
	ToGQLs(in []*v1.Service) []gqlschema.Service
}

func newServiceResolver(serviceSvc serviceSvc) *serviceResolver {
	return &serviceResolver{
		serviceSvc:          serviceSvc,
		gqlServiceConverter: &serviceConverter{},
	}
}

func (r *serviceResolver) ServicesQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.Service, error) {
	services, err := r.serviceSvc.List(namespace, pager.PagingParams{
		First:  first,
		Offset: offset,
	})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s from namespace %s", pretty.Services, namespace))
		return nil, gqlerror.New(err, pretty.Services, gqlerror.WithNamespace(namespace))
	}
	converted := r.gqlServiceConverter.ToGQLs(services)
	return converted, nil
}

func (r *serviceResolver) ServiceQuery(ctx context.Context, name string, namespace string) (*gqlschema.Service, error) {
	service, err := r.serviceSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s from namespace %s", pretty.Service, name, namespace))
		return nil, gqlerror.New(err, pretty.Service, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	converted := r.gqlServiceConverter.ToGQL(service)
	return converted, nil
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
