package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	switch {
	case apierrors.IsAlreadyExists(err):
		return nil, fmt.Errorf("ServiceBinding %s already exists", serviceBindingName)
	case err != nil:
		glog.Error(errors.Wrapf(err, "while creating ServiceBinding %s", serviceBindingName))
		return nil, errors.New("cannot create ServiceBinding")
	}

	return r.converter.ToCreateOutputGQL(sb), nil
}

func (r *serviceBindingResolver) DeleteServiceBindingMutation(ctx context.Context, serviceBindingName, env string) (*gqlschema.DeleteServiceBindingOutput, error) {
	err := r.operations.Delete(env, serviceBindingName)
	switch {
	case apierrors.IsNotFound(err):
		return nil, fmt.Errorf("ServiceBinding %s not found", serviceBindingName)
	case err != nil:
		glog.Error(errors.Wrapf(err, "while deleting ServiceBinding %s", serviceBindingName))
		return nil, errors.New("cannot delete ServiceBinding")
	}

	return &gqlschema.DeleteServiceBindingOutput{
		Environment: env,
		Name:        serviceBindingName,
	}, nil
}

func (r *serviceBindingResolver) ServiceBindingQuery(ctx context.Context, name, env string) (*gqlschema.ServiceBinding, error) {
	binding, err := r.operations.Find(env, name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceBinding [name: %s, environment: %s]", name, env))
		return nil, errors.New("cannot get ServiceBinding")
	}

	return r.converter.ToGQL(binding), nil
}

func (r *serviceBindingResolver) ServiceBindingsToInstanceQuery(ctx context.Context, instanceName, environment string) ([]gqlschema.ServiceBinding, error) {
	list, err := r.operations.ListForServiceInstance(environment, instanceName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting many ServiceBindings to Instance [instance name: %s. environment: %s]", instanceName, environment))
		return []gqlschema.ServiceBinding{}, errors.New("cannot get ServiceBindings")
	}
	return r.converter.ToGQLs(list), nil
}
