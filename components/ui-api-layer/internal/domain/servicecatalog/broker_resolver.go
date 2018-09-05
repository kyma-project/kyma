package servicecatalog

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

//TODO: Write unit tests for brokerResolver

type brokerResolver struct {
	brokerLister    brokerListGetter
	brokerConverter gqlBrokerConverter
}

func newBrokerResolver(brokerLister brokerListGetter) *brokerResolver {
	return &brokerResolver{
		brokerLister:    brokerLister,
		brokerConverter: &brokerConverter{},
	}
}

func (r *brokerResolver) ServiceBrokersQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ServiceBroker, error) {
	items, err := r.brokerLister.List(pager.PagingParams{
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

func (r *brokerResolver) ServiceBrokerQuery(ctx context.Context, name string) (*gqlschema.ServiceBroker, error) {
	serviceBroker, err := r.brokerLister.Find(name)
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
