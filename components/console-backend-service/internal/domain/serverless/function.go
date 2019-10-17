package serverless

import (
	"context"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/convert"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

func (r *resolver) FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error) {
	items, err := r.functionService.List(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.Functions))
		return nil, gqlerror.New(err, pretty.Functions, gqlerror.WithNamespace(namespace))
	}

	functions := convert.FunctionsToGQLs(items)

	return functions, nil
}

func (r *resolver) FunctionQuery(ctx context.Context, name, namespace string) (*gqlschema.Function, error) {
	item, err := r.functionService.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s", pretty.Function))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	function := convert.FunctionToGQL(item)

	return function, nil
}

func (r *resolver) CreateFunction(ctx context.Context, name string, namespace string, labels gqlschema.Labels, size string, runtime string) (gqlschema.Function, error) {
	item, err := r.functionService.Create(name, namespace, labels, size, runtime)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.Function, name))
		return gqlschema.Function{}, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	function := convert.FunctionToGQL(item)

	return *function, nil
}

func (r *resolver) UpdateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionUpdateInput) (gqlschema.Function, error) {
	item, err := r.functionService.Update(name, namespace, params)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while editing %s `%s`", pretty.Function, name))
		return gqlschema.Function{}, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	function := convert.FunctionToGQL(item)

	return *function, nil
}

func (r *resolver) DeleteFunction(ctx context.Context, name string, namespace string) (gqlschema.FunctionMutationOutput, error) {
	err := r.functionService.Delete(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s`", pretty.Function, name))
		return gqlschema.FunctionMutationOutput{}, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	return gqlschema.FunctionMutationOutput{
		Name:      name,
		Namespace: namespace,
	}, nil
}
