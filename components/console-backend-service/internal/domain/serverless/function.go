package serverless

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/function"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

//go:generate mockgen -source=function.go -destination=mocks/function_service.go
type FunctionService interface {
	List(namespace string) ([]*v1alpha1.Function, error)
	Delete(name string, namespace string) error
	Find(name string, namespace string) (*v1alpha1.Function, error)
	Create(name string, namespace string, labels gqlschema.Labels, size string, runtime string) (*v1alpha1.Function, error)
	Update(name string, namespace string, params gqlschema.FunctionUpdateInput) (*v1alpha1.Function, error)
}

const (
	USAGE_KIND = "knative-service"
)

func (r *resolver) FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error) {
	items, err := r.functionService.List(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", KindFunctions))
		return nil, gqlerror.New(err, KindFunctions, gqlerror.WithNamespace(namespace))
	}

	sortedItems := function.SortFunctions(items)

	functions := function.ToGQLs(sortedItems)

	return functions, nil
}

func (r *resolver) FunctionQuery(ctx context.Context, name, namespace string) (*gqlschema.Function, error) {
	item, err := r.functionService.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s `%s`", KindFunction, name))
		return nil, gqlerror.New(err, KindFunction, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	f := function.ToGQL(item)

	return f, nil
}

func (r *resolver) ServiceBindingUsagesField(ctx context.Context, function *gqlschema.Function) ([]gqlschema.ServiceBindingUsage, error) {
	if function == nil {
		glog.Error(errors.Errorf("%s cannot be empty in order to resolve `serviceBindingUsages` field", KindFunction))
		return nil, gqlerror.NewInternal()
	}

	items, err := r.scaRetriever.ServiceBindingUsage().ListByUsageKind(function.Namespace, USAGE_KIND, function.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", pretty.ServiceBindingUsages, KindFunction, function.Name))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsages)
	}

	serviceBindingUsages, err := r.scaRetriever.ServiceBindingUsageConverter().ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s for %s %s", pretty.ServiceBindingUsages, KindFunction, function.Name))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsages)
	}

	return serviceBindingUsages, nil
}

func (r *resolver) CreateFunction(ctx context.Context, name string, namespace string, labels gqlschema.Labels, size string, runtime string) (*gqlschema.Function, error) {
	item, err := r.functionService.Create(name, namespace, labels, size, runtime)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", KindFunction, name))
		return nil, gqlerror.New(err, KindFunction, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	f := function.ToGQL(item)

	return f, nil
}

func (r *resolver) UpdateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionUpdateInput) (*gqlschema.Function, error) {
	item, err := r.functionService.Update(name, namespace, params)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s`", KindFunction, name))
		return nil, gqlerror.New(err, KindFunction, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	f := function.ToGQL(item)

	return f, nil
}

func (r *resolver) DeleteFunction(ctx context.Context, name string, namespace string) (*gqlschema.FunctionMutationOutput, error) {
	err := r.functionService.Delete(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s`", KindFunction, name))
		return nil, gqlerror.New(err, KindFunction, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	return &gqlschema.FunctionMutationOutput{
		Name:      name,
		Namespace: namespace,
	}, nil
}
