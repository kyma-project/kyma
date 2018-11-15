package servicecatalog

import (
	"context"
	"encoding/json"

	"github.com/golang/glog"
	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/name"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type serviceBindingResolver struct {
	operations serviceBindingOperations
	converter  serviceBindingConverter
}

func newServiceBindingResolver(op serviceBindingOperations) *serviceBindingResolver {
	return &serviceBindingResolver{
		operations: op,
		converter:  serviceBindingConverter{},
	}
}

func (r *serviceBindingResolver) CreateServiceBindingMutation(ctx context.Context, serviceBindingName *string, serviceInstanceName, env string, parameters *gqlschema.JSON) (*gqlschema.CreateServiceBindingOutput, error) {
	sbToCreate := &api.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name.EmptyIfNil(serviceBindingName),
		},
		Spec: api.ServiceBindingSpec{
			ServiceInstanceRef: api.LocalObjectReference{
				Name: serviceInstanceName,
			},
		},
	}
	if parameters != nil {
		byteArray, err := json.Marshal(parameters)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while marshalling parameters %s `%s` parameters: %+v", pretty.ServiceBinding, serviceBindingName, parameters))
			return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(name.EmptyIfNil(serviceBindingName)), gqlerror.WithEnvironment(env))
		}
		sbToCreate.Spec.Parameters = &runtime.RawExtension{
			Raw: byteArray,
		}
	}

	sb, err := r.operations.Create(env, sbToCreate)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.ServiceBinding, serviceBindingName))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(name.EmptyIfNil(serviceBindingName)), gqlerror.WithEnvironment(env))
	}

	return r.converter.ToCreateOutputGQL(sb), nil
}

func (r *serviceBindingResolver) DeleteServiceBindingMutation(ctx context.Context, serviceBindingName, env string) (*gqlschema.DeleteServiceBindingOutput, error) {
	err := r.operations.Delete(env, serviceBindingName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s`", pretty.ServiceBinding, serviceBindingName))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(serviceBindingName), gqlerror.WithEnvironment(env))
	}

	return &gqlschema.DeleteServiceBindingOutput{
		Environment: env,
		Name:        serviceBindingName,
	}, nil
}

func (r *serviceBindingResolver) ServiceBindingQuery(ctx context.Context, name, env string) (*gqlschema.ServiceBinding, error) {
	binding, err := r.operations.Find(env, name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` in namespace `%s`", pretty.ServiceBinding, name, env))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(name), gqlerror.WithEnvironment(env))
	}

	out, err := r.converter.ToGQL(binding)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s `%s` in namespace `%s`", pretty.ServiceBinding, name, env))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(name), gqlerror.WithEnvironment(env))
	}

	return out, nil
}

func (r *serviceBindingResolver) ServiceBindingsToInstanceQuery(ctx context.Context, instanceName, environment string) (gqlschema.ServiceBindings, error) {
	list, err := r.operations.ListForServiceInstance(environment, instanceName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting many %s to Instance [instance name: %s. environment: %s]", pretty.ServiceBindings, instanceName, environment))
		return gqlschema.ServiceBindings{}, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithEnvironment(environment))
	}

	out, err := r.converter.ToGQLs(list)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting many %s for %s `%s`", pretty.ServiceBindings, pretty.ServiceInstance, instanceName))
		return gqlschema.ServiceBindings{}, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(instanceName), gqlerror.WithEnvironment(environment))
	}

	return out, nil
}

func (r *serviceBindingResolver) ServiceBindingEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceBindingEvent, error) {
	channel := make(chan gqlschema.ServiceBindingEvent, 1)
	filter := func(binding *api.ServiceBinding) bool {
		return binding != nil && binding.Namespace == environment
	}

	bindingListener := listener.NewBinding(channel, filter, &r.converter)

	r.operations.Subscribe(bindingListener)
	go func() {
		defer close(channel)
		defer r.operations.Unsubscribe(bindingListener)
		<-ctx.Done()
	}()

	return channel, nil
}
