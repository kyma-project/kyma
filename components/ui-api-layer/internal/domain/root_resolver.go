package domain

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/apicontroller"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/kubeless"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/ui"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type RootResolver struct {
	k8s       *k8s.Resolver
	kubeless  *kubeless.Resolver
	sc        *servicecatalog.Resolver
	re        *remoteenvironment.Resolver
	content   *content.Resolver
	ac        *apicontroller.Resolver
	idpPreset *ui.Resolver
}

func New(restConfig *rest.Config, contentCfg content.Config, reCfg remoteenvironment.Config, informerResyncPeriod time.Duration) (*RootResolver, error) {
	contentContainer, err := content.New(contentCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Content resolver")
	}

	scContainer, err := servicecatalog.New(restConfig, informerResyncPeriod, contentContainer.AsyncApiSpecGetter, contentContainer.ApiSpecGetter, contentContainer.ContentGetter)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing ServiceCatalog container")
	}

	reContainer, err := remoteenvironment.New(restConfig, reCfg, informerResyncPeriod, contentContainer.AsyncApiSpecGetter)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing RemoteEnvironment resolver")
	}

	k8sResolver, err := k8s.New(restConfig, reContainer.RELister, informerResyncPeriod, scContainer.ServiceBindingUsageLister, scContainer.ServiceBindingGetter)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing K8S resolver")
	}

	kubelessResolver, err := kubeless.New(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Kubeless resolver")
	}

	acResolver, err := apicontroller.New(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing API controller resolver")
	}

	idpPresetResolver, err := ui.New(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing idpPreset resolver")
	}

	return &RootResolver{
		k8s:       k8sResolver,
		kubeless:  kubelessResolver,
		re:        reContainer.Resolver,
		sc:        scContainer.Resolver,
		content:   contentContainer.Resolver,
		ac:        acResolver,
		idpPreset: idpPresetResolver,
	}, nil
}

// WaitForCacheSync waits for caches to populate. This is blocking operation.
func (r *RootResolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.re.WaitForCacheSync(stopCh)
	r.sc.WaitForCacheSync(stopCh)
	r.k8s.WaitForCacheSync(stopCh)
	r.kubeless.WaitForCacheSync(stopCh)
	r.ac.WaitForCacheSync(stopCh)
	r.idpPreset.WaitForCacheSync(stopCh)
	r.content.WaitForCacheSync(stopCh)
}

// K8S

func (r *RootResolver) Query_environments(ctx context.Context, remoteEnvironment *string) ([]gqlschema.Environment, error) {
	return r.k8s.EnvironmentsQuery(ctx, remoteEnvironment)
}

func (r *RootResolver) Query_deployments(ctx context.Context, environment string, excludeFunctions *bool) ([]gqlschema.Deployment, error) {
	return r.k8s.DeploymentsQuery(ctx, environment, excludeFunctions)
}

func (r *RootResolver) Deployment_boundServiceInstanceNames(ctx context.Context, deployment *gqlschema.Deployment) ([]string, error) {
	return r.k8s.DeploymentBoundServiceInstanceNamesField(ctx, deployment)
}

func (r *RootResolver) Query_limitRanges(ctx context.Context, env string) ([]gqlschema.LimitRange, error) {
	return r.k8s.LimitRangesQuery(ctx, env)
}

func (r *RootResolver) Query_resourceQuotas(ctx context.Context, environment string) ([]gqlschema.ResourceQuota, error) {
	return r.k8s.ResourceQuotasQuery(ctx, environment)
}

// Kubeless

func (r *RootResolver) Query_functions(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.Function, error) {
	return r.kubeless.FunctionsQuery(ctx, environment, first, offset)
}

// Service Catalog

func (r *RootResolver) Mutation_createServiceInstance(ctx context.Context, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error) {
	return r.sc.CreateServiceInstanceMutation(ctx, params)
}

func (r *RootResolver) Subscription_serviceInstanceEvent(ctx context.Context, environment string) (<-chan gqlschema.ServiceInstanceEvent, error) {
	return r.sc.ServiceInstanceEventSubscription(ctx, environment)
}

func (r *RootResolver) Query_serviceInstance(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	return r.sc.ServiceInstanceQuery(ctx, name, environment)
}

func (r *RootResolver) Query_serviceInstances(ctx context.Context, environment string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error) {
	return r.sc.ServiceInstancesQuery(ctx, environment, first, offset, status)
}

