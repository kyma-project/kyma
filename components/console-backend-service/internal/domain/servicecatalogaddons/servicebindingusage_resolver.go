package servicecatalogaddons

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	api "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=serviceBindingUsageOperations -output=automock -outpkg=automock -case=underscore
type serviceBindingUsageOperations interface {
	Create(namespace string, sb *api.ServiceBindingUsage) (*api.ServiceBindingUsage, error)
	Delete(namespace string, name string) error
	Find(namespace string, name string) (*api.ServiceBindingUsage, error)
	ListForServiceInstance(namespace string, instanceName string) ([]*api.ServiceBindingUsage, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type serviceBindingUsageResolver struct {
	operations serviceBindingUsageOperations
	converter  gqlServiceBindingUsageConverter
}

func newServiceBindingUsageResolver(op serviceBindingUsageOperations, converter gqlServiceBindingUsageConverter) *serviceBindingUsageResolver {
	return &serviceBindingUsageResolver{
		operations: op,
		converter:  converter,
	}
}

func (r *serviceBindingUsageResolver) CreateServiceBindingUsageMutation(ctx context.Context, namespace string, input *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error) {
	inBindingUsage, err := r.converter.InputToK8s(input)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s from input [%+v]", pretty.ServiceBindingUsage, input))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage)
	}

	bu, err := r.operations.Create(namespace, inBindingUsage)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s from input [%v]", pretty.ServiceBindingUsage, input))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(*input.Name), gqlerror.WithNamespace(namespace))
	}

	out, err := r.converter.ToGQL(bu)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceBindingUsage))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(*input.Name), gqlerror.WithNamespace(namespace))
	}

	return out, nil
}

func (r *serviceBindingUsageResolver) DeleteServiceBindingUsageMutation(ctx context.Context, serviceBindingUsageName, namespace string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	err := r.operations.Delete(namespace, serviceBindingUsageName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s with name `%s` from namespace `%s`", pretty.ServiceBindingUsage, serviceBindingUsageName, namespace))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(serviceBindingUsageName), gqlerror.WithNamespace(namespace))
	}

	return &gqlschema.DeleteServiceBindingUsageOutput{
		Namespace: namespace,
		Name:      serviceBindingUsageName,
	}, nil
}

func (r *serviceBindingUsageResolver) DeleteServiceBindingUsagesMutation(ctx context.Context, serviceBindingUsageNames []string, namespace string) ([]*gqlschema.DeleteServiceBindingUsageOutput, error) {
	output := []*gqlschema.DeleteServiceBindingUsageOutput{}

	for _, serviceBindingUsageName := range serviceBindingUsageNames {
		out, err := r.DeleteServiceBindingUsageMutation(ctx, serviceBindingUsageName, namespace)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while deleting %s with names %v from namespace `%s`", pretty.ServiceBindingUsages, serviceBindingUsageNames, namespace))
			return nil, gqlerror.New(err, pretty.ServiceBindingUsages, gqlerror.WithName(serviceBindingUsageName), gqlerror.WithNamespace(namespace))
		}
		output = append(output, out)
	}

	return output, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsageQuery(ctx context.Context, name, namespace string) (*gqlschema.ServiceBindingUsage, error) {
	usage, err := r.operations.Find(namespace, name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting single %s [name: %s, namespace: %s]", pretty.ServiceBindingUsage, name, namespace))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	out, err := r.converter.ToGQL(usage)
	if err != nil {
		glog.Error(
			errors.Wrapf(err,
				"while getting single %s [name: %s, namespace: %s]: while converting %s to QL representation", pretty.ServiceBindingUsage,
				name, namespace, pretty.ServiceBindingUsage))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	return out, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsagesOfInstanceQuery(ctx context.Context, instanceName, namespace string) ([]gqlschema.ServiceBindingUsage, error) {
	usages, err := r.operations.ListForServiceInstance(namespace, instanceName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s of instance [namespace: %s, instance: %s]", pretty.ServiceBindingUsages, namespace, instanceName))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsages, gqlerror.WithNamespace(namespace), gqlerror.WithCustomArgument("instanceName", instanceName))
	}
	out, err := r.converter.ToGQLs(usages)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s of instance [namespace: %s, instance: %s]", pretty.ServiceBindingUsages, namespace, instanceName))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsages, gqlerror.WithNamespace(namespace), gqlerror.WithCustomArgument("instanceName", instanceName))
	}
	return out, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsageEventSubscription(ctx context.Context, namespace string, resourceKind, resourceName *string) (<-chan gqlschema.ServiceBindingUsageEvent, error) {
	channel := make(chan gqlschema.ServiceBindingUsageEvent, 1)
	filter := func(bindingUsage *api.ServiceBindingUsage) bool {
		if bindingUsage == nil || bindingUsage.Namespace != namespace {
			return false
		}
		if resourceKind != nil {
			kindFilter := bindingUsage.Spec.UsedBy.Kind == *resourceKind
			nameFilter := true

			if resourceName != nil {
				nameFilter = bindingUsage.Spec.UsedBy.Name == *resourceName
			}
			return kindFilter && nameFilter
		}
		return true
	}

	bindingUsageListener := listener.NewBindingUsage(channel, filter, r.converter)

	r.operations.Subscribe(bindingUsageListener)
	go func() {
		defer close(channel)
		defer r.operations.Unsubscribe(bindingUsageListener)
		<-ctx.Done()
	}()

	return channel, nil
}
