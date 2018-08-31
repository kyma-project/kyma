package servicecatalog

import (
	"context"

	"github.com/golang/glog"
	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r *serviceBindingResolver) CreateServiceBindingMutation(ctx context.Context, serviceBindingName, serviceInstanceName, env string) (*gqlschema.CreateServiceBindingOutput, error) {
	sb, err := r.operations.Create(env, &api.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceBindingName,
		},
		Spec: api.ServiceBindingSpec{
			ServiceInstanceRef: api.LocalObjectReference{
				Name: serviceInstanceName,
			},
		},
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.ServiceBinding, serviceBindingName))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithName(serviceBindingName), gqlerror.WithEnvironment(env))
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

	return r.converter.ToGQL(binding), nil
}

func (r *serviceBindingResolver) ServiceBindingsToInstanceQuery(ctx context.Context, instanceName, environment string) ([]gqlschema.ServiceBinding, error) {
	list, err := r.operations.ListForServiceInstance(environment, instanceName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting many %s to Instance [instance name: %s. environment: %s]", pretty.ServiceBindings, instanceName, environment))
		return nil, gqlerror.New(err, pretty.ServiceBinding, gqlerror.WithEnvironment(environment))
	}

	return r.converter.ToGQLs(list), nil
}