func (r *RootResolver) Query_serviceClasses(ctx context.Context, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
	return r.sc.ServiceClassesQuery(ctx, first, offset)
}

func (r *RootResolver) Query_serviceClass(ctx context.Context, name string) (*gqlschema.ServiceClass, error) {
	return r.sc.ServiceClassQuery(ctx, name)
}

func (r *RootResolver) Query_serviceBrokers(ctx context.Context, first *int, offset *int) ([]gqlschema.ServiceBroker, error) {
	return r.sc.ServiceBrokersQuery(ctx, first, offset)
}

func (r *RootResolver) Query_serviceBroker(ctx context.Context, name string) (*gqlschema.ServiceBroker, error) {
	return r.sc.ServiceBrokerQuery(ctx, name)
}

func (r *RootResolver) Query_usageKinds(ctx context.Context, first *int, offset *int) ([]gqlschema.UsageKind, error) {
	return r.sc.ListUsageKinds(ctx, first, offset)
}

func (r *RootResolver) Query_usageKindResources(ctx context.Context, usageKind string, environment string) ([]gqlschema.UsageKindResource, error) {
	return r.sc.ListServiceUsageKindResources(ctx, usageKind, environment)
}

func (r *RootResolver) ServiceClass_activated(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	return r.sc.ServiceClassActivatedField(ctx, obj)
}

func (r *RootResolver) ServiceClass_plans(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error) {
	return r.sc.ServiceClassPlansField(ctx, obj)
}

func (r *RootResolver) ServiceClass_apiSpec(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	return r.sc.ServiceClassApiSpecField(ctx, obj)
}

func (r *RootResolver) ServiceClass_asyncApiSpec(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	return r.sc.ServiceClassAsyncApiSpecField(ctx, obj)
}

func (r *RootResolver) ServiceClass_content(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	return r.sc.ServiceClassContentField(ctx, obj)
}

func (r *RootResolver) ServiceInstance_servicePlan(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServicePlan, error) {
	return r.sc.ServiceInstanceServicePlanField(ctx, obj)
}

func (r *RootResolver) ServiceInstance_serviceClass(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error) {
	return r.sc.ServiceInstanceServiceClassField(ctx, obj)
}

func (r *RootResolver) ServiceInstance_bindable(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error) {
	return r.sc.ServiceInstanceBindableField(ctx, obj)
}

func (r *RootResolver) ServiceInstance_serviceBindings(ctx context.Context, obj *gqlschema.ServiceInstance) ([]gqlschema.ServiceBinding, error) {
	return r.sc.ServiceBindingsToInstanceQuery(ctx, obj.Name, obj.Environment)
}

func (r *RootResolver) ServiceInstance_serviceBindingUsages(ctx context.Context, obj *gqlschema.ServiceInstance) ([]gqlschema.ServiceBindingUsage, error) {
	return r.sc.ServiceBindingUsagesOfInstanceQuery(ctx, obj.Name, obj.Environment)
}

func (r *RootResolver) Mutation_deleteServiceInstance(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	return r.sc.DeleteServiceInstanceMutation(ctx, name, environment)
}

func (r *RootResolver) Mutation_createServiceBinding(ctx context.Context, serviceBindingName, serviceInstanceName, env string) (*gqlschema.CreateServiceBindingOutput, error) {
	return r.sc.CreateServiceBindingMutation(ctx, serviceBindingName, serviceInstanceName, env)
}

func (r *RootResolver) Mutation_deleteServiceBinding(ctx context.Context, serviceBindingName string, env string) (*gqlschema.DeleteServiceBindingOutput, error) {
	return r.sc.DeleteServiceBindingMutation(ctx, serviceBindingName, env)
}

func (r *RootResolver) Query_serviceBinding(ctx context.Context, name string, environment string) (*gqlschema.ServiceBinding, error) {
	return r.sc.ServiceBindingQuery(ctx, name, environment)
}

func (r *RootResolver) Mutation_createServiceBindingUsage(ctx context.Context, createServiceBindingUsageInput *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error) {
	return r.sc.CreateServiceBindingUsageMutation(ctx, createServiceBindingUsageInput)
}

func (r *RootResolver) Mutation_deleteServiceBindingUsage(ctx context.Context, serviceBindingUsageName string, env string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	return r.sc.DeleteServiceBindingUsageMutation(ctx, serviceBindingUsageName, env)
}

