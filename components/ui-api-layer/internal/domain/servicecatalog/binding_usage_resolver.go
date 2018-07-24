package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type serviceBindingUsageResolver struct {
	operations serviceBindingUsageOperations
	converter  bindingUsageConverter
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
		glog.Error(errors.Wrapf(err, "while creating ServiceBindingUsage from input [%+v]", input))
		return nil, r.genericErrorOnCreate()
	}
	bu, err := r.operations.Create(input.Environment, inBindingUsage)
	switch {
	case apierrors.IsAlreadyExists(err):
		return nil, fmt.Errorf("ServiceBindingUsage %s already exists", input.Name)
	case err != nil:
		glog.Error(errors.Wrapf(err, "while creating ServiceBindingUsage from input [%v]", input))
		return nil, r.genericErrorOnCreate()
	}

	out, err := r.converter.ToGQL(bu)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *serviceBindingUsageResolver) DeleteServiceBindingUsageMutation(ctx context.Context, serviceBindingUsageName, namespace string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	err := r.operations.Delete(namespace, serviceBindingUsageName)
	switch {
	case apierrors.IsNotFound(err):
		return nil, fmt.Errorf("ServiceBindingUsage %s not found", serviceBindingUsageName)
	case err != nil:
		glog.Error(errors.Wrapf(err, "while deleting ServiceBindingUsage %s", serviceBindingUsageName))
		return nil, errors.New("cannot delete ServiceBindingUsage")
	}

	return &gqlschema.DeleteServiceBindingUsageOutput{
		Environment: namespace,
		Name:        serviceBindingUsageName,
	}, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsageQuery(ctx context.Context, name, environment string) (*gqlschema.ServiceBindingUsage, error) {
	usage, err := r.operations.Find(environment, name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting single ServiceBindingUsage [name: %s, environment: %s]", name, environment))
		return nil, r.genericErrorOnSingleGet()
	}

	out, err := r.converter.ToGQL(usage)
	if err != nil {
		glog.Error(
			errors.Wrapf(err,
				"while getting single ServiceBindingUsage [name: %s, environment: %s]: while converting ServiceBindingUsage to QL representation",
				name, environment))
		return nil, r.genericErrorOnSingleGet()
	}
	return out, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsagesOfInstanceQuery(ctx context.Context, instanceName, env string) ([]gqlschema.ServiceBindingUsage, error) {
	usages, err :=
		r.operations.ListForServiceInstance(env, instanceName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceBindingUsages of instance [environment: %s, instance: %s]", env, instanceName))
		return nil, r.genericErrorOnMultipleGet()
	}
	out, err := r.converter.ToGQLs(usages)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting ServiceBindingUsages of instance [environment: %s, instance: %s]", env, instanceName))
		return nil, r.genericErrorOnMultipleGet()
	}
	return out, nil
}

func (*serviceBindingUsageResolver) genericErrorOnCreate() error {
	return errors.New("cannot create ServiceBindingUsage")
}

func (*serviceBindingUsageResolver) genericErrorOnSingleGet() error {
	return errors.New("cannot get ServiceBindingUsage")
}

func (*serviceBindingUsageResolver) genericErrorOnMultipleGet() error {
	return errors.New("Cannot get ServiceBindingUsages")
}
