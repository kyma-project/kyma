package serverless

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/pretty"
	scaPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type functionResolver struct {
	functionService   functionSvc
	functionConverter gqlFunctionConverter
	cfg               *Config
	scaRetriever      shared.ServiceCatalogAddonsRetriever
}

func newFunctionResolver(functionService functionSvc, functionConverter gqlFunctionConverter, cfg *Config, scaRetriever shared.ServiceCatalogAddonsRetriever) *functionResolver {
	return &functionResolver{
		functionService:   functionService,
		functionConverter: functionConverter,
		cfg:               cfg,
		scaRetriever:      scaRetriever,
	}
}

func (r *functionResolver) FunctionQuery(ctx context.Context, name string, namespace string) (*gqlschema.Function, error) {
	item, err := r.functionService.Find(namespace, name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s [name: %s, namespace: %s]", pretty.Functions, name, namespace))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	function, err := r.functionConverter.ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s GQL [name: %s, namespace: %s]", pretty.Function, name, namespace))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return function, nil
}

func (r *functionResolver) FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error) {
	items, err := r.functionService.List(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s [namespace: %s]", pretty.Functions, namespace))
		return nil, gqlerror.New(err, pretty.Functions, gqlerror.WithNamespace(namespace))
	}

	functions, err := r.functionConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s GQLs [namespace: %s]", pretty.Functions, namespace))
		return nil, gqlerror.New(err, pretty.Functions, gqlerror.WithNamespace(namespace))
	}

	return functions, nil
}

func (r *functionResolver) CreateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionMutationInput) (*gqlschema.Function, error) {
	item, err := r.functionConverter.ToFunction(name, namespace, params)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting GQL to %s [name: %s, namespace: %s, params: %v]", pretty.Function, name, namespace, params))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	createdItem, err := r.functionService.Create(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s [name: %s, namespace: %s]", pretty.Function, name, namespace))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	function, err := r.functionConverter.ToGQL(createdItem)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s to GQL [name: %s, namespace: %s]", pretty.Function, name, namespace))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return function, nil
}

func (r *functionResolver) UpdateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionMutationInput) (*gqlschema.Function, error) {
	item, err := r.functionConverter.ToFunction(name, namespace, params)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting GQL to %s [name: %s, namespace: %s, params: %v]", pretty.Function, name, namespace, params))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	updatedItem, err := r.functionService.Update(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s [name: %s, namespace: %s]", pretty.Function, name, namespace))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	function, err := r.functionConverter.ToGQL(updatedItem)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s to GQL [name: %s, namespace: %s]", pretty.Function, name, namespace))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return function, nil
}

func (r *functionResolver) DeleteFunction(ctx context.Context, function gqlschema.FunctionMetadataInput) (*gqlschema.FunctionMetadata, error) {
	err := r.functionService.Delete(function)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s [name: %s, namespace: %s]", pretty.Function, function.Name, function.Namespace))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(function.Name), gqlerror.WithNamespace(function.Namespace))
	}

	err = r.scaRetriever.ServiceBindingUsage().DeleteAllByUsageKind(function.Namespace, r.cfg.UsageKind, function.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while deleting %s for %s [name: %s, namespace: %s]", scaPretty.ServiceBindingUsages, pretty.Function, function.Name, function.Namespace))
		return nil, gqlerror.New(err, pretty.Function, gqlerror.WithName(function.Name), gqlerror.WithNamespace(function.Namespace))
	}

	return &gqlschema.FunctionMetadata{
		Name:      function.Name,
		Namespace: function.Namespace,
	}, nil
}

func (r *functionResolver) DeleteManyFunctions(ctx context.Context, functions []gqlschema.FunctionMetadataInput) ([]gqlschema.FunctionMetadata, error) {
	deletedFunctions := make([]gqlschema.FunctionMetadata, 0)
	for _, function := range functions {
		_, err := r.DeleteFunction(ctx, function)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while deleting %s [namespace: %s]", pretty.Functions, function.Namespace))
			return deletedFunctions, gqlerror.New(err, pretty.Functions, gqlerror.WithNamespace(function.Namespace))
		}

		deletedFunctions = append(deletedFunctions, gqlschema.FunctionMetadata{
			Name:      function.Name,
			Namespace: function.Namespace,
		})
	}
	return deletedFunctions, nil
}

func (r *functionResolver) FunctionEventSubscription(ctx context.Context, namespace string, functionName *string) (<-chan gqlschema.FunctionEvent, error) {
	channel := make(chan gqlschema.FunctionEvent, 1)
	filter := func(entity *v1alpha1.Function) bool {
		if entity == nil {
			return false
		}

		correctNamespace := entity.Namespace == namespace
		if functionName != nil {
			return correctNamespace && entity.Name == *functionName
		}
		return correctNamespace
	}

	listener := newFunctionListener(channel, filter, r.functionConverter)

	r.functionService.Subscribe(listener)
	go func() {
		defer close(channel)
		defer r.functionService.Unsubscribe(listener)
		<-ctx.Done()
	}()

	return channel, nil
}
