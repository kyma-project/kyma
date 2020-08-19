package servicecatalog

import (
	"context"

	"github.com/golang/glog"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlClusterServiceBrokerConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterServiceBrokerConverter interface {
	ToGQL(in *v1beta1.ClusterServiceBroker) (*gqlschema.ClusterServiceBroker, error)
	ToGQLs(in []*v1beta1.ClusterServiceBroker) ([]*gqlschema.ClusterServiceBroker, error)
}

//go:generate mockery -name=clusterServiceBrokerSvc -output=automock -outpkg=automock -case=underscore
type clusterServiceBrokerSvc interface {
	Find(name string) (*v1beta1.ClusterServiceBroker, error)
	List(pagingParams pager.PagingParams) ([]*v1beta1.ClusterServiceBroker, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type clusterServiceBrokerResolver struct {
	clusterServiceBrokerSvc clusterServiceBrokerSvc
	brokerConverter         gqlClusterServiceBrokerConverter
}

func newClusterServiceBrokerResolver(clusterServiceBrokerSvc clusterServiceBrokerSvc) *clusterServiceBrokerResolver {
	return &clusterServiceBrokerResolver{
		clusterServiceBrokerSvc: clusterServiceBrokerSvc,
		brokerConverter:         &clusterServiceBrokerConverter{},
	}
}

func (r *clusterServiceBrokerResolver) ClusterServiceBrokersQuery(ctx context.Context, first *int, offset *int) ([]*gqlschema.ClusterServiceBroker, error) {
	items, err := r.clusterServiceBrokerSvc.List(pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ClusterServiceBrokers))
		return nil, gqlerror.New(err, pretty.ClusterServiceBrokers)
	}

	serviceBrokers, err := r.brokerConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ClusterServiceBrokers))
		return nil, gqlerror.New(err, pretty.ClusterServiceBrokers)
	}

	return serviceBrokers, nil
}

func (r *clusterServiceBrokerResolver) ClusterServiceBrokerQuery(ctx context.Context, name string) (*gqlschema.ClusterServiceBroker, error) {
	serviceBroker, err := r.clusterServiceBrokerSvc.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s", pretty.ClusterServiceBroker))
		return nil, gqlerror.New(err, pretty.ClusterServiceBroker, gqlerror.WithName(name))
	}
	if serviceBroker == nil {
		return nil, nil
	}

	result, err := r.brokerConverter.ToGQL(serviceBroker)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting to %s type", pretty.ClusterServiceBroker))
		return nil, gqlerror.New(err, pretty.ClusterServiceBroker, gqlerror.WithName(name))
	}

	return result, nil
}

func (r *clusterServiceBrokerResolver) ClusterServiceBrokerEventSubscription(ctx context.Context) (<-chan *gqlschema.ClusterServiceBrokerEvent, error) {
	channel := make(chan *gqlschema.ClusterServiceBrokerEvent, 1)
	filter := func(entity *v1beta1.ClusterServiceBroker) bool {
		return true
	}

	brokerListener := listener.NewClusterServiceBroker(channel, filter, r.brokerConverter)

	r.clusterServiceBrokerSvc.Subscribe(brokerListener)
	go func() {
		defer close(channel)
		defer r.clusterServiceBrokerSvc.Unsubscribe(brokerListener)
		<-ctx.Done()
	}()

	return channel, nil
}
