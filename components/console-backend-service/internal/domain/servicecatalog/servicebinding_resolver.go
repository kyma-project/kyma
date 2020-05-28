package servicecatalog

import (
	"context"
	"encoding/json"

	"github.com/golang/glog"
	api "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/name"
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

func (r *serviceBindingResolver) CreateServiceBindingMutation(ctx context.Context, serviceBindingName *string, serviceInstanceName, namespace string, parameters gqlschema.JSON) (*gqlschema.CreateServiceBindingOutput, error) {
	sbToCreate := &api.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name.EmptyIfNil(serviceBindingName),
		},
		Spec: api.ServiceBindingSpec{
			InstanceRef: api.LocalObjectReference{
				Name: serviceInstanceName,
			},
		},
	}
	if parameters != nil {
		byteArray, err := json.Marshal(parameters)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while marshalling parameters %s `%s` parameters: %+v", pretty.ServiceBinding, name.EmptyIfNil(serviceBindingName), parameters))
			return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(name.EmptyIfNil(serviceBindingName)), gqlerror.WithNamespace(namespace))
		}
		sbToCreate.Spec.Parameters = &runtime.RawExtension{
			Raw: byteArray,
		}
	}

	sb, err := r.operations.Create(namespace, sbToCreate)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.ServiceBinding, name.EmptyIfNil(serviceBindingName)))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(name.EmptyIfNil(serviceBindingName)), gqlerror.WithNamespace(namespace))
	}

	return r.converter.ToCreateOutputGQL(sb), nil
}

func (r *serviceBindingResolver) DeleteServiceBindingMutation(ctx context.Context, serviceBindingName, namespace string) (*gqlschema.DeleteServiceBindingOutput, error) {
	err := r.operations.Delete(namespace, serviceBindingName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s`", pretty.ServiceBinding, serviceBindingName))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(serviceBindingName), gqlerror.WithNamespace(namespace))
	}

	return &gqlschema.DeleteServiceBindingOutput{
		Namespace: namespace,
		Name:      serviceBindingName,
	}, nil
}

func (r *serviceBindingResolver) ServiceBindingQuery(ctx context.Context, name, namespace string) (*gqlschema.ServiceBinding, error) {
	binding, err := r.operations.Find(namespace, name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` in namespace `%s`", pretty.ServiceBinding, name, namespace))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	out, err := r.converter.ToGQL(binding)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s `%s` in namespace `%s`", pretty.ServiceBinding, name, namespace))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return out, nil
}

func (r *serviceBindingResolver) ServiceBindingsToInstanceQuery(ctx context.Context, instanceName, namespace string) (*gqlschema.ServiceBindings, error) {
	list, err := r.operations.ListForServiceInstance(namespace, instanceName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting many %s to Instance [instance name: %s. namespace: %s]", pretty.ServiceBindings, instanceName, namespace))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithNamespace(namespace))
	}

	out, err := r.converter.ToGQLs(list)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting many %s for %s `%s`", pretty.ServiceBindings, pretty.ServiceInstance, instanceName))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(instanceName), gqlerror.WithNamespace(namespace))
	}

	return out, nil
}

func (r *serviceBindingResolver) ServiceBindingEventSubscription(ctx context.Context, namespace string) (<-chan *gqlschema.ServiceBindingEvent, error) {
	channel := make(chan *gqlschema.ServiceBindingEvent, 1)
	filter := func(binding *api.ServiceBinding) bool {
		return binding != nil && binding.Namespace == namespace
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
