package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
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
	externalErr := errors.New("Cannot query ServiceBrokers")

	items, err := r.brokerLister.List(pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrap(err, "while listing ServiceBrokers"))
		return nil, externalErr
	}

	serviceBrokers, err := r.brokerConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting ServiceBrokers"))
		return nil, externalErr
	}

	return serviceBrokers, nil
}

func (r *brokerResolver) ServiceBrokerQuery(ctx context.Context, name string) (*gqlschema.ServiceBroker, error) {
	externalErr := fmt.Errorf("Cannot query ServiceBroker with name `%s`", name)

	serviceBroker, err := r.brokerLister.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceBroker"))
		return nil, externalErr
	}
	if serviceBroker == nil {
		return nil, nil
	}

	result, err := r.brokerConverter.ToGQL(serviceBroker)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting to ServiceBroker type"))
		return nil, externalErr
	}

	return result, nil
}
