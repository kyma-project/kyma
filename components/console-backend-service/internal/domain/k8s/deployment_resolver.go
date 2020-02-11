package k8s

import (
	"context"
	"github.com/blang/semver"
	"strings"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	scPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/pretty"
	scaPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
)

//go:generate mockery -name=deploymentLister -output=automock -outpkg=automock -case=underscore
type deploymentLister interface {
	List(namespace string) ([]*v1.Deployment, error)
	ListWithoutFunctions(namespace string) ([]*v1.Deployment, error)
	Find(name string, namespace string) (*v1.Deployment, error)
}

type deploymentResolver struct {
	deploymentLister    deploymentLister
	deploymentConverter *deploymentConverter
	scRetriever         shared.ServiceCatalogRetriever
	scaRetriever        shared.ServiceCatalogAddonsRetriever
}

func newDeploymentResolver(deploymentLister deploymentLister, scRetriever shared.ServiceCatalogRetriever, scaRetriever shared.ServiceCatalogAddonsRetriever) *deploymentResolver {
	return &deploymentResolver{
		deploymentLister: deploymentLister,
		scRetriever:      scRetriever,
		scaRetriever:     scaRetriever,
	}
}

func (r *deploymentResolver) DeploymentsQuery(ctx context.Context, namespace string, excludeFunctions *bool) ([]gqlschema.Deployment, error) {
	var deployments []*v1.Deployment
	var err error
	if excludeFunctions == nil || !*excludeFunctions {
		deployments, err = r.deploymentLister.List(namespace)
	} else {
		deployments, err = r.deploymentLister.ListWithoutFunctions(namespace)
	}

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s in namespace `%s`", pretty.Deployments, namespace))
		return nil, gqlerror.New(err, pretty.Deployments, gqlerror.WithNamespace(namespace))
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

	usages, err := r.scaRetriever.ServiceBindingUsage().ListForDeployment(deployment.Namespace, kind, deployment.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		glog.Error(errors.Wrapf(err, "while listing %s for %s in namespace `%s`, name `%s` and kind `%s`", scaPretty.ServiceBindingUsages, pretty.Deployment, deployment.Namespace, deployment.Name, kind))
		return nil, gqlerror.New(err, scaPretty.ServiceBindingUsages, gqlerror.WithNamespace(deployment.Namespace))
	}

	instanceNames := make(map[string]struct{})
	for _, usage := range usages {
		binding, err := r.scRetriever.ServiceBinding().Find(deployment.Namespace, usage.Spec.ServiceBindingRef.Name)
		if err != nil {
			if module.IsDisabledModuleError(err) {
				return nil, err
			}

			glog.Error(errors.Wrapf(err, "while gathering %s for namespace `%s` with name `%s`", scPretty.ServiceBinding, deployment.Namespace, usage.Spec.ServiceBindingRef.Name))
			return nil, gqlerror.New(err, scPretty.ServiceBinding, gqlerror.WithName(usage.Spec.ServiceBindingRef.Name), gqlerror.WithNamespace(deployment.Namespace))
		}

		if binding != nil {
			instanceNames[binding.Spec.InstanceRef.Name] = struct{}{}
		}
	}

	result := make([]string, 0, len(instanceNames))
	for name := range instanceNames {
		result = append(result, name)
	}

	return result, nil
}

func (r *deploymentResolver) KymaVersionQuery(ctx context.Context) (string, error) {
	name := "kyma-installer"
	namespace := "kyma-installer"

	deployment, err := r.deploymentLister.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting the %s in namespace `%s`", pretty.Deployment, namespace))
		return "", gqlerror.New(err, pretty.Deployment, gqlerror.WithNamespace(namespace))
	}

	deploymentImage := deployment.Spec.Template.Spec.Containers[0].Image
	deploymentImageSeparated := strings.FieldsFunc(deploymentImage, Split)

	source := deploymentImageSeparated[0]
	if source != "eu.gcr.io" {
		return deploymentImage, nil
	}

	version := deploymentImageSeparated[len(deploymentImageSeparated)-1]
	_, err = semver.Parse(version)
	if err != nil {
		branch := "master"
		if strings.HasPrefix(version,"PR-") {
			branch = "pull request"
		}
		return strings.Join([]string{branch, version}, " "), nil
	}

	return version, nil
}

func Split(r rune) bool {
	return r == '/' || r == ':'
}