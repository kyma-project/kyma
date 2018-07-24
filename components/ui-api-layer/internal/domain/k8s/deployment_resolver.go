package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/api/apps/v1beta2"
)

type deploymentResolver struct {
	deploymentLister          deploymentLister
	deploymentConverter       *deploymentConverter
	serviceBindingUsageLister ServiceBindingUsageLister
	serviceBindingGetter      ServiceBindingGetter
}

func newDeploymentResolver(deploymentLister deploymentLister, serviceBindingUsageLister ServiceBindingUsageLister, serviceBindingGetter ServiceBindingGetter) *deploymentResolver {
	return &deploymentResolver{
		deploymentLister:          deploymentLister,
		serviceBindingUsageLister: serviceBindingUsageLister,
		serviceBindingGetter:      serviceBindingGetter,
	}
}

func (r *deploymentResolver) DeploymentsQuery(ctx context.Context, environment string, excludeFunctions *bool) ([]gqlschema.Deployment, error) {
	var deployments []*v1beta2.Deployment
	var err error
	if excludeFunctions == nil || !*excludeFunctions {
		deployments, err = r.deploymentLister.List(environment)
	} else {
		deployments, err = r.deploymentLister.ListWithoutFunctions(environment)
	}

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing Deployments in environment `%s`", environment))
		return nil, r.genericError()
	}

	return r.deploymentConverter.ToGQLs(deployments), nil
}

func (r *deploymentResolver) DeploymentBoundServiceInstanceNamesField(ctx context.Context, deployment *gqlschema.Deployment) ([]string, error) {
	if deployment == nil {
		glog.Error(errors.New("Deployment cannot be empty in order to resolve ServiceInstanceNames for Deployment"))
		return nil, r.serviceInstanceNamesError()
	}

	kind := "deployment"
	if _, exists := deployment.Labels["function"]; exists {
		kind = "function"
	}

	usages, err := r.serviceBindingUsageLister.ListForDeployment(deployment.Environment, kind, deployment.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing ServiceBindingUsages for Deployment in environment `%s`, name `%s` and kind `%s`", deployment.Environment, deployment.Name, kind))
		return nil, r.serviceInstanceNamesError()
	}

	instanceNames := make(map[string]struct{})
	for _, usage := range usages {
		binding, err := r.serviceBindingGetter.Find(deployment.Environment, usage.Spec.ServiceBindingRef.Name)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while gathering ServiceBinding for environment `%s` environment with name `%s`", deployment.Environment, usage.Spec.ServiceBindingRef.Name))
			return nil, r.serviceInstanceNamesError()
		}

		if binding != nil {
			instanceNames[binding.Spec.ServiceInstanceRef.Name] = struct{}{}
		}
	}

	result := make([]string, 0, len(instanceNames))
	for name := range instanceNames {
		result = append(result, name)
	}

	return result, nil
}

func (r *deploymentResolver) genericError() error {
	return errors.New("Cannot get Deployment")
}

func (r *deploymentResolver) serviceInstanceNamesError() error {
	return errors.New("Cannot list ServiceInstance names")
}
