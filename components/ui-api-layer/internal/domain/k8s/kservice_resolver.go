package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type kserviceResolver struct {
	gqlKServiceConverter
	kserviceSvc
}

//go:generate mockery -name=kserviceSvc -output=automock -outpkg=automock -case=underscore
type kserviceSvc interface {
	Find(name, namespace string) (*v1.Service, error)
	List(namespace string, pagingParams pager.PagingParams) ([]*v1.Service, error)
	Update(name, namespace string, update v1.Service) (*v1.Service, error)
	Delete(name, namespace string) error
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=gqlKServiceConverter -output=automock -outpkg=automock -case=underscore
type gqlKServiceConverter interface {
	ToGQL(in *v1.Service) *gqlschema.KService
	ToGQLs(in []*v1.Service) []gqlschema.KService
}

func newKserviceResolver(kserviceSvc kserviceSvc) *kserviceResolver {
	return &kserviceResolver{
		kserviceSvc:          kserviceSvc,
		gqlKServiceConverter: &kserviceConverter{},
	}
}

func (r *kserviceResolver) KServicesQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.KService, error) {
	services, err := r.kserviceSvc.List(namespace, pager.PagingParams{
		First:  first,
		Offset: offset,
	})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s from namespace %s", pretty.Services, namespace))
		return nil, gqlerror.New(err, pretty.Services, gqlerror.WithNamespace(namespace))
	}
	converted := r.gqlKServiceConverter.ToGQLs(services)
	return converted, nil
}

func (r *kserviceResolver) KServiceQuery(ctx context.Context, name string, namespace string) (*gqlschema.KService, error) {
	service, err := r.kserviceSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s from namespace %s", pretty.Service, name, namespace))
		return nil, gqlerror.New(err, pretty.Service, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	converted := r.gqlKServiceConverter.ToGQL(service)
	return converted, nil
}

func (r *kserviceResolver) ServiceEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceEvent, error) {

	channel := make(chan gqlschema.ServiceEvent, 1)
	filter := func(service *v1.Service) bool {
		return service != nil && service.Namespace == namespace
	}

	serviceListener := listener.NewService(channel, filter, r.gqlKServiceConverter)

	r.kserviceSvc.Subscribe(serviceListener)
	go func() {
		defer close(channel)
		defer r.kserviceSvc.Unsubscribe(serviceListener)
		<-ctx.Done()
	}()

	return channel, nil
}
