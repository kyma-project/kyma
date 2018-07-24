package kubeless

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
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
	externalErr := fmt.Errorf("Cannot query functions in environment `%s`", environment)
	functions, err := r.functionLister.List(environment, pager.PagingParams{
		First:  first,
		Offset: offset,
	})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing Functions for environment %s", environment))
		return nil, externalErr
	}

	return r.functionConverter.ToGQLs(functions), nil
}
