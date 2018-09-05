package kubeless

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/kubeless/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

type functionResolver struct {
	functionLister    functionLister
	functionConverter *functionConverter
}

func newFunctionResolver(functionLister functionLister) *functionResolver {
	return &functionResolver{
		functionLister:    functionLister,
		functionConverter: &functionConverter{},
	}
}

func (r *functionResolver) FunctionsQuery(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.Function, error) {
	functions, err := r.functionLister.List(environment, pager.PagingParams{
		First:  first,
		Offset: offset,
	})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for environment %s", pretty.Functions, environment))
		return nil, gqlerror.New(err, pretty.Functions, gqlerror.WithEnvironment(environment))
	}

	return r.functionConverter.ToGQLs(functions), nil
}
