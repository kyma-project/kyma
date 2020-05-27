package servicecatalogaddons

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=bindableResourceLister -output=automock -outpkg=automock -case=underscore
type bindableResourceLister interface {
	ListResources(namespace string) ([]*gqlschema.BindableResourcesOutputItem, error)
}

type bindableResourcesResolver struct {
	converter gqlUsageKindConverter
	lister    bindableResourceLister
}

func newBindableResourcesResolver(lister bindableResourceLister) *bindableResourcesResolver {
	return &bindableResourcesResolver{
		lister:    lister,
		converter: &usageKindConverter{},
	}
}

func (rsv *bindableResourcesResolver) ListBindableResources(ctx context.Context, namespace string) ([]*gqlschema.BindableResourcesOutputItem, error) {
	res, err := rsv.lister.ListResources(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for %s in namespace `%s`", pretty.BindableResources, pretty.UsageKind, namespace))
		return nil, gqlerror.New(err, pretty.BindableResources, gqlerror.WithNamespace(namespace))
	}

	return res, nil
}