func (r *RootResolver) Query_serviceBindingUsage(ctx context.Context, name, environment string) (*gqlschema.ServiceBindingUsage, error) {
	return r.sc.ServiceBindingUsageQuery(ctx, name, environment)
}

func (r *RootResolver) ServiceBindingUsage_serviceBinding(ctx context.Context, obj *gqlschema.ServiceBindingUsage) (*gqlschema.ServiceBinding, error) {
	return r.sc.ServiceBindingQuery(ctx, obj.ServiceBindingName, obj.Environment)
}

func (r *RootResolver) ServiceBinding_secret(ctx context.Context, serviceBinding *gqlschema.ServiceBinding) (*gqlschema.Secret, error) {
	return r.k8s.SecretQuery(ctx, serviceBinding.SecretName, serviceBinding.Environment)
}

func (r *RootResolver) Query_content(ctx context.Context, contentType, id string) (*gqlschema.JSON, error) {
	return r.content.ContentQuery(ctx, contentType, id)
}

func (r *RootResolver) Query_topics(ctx context.Context, input []gqlschema.InputTopic, internal *bool) ([]gqlschema.TopicEntry, error) {
	return r.content.TopicsQuery(ctx, input, internal)
}

// Remote Environments

func (r *RootResolver) Mutation_enableRemoteEnvironment(ctx context.Context, remoteEnvironment string, environment string) (*gqlschema.EnvironmentMapping, error) {
	return r.re.EnableRemoteEnvironmentMutation(ctx, remoteEnvironment, environment)
}

func (r *RootResolver) Mutation_disableRemoteEnvironment(ctx context.Context, remoteEnvironment string, environment string) (*gqlschema.EnvironmentMapping, error) {
	return r.re.DisableRemoteEnvironmentMutation(ctx, remoteEnvironment, environment)
}

func (r *RootResolver) Query_remoteEnvironment(ctx context.Context, name string) (*gqlschema.RemoteEnvironment, error) {
	return r.re.RemoteEnvironmentQuery(ctx, name)
}

func (r *RootResolver) Query_remoteEnvironments(ctx context.Context, environment *string, first *int, offset *int) ([]gqlschema.RemoteEnvironment, error) {
	return r.re.RemoteEnvironmentsQuery(ctx, environment, first, offset)
}

func (r *RootResolver) Query_connectorService(ctx context.Context, remoteEnvironment string) (gqlschema.ConnectorService, error) {
	return r.re.ConnectorServiceQuery(ctx, remoteEnvironment)
}

func (r *RootResolver) RemoteEnvironment_enabledInEnvironments(ctx context.Context, obj *gqlschema.RemoteEnvironment) ([]string, error) {
	return r.re.RemoteEnvironmentEnabledInEnvironmentsField(ctx, obj)
}

func (r *RootResolver) RemoteEnvironment_status(ctx context.Context, obj *gqlschema.RemoteEnvironment) (gqlschema.RemoteEnvironmentStatus, error) {
	return r.re.RemoteEnvironmentStatusField(ctx, obj)
}

func (r *RootResolver) Query_eventActivations(ctx context.Context, environment string) ([]gqlschema.EventActivation, error) {
	return r.re.EventActivationsQuery(ctx, environment)
}

func (r *RootResolver) EventActivation_events(ctx context.Context, eventActivation *gqlschema.EventActivation) ([]gqlschema.EventActivationEvent, error) {
	return r.re.EventActivationEventsField(ctx, eventActivation)
}

// API controller

func (r *RootResolver) Query_apis(ctx context.Context, environment string, serviceName *string, hostname *string) ([]gqlschema.API, error) {
	return r.ac.APIsQuery(ctx, environment, serviceName, hostname)
}

// UI

func (r *RootResolver) Query_IDPPreset(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	return r.idpPreset.IDPPresetQuery(ctx, name)
}

func (r *RootResolver) Query_IDPPresets(ctx context.Context, first *int, offset *int) ([]gqlschema.IDPPreset, error) {
	return r.idpPreset.IDPPresetsQuery(ctx, first, offset)
}

func (r *RootResolver) Mutation_createIDPPreset(ctx context.Context, name string, issuer string, jwksUri string) (*gqlschema.IDPPreset, error) {
	return r.idpPreset.CreateIDPPresetMutation(ctx, name, issuer, jwksUri)
}

func (r *RootResolver) Mutation_deleteIDPPreset(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	return r.idpPreset.DeleteIDPPresetMutation(ctx, name)
}
