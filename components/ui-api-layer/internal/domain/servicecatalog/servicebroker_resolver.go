package servicecatalog

import (
	"context"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/resource"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlServiceBrokerConverter -output=automock -outpkg=automock -case=underscore
type gqlServiceBrokerConverter interface {
	ToGQL(in *v1beta1.ServiceBroker) (*gqlschema.ServiceBroker, error)
	ToGQLs(in []*v1beta1.ServiceBroker) ([]gqlschema.ServiceBroker, error)
}

//go:generate mockery -name=serviceBrokerSvc -output=automock -outpkg=automock -case=underscore
type serviceBrokerSvc interface {
	Find(name, environment string) (*v1beta1.ServiceBroker, error)
	List(environment string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceBroker, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type serviceBrokerResolver struct {
	serviceBrokerSvc serviceBrokerSvc
	brokerConverter  gqlServiceBrokerConverter
}

func newServiceBrokerResolver(serviceBrokerSvc serviceBrokerSvc) *serviceBrokerResolver {
	return &serviceBrokerResolver{
		serviceBrokerSvc: serviceBrokerSvc,
		brokerConverter:  &serviceBrokerConverter{},
	}
}

func (r *serviceBrokerResolver) ServiceBrokersQuery(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceBroker, error) {
	items, err := r.serviceBrokerSvc.List(environment, pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ServiceBrokers))
		return nil, gqlerror.New(err, pretty.ServiceBrokers)
	}

	serviceBrokers, err := r.brokerConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceBrokers))
		return nil, gqlerror.New(err, pretty.ServiceBrokers)
	}

	return serviceBrokers, nil
}

func (r *serviceBrokerResolver) ServiceBrokerQuery(ctx context.Context, name string, environment string) (*gqlschema.ServiceBroker, error) {
	serviceBroker, err := r.serviceBrokerSvc.Find(name, environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s", pretty.ServiceBroker))
		return nil, gqlerror.New(err, pretty.ServiceBroker, gqlerror.WithName(name))
	}
	if serviceBroker == nil {
		return nil, nil
	}

	result, err := r.brokerConverter.ToGQL(serviceBroker)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting to %s type", pretty.ServiceBroker))
		return nil, gqlerror.New(err, pretty.ServiceBroker, gqlerror.WithName(name))
	}

	return result, nil
}

func (r *serviceBrokerResolver) ServiceBrokerEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceBrokerEvent, error) {
	channel := make(chan gqlschema.ServiceBrokerEvent, 1)
	filter := func(entity *v1beta1.ServiceBroker) bool {
		return entity != nil && entity.Namespace == environment
	}

	instanceListener := listener.NewServiceBroker(channel, filter, r.brokerConverter)

	r.serviceBrokerSvc.Subscribe(instanceListener)
	go func() {
		defer close(channel)
		defer r.serviceBrokerSvc.Unsubscribe(instanceListener)
		<-ctx.Done()
	}()

	return channel, nil
}
