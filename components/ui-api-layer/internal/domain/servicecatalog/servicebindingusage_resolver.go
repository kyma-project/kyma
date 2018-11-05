package servicecatalog

import (
	"context"

	"github.com/golang/glog"
	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

type serviceBindingUsageResolver struct {
	operations serviceBindingUsageOperations
	converter  serviceBindingUsageConverter
}

func newServiceBindingUsageResolver(op serviceBindingUsageOperations) *serviceBindingUsageResolver {
	return &serviceBindingUsageResolver{
		operations: op,
		converter:  newBindingUsageConverter(),
	}
}

func (r *serviceBindingUsageResolver) CreateServiceBindingUsageMutation(ctx context.Context, input *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error) {
	inBindingUsage, err := r.converter.InputToK8s(input)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s from input [%+v]", pretty.ServiceBindingUsage, input))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage)
	}
	bu, err := r.operations.Create(input.Environment, inBindingUsage)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s from input [%v]", pretty.ServiceBindingUsage, input))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(*input.Name), gqlerror.WithEnvironment(input.Environment))
	}

	out, err := r.converter.ToGQL(bu)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceBindingUsage))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(*input.Name), gqlerror.WithEnvironment(input.Environment))
	}

	return out, nil
}

func (r *serviceBindingUsageResolver) DeleteServiceBindingUsageMutation(ctx context.Context, serviceBindingUsageName, namespace string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	err := r.operations.Delete(namespace, serviceBindingUsageName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s with name `%s` from environment `%s`", pretty.ServiceBindingUsage, serviceBindingUsageName, namespace))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(serviceBindingUsageName), gqlerror.WithEnvironment(namespace))
	}

	return &gqlschema.DeleteServiceBindingUsageOutput{
		Environment: namespace,
		Name:        serviceBindingUsageName,
	}, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsageQuery(ctx context.Context, name, environment string) (*gqlschema.ServiceBindingUsage, error) {
	usage, err := r.operations.Find(environment, name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting single %s [name: %s, environment: %s]", pretty.ServiceBindingUsage, name, environment))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(name), gqlerror.WithEnvironment(environment))
	}

	out, err := r.converter.ToGQL(usage)
	if err != nil {
		glog.Error(
			errors.Wrapf(err,
				"while getting single %s [name: %s, environment: %s]: while converting %s to QL representation", pretty.ServiceBindingUsage,
				name, environment, pretty.ServiceBindingUsage))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(name), gqlerror.WithEnvironment(environment))
	}
	return out, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsagesOfInstanceQuery(ctx context.Context, instanceName, env string) ([]gqlschema.ServiceBindingUsage, error) {
	usages, err := r.operations.ListForServiceInstance(env, instanceName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s of instance [environment: %s, instance: %s]", pretty.ServiceBindingUsages, env, instanceName))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsages, gqlerror.WithEnvironment(env), gqlerror.WithCustomArgument("instanceName", instanceName))
	}
	out, err := r.converter.ToGQLs(usages)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s of instance [environment: %s, instance: %s]", pretty.ServiceBindingUsages, env, instanceName))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsages, gqlerror.WithEnvironment(env), gqlerror.WithCustomArgument("instanceName", instanceName))
	}
	return out, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsageEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceBindingUsageEvent, error) {
	channel := make(chan gqlschema.ServiceBindingUsageEvent, 1)
	filter := func(bindingUsage *api.ServiceBindingUsage) bool {
		return bindingUsage != nil && bindingUsage.Namespace == environment
	}

	bindingUsageListener := listener.NewBindingUsage(channel, filter, &r.converter)

	r.operations.Subscribe(bindingUsageListener)
	go func() {
		defer close(channel)
		defer r.operations.Unsubscribe(bindingUsageListener)
		<-ctx.Done()
	}()

	return channel, nil
}
