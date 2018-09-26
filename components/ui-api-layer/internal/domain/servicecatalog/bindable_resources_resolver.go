package servicecatalog

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

//go:generate mockery -name=bindableResourceLister -output=automock -outpkg=automock -case=underscore
type bindableResourceLister interface {
	ListResources(environment string) ([]gqlschema.BindableResourcesOutputItem, error)
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

func (rsv *bindableResourcesResolver) ListBindableResources(ctx context.Context, environment string) ([]gqlschema.BindableResourcesOutputItem, error) {
	res, err := rsv.lister.ListResources(environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for %s in environment `%s`", pretty.BindableResources, pretty.UsageKind, environment))
		return nil, gqlerror.New(err, pretty.BindableResources, gqlerror.WithEnvironment(environment))
	}

	return res, nil
}
