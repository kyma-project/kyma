package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	scPretty "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
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
		glog.Error(errors.Wrapf(err, "while listing %s in environment `%s`", pretty.Deployments, environment))
		return nil, gqlerror.New(err, pretty.Deployments, gqlerror.WithEnvironment(environment))
	}

	return r.deploymentConverter.ToGQLs(deployments), nil
}

func (r *deploymentResolver) DeploymentBoundServiceInstanceNamesField(ctx context.Context, deployment *gqlschema.Deployment) ([]string, error) {
	if deployment == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve ServiceInstanceNames for %s"), pretty.Deployment, pretty.Deployment)
		return nil, gqlerror.NewInternal()
	}

	kind := "deployment"
	if _, exists := deployment.Labels["function"]; exists {
		kind = "function"
	}

	usages, err := r.serviceBindingUsageLister.ListForDeployment(deployment.Environment, kind, deployment.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for %s in environment `%s`, name `%s` and kind `%s`", scPretty.ServiceBindingUsages, pretty.Deployment, deployment.Environment, deployment.Name, kind))
		return nil, gqlerror.New(err, scPretty.ServiceBindingUsages, gqlerror.WithEnvironment(deployment.Environment))
	}

	instanceNames := make(map[string]struct{})
	for _, usage := range usages {
		binding, err := r.serviceBindingGetter.Find(deployment.Environment, usage.Spec.ServiceBindingRef.Name)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while gathering %s for environment `%s` with name `%s`", scPretty.ServiceBinding, deployment.Environment, usage.Spec.ServiceBindingRef.Name))
			return nil, gqlerror.New(err, scPretty.ServiceBinding, gqlerror.WithName(usage.Spec.ServiceBindingRef.Name), gqlerror.WithEnvironment(deployment.Environment))
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
