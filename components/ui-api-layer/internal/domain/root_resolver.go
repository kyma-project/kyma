package domain

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalogaddons"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/ui"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/experimental"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/apicontroller"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/authentication"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/kubeless"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type RootResolver struct {
	ui  *ui.Resolver
	k8s *k8s.Resolver

	sc             *servicecatalog.PluggableContainer
	sca            *servicecatalogaddons.PluggableContainer
	app            *application.PluggableContainer
	content        *content.PluggableContainer
	kubeless       *kubeless.PluggableResolver
	ac             *apicontroller.PluggableResolver
	authentication *authentication.PluggableResolver
}

func New(restConfig *rest.Config, contentCfg content.Config, appCfg application.Config, informerResyncPeriod time.Duration, featureToggles experimental.FeatureToggles) (*RootResolver, error) {
	uiContainer, err := ui.New(restConfig, informerResyncPeriod)

	makePluggable := module.MakePluggableFunc(uiContainer.BackendModuleInformer, featureToggles.ModulePluggability)

	contentContainer, err := content.New(contentCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Content resolver")
	}
	makePluggable(contentContainer)

	scContainer, err := servicecatalog.New(restConfig, informerResyncPeriod, contentContainer.ContentRetriever)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing ServiceCatalog container")
	}
	makePluggable(scContainer)

	scaContainer, err := servicecatalogaddons.New(restConfig, informerResyncPeriod, scContainer.ServiceCatalogRetriever)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing ServiceCatalog container")
	}
	makePluggable(scaContainer)

	appContainer, err := application.New(restConfig, appCfg, informerResyncPeriod, contentContainer.ContentRetriever)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Application resolver")
	}
	makePluggable(appContainer)

	k8sResolver, err := k8s.New(restConfig, informerResyncPeriod, appContainer.ApplicationRetriever, scContainer.ServiceCatalogRetriever, scaContainer.ServiceCatalogAddonsRetriever)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing K8S resolver")
	}

	kubelessResolver, err := kubeless.New(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Kubeless resolver")
	}
	makePluggable(kubelessResolver)

	acResolver, err := apicontroller.New(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing API controller resolver")
	}
	makePluggable(acResolver)

	authenticationResolver, err := authentication.New(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing authentication resolver")
	}
	makePluggable(authenticationResolver)

	return &RootResolver{
		k8s:            k8sResolver,
		ui:             uiContainer.Resolver,
		sc:             scContainer,
		sca:            scaContainer,
		app:            appContainer,
		content:        contentContainer,
		ac:             acResolver,
		kubeless:       kubelessResolver,
		authentication: authenticationResolver,
	}, nil
}

// WaitForCacheSync waits for caches to populate. This is blocking operation.
func (r *RootResolver) WaitForCacheSync(stopCh <-chan struct{}) {
	// Not pluggable modules
	r.k8s.WaitForCacheSync(stopCh)
	r.ui.WaitForCacheSync(stopCh)

	// Pluggable modules
	r.sc.StopCacheSyncOnClose(stopCh)
	r.sca.StopCacheSyncOnClose(stopCh)
	r.app.StopCacheSyncOnClose(stopCh)
	r.content.StopCacheSyncOnClose(stopCh)
	r.ac.StopCacheSyncOnClose(stopCh)
	r.kubeless.StopCacheSyncOnClose(stopCh)
	r.authentication.StopCacheSyncOnClose(stopCh)
}

func (r *RootResolver) Deployment() gqlschema.DeploymentResolver {
	return &deploymentResolver{r.k8s}
}

func (r *RootResolver) EventActivation() gqlschema.EventActivationResolver {
	return &eventActivationResolver{r.app}
}

func (r *RootResolver) Application() gqlschema.ApplicationResolver {
	return &appResolver{r.app}
}

func (r *RootResolver) ServiceBinding() gqlschema.ServiceBindingResolver {
	return &serviceBindingResolver{r.k8s}
}

func (r *RootResolver) ServiceBindingUsage() gqlschema.ServiceBindingUsageResolver {
	return &serviceBindingUsageResolver{r.sc}
}

func (r *RootResolver) ServiceClass() gqlschema.ServiceClassResolver {
	return &serviceClassResolver{r.sc}
}

func (r *RootResolver) ClusterServiceClass() gqlschema.ClusterServiceClassResolver {
	return &clusterServiceClassResolver{r.sc}
}

func (r *RootResolver) ServiceInstance() gqlschema.ServiceInstanceResolver {
	return &serviceInstanceResolver{r.sc, r.sca}
}

func (r *RootResolver) Environment() gqlschema.EnvironmentResolver {
	return &environmentResolver{r.k8s}
}

func (r *RootResolver) Query() gqlschema.QueryResolver {
	return &queryResolver{r}
}

func (r *RootResolver) Mutation() gqlschema.MutationResolver {
	return &mutationResolver{r}
}

func (r *RootResolver) Subscription() gqlschema.SubscriptionResolver {
	return &subscriptionResolver{r}
}

// Mutations

type mutationResolver struct {
	*RootResolver
}

func (r *mutationResolver) CreateServiceInstance(ctx context.Context, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.CreateServiceInstanceMutation(ctx, params)
}

func (r *mutationResolver) DeleteServiceInstance(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.DeleteServiceInstanceMutation(ctx, name, environment)
}

func (r *mutationResolver) CreateServiceBinding(ctx context.Context, serviceBindingName *string, serviceInstanceName, env string, parameters *gqlschema.JSON) (*gqlschema.CreateServiceBindingOutput, error) {
	return r.sc.Resolver.CreateServiceBindingMutation(ctx, serviceBindingName, serviceInstanceName, env, parameters)
}

func (r *mutationResolver) DeleteServiceBinding(ctx context.Context, serviceBindingName string, env string) (*gqlschema.DeleteServiceBindingOutput, error) {
	return r.sc.Resolver.DeleteServiceBindingMutation(ctx, serviceBindingName, env)
}

func (r *mutationResolver) CreateServiceBindingUsage(ctx context.Context, createServiceBindingUsageInput *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.CreateServiceBindingUsageMutation(ctx, createServiceBindingUsageInput)
}

func (r *mutationResolver) DeleteServiceBindingUsage(ctx context.Context, serviceBindingUsageName string, env string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	return r.sca.Resolver.DeleteServiceBindingUsageMutation(ctx, serviceBindingUsageName, env)
}

func (r *mutationResolver) EnableApplication(ctx context.Context, application string, namespace string) (*gqlschema.ApplicationMapping, error) {
	return r.app.Resolver.EnableApplicationMutation(ctx, application, namespace)
}

func (r *mutationResolver) DisableApplication(ctx context.Context, application string, namespace string) (*gqlschema.ApplicationMapping, error) {
	return r.app.Resolver.DisableApplicationMutation(ctx, application, namespace)
}

func (r *mutationResolver) CreateIDPPreset(ctx context.Context, name string, issuer string, jwksURI string) (*gqlschema.IDPPreset, error) {
	return r.authentication.CreateIDPPresetMutation(ctx, name, issuer, jwksURI)
}

func (r *mutationResolver) DeleteIDPPreset(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	return r.authentication.DeleteIDPPresetMutation(ctx, name)
}

func (r *mutationResolver) CreateApplication(ctx context.Context, name string, description *string, labels *gqlschema.Labels) (gqlschema.ApplicationMutationOutput, error) {
	return r.app.Resolver.CreateApplication(ctx, name, description, labels)
}

func (r *mutationResolver) UpdateApplication(ctx context.Context, name string, description *string, labels *gqlschema.Labels) (gqlschema.ApplicationMutationOutput, error) {
	return r.app.Resolver.UpdateApplication(ctx, name, description, labels)
}

func (r *mutationResolver) DeleteApplication(ctx context.Context, name string) (gqlschema.DeleteApplicationOutput, error) {
	return r.app.Resolver.DeleteApplication(ctx, name)
}

// Queries

type queryResolver struct {
	*RootResolver
}

func (r *queryResolver) Environments(ctx context.Context, application *string) ([]gqlschema.Environment, error) {
	return r.k8s.EnvironmentsQuery(ctx, application)
}

func (r *queryResolver) Deployments(ctx context.Context, environment string, excludeFunctions *bool) ([]gqlschema.Deployment, error) {
	return r.k8s.DeploymentsQuery(ctx, environment, excludeFunctions)
}

func (r *queryResolver) LimitRanges(ctx context.Context, env string) ([]gqlschema.LimitRange, error) {
	return r.k8s.LimitRangesQuery(ctx, env)
}

func (r *queryResolver) ResourceQuotas(ctx context.Context, environment string) ([]gqlschema.ResourceQuota, error) {
	return r.k8s.ResourceQuotasQuery(ctx, environment)
}

func (r *RootResolver) ResourceQuotasStatus(ctx context.Context, environment string) (gqlschema.ResourceQuotasStatus, error) {
	return r.k8s.ResourceQuotasStatus(ctx, environment)
}

func (r *queryResolver) Pod(ctx context.Context, name string, namespace string) (*gqlschema.Pod, error) {
	return r.k8s.PodQuery(ctx, name, namespace)
}

func (r *queryResolver) Pods(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.Pod, error) {
	return r.k8s.PodsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) Functions(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.Function, error) {
	return r.kubeless.FunctionsQuery(ctx, environment, first, offset)
}

func (r *queryResolver) ServiceInstance(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ServiceInstanceQuery(ctx, name, environment)
}

func (r *queryResolver) ServiceInstances(ctx context.Context, environment string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ServiceInstancesQuery(ctx, environment, first, offset, status)
}

func (r *queryResolver) ServiceClasses(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
	return r.sc.Resolver.ServiceClassesQuery(ctx, environment, first, offset)
}

func (r *queryResolver) ServiceClass(ctx context.Context, environment string, name string) (*gqlschema.ServiceClass, error) {
	return r.sc.Resolver.ServiceClassQuery(ctx, name, environment)
}

func (r *queryResolver) ClusterServiceClasses(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceClass, error) {
	return r.sc.Resolver.ClusterServiceClassesQuery(ctx, first, offset)
}

func (r *queryResolver) ClusterServiceClass(ctx context.Context, name string) (*gqlschema.ClusterServiceClass, error) {
	return r.sc.Resolver.ClusterServiceClassQuery(ctx, name)
}

func (r *queryResolver) ServiceBrokers(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceBroker, error) {
	return r.sc.Resolver.ServiceBrokersQuery(ctx, environment, first, offset)
}

func (r *queryResolver) ServiceBroker(ctx context.Context, environment string, name string) (*gqlschema.ServiceBroker, error) {
	return r.sc.Resolver.ServiceBrokerQuery(ctx, environment, name)
}

func (r *queryResolver) ClusterServiceBrokers(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceBroker, error) {
	return r.sc.Resolver.ClusterServiceBrokersQuery(ctx, first, offset)
}

func (r *queryResolver) ClusterServiceBroker(ctx context.Context, name string) (*gqlschema.ClusterServiceBroker, error) {
	return r.sc.Resolver.ClusterServiceBrokerQuery(ctx, name)
}

func (r *queryResolver) UsageKinds(ctx context.Context, first *int, offset *int) ([]gqlschema.UsageKind, error) {
	return r.sca.Resolver.ListUsageKinds(ctx, first, offset)
}

func (r *queryResolver) UsageKindResources(ctx context.Context, usageKind string, environment string) ([]gqlschema.UsageKindResource, error) {
	return r.sca.Resolver.ListServiceUsageKindResources(ctx, usageKind, environment)
}

func (r *queryResolver) BindableResources(ctx context.Context, environment string) ([]gqlschema.BindableResourcesOutputItem, error) {
	return r.sca.Resolver.ListBindableResources(ctx, environment)
}

func (r *queryResolver) ServiceBinding(ctx context.Context, name string, environment string) (*gqlschema.ServiceBinding, error) {
	return r.sc.Resolver.ServiceBindingQuery(ctx, name, environment)
}

func (r *queryResolver) ServiceBindingUsage(ctx context.Context, name, environment string) (*gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.ServiceBindingUsageQuery(ctx, name, environment)
}

func (r *queryResolver) Content(ctx context.Context, contentType, id string) (*gqlschema.JSON, error) {
	return r.content.Resolver.ContentQuery(ctx, contentType, id)
}

func (r *queryResolver) Topics(ctx context.Context, input []gqlschema.InputTopic, internal *bool) ([]gqlschema.TopicEntry, error) {
	return r.content.Resolver.TopicsQuery(ctx, input, internal)
}

func (r *queryResolver) Application(ctx context.Context, name string) (*gqlschema.Application, error) {
	return r.app.Resolver.ApplicationQuery(ctx, name)
}

func (r *queryResolver) Applications(ctx context.Context, namespace *string, first *int, offset *int) ([]gqlschema.Application, error) {
	return r.app.Resolver.ApplicationsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ConnectorService(ctx context.Context, application string) (gqlschema.ConnectorService, error) {
	return r.app.Resolver.ConnectorServiceQuery(ctx, application)
}

func (r *queryResolver) EventActivations(ctx context.Context, namespace string) ([]gqlschema.EventActivation, error) {
	return r.app.Resolver.EventActivationsQuery(ctx, namespace)
}

func (r *queryResolver) Apis(ctx context.Context, environment *string, namespace *string, serviceName *string, hostname *string) ([]gqlschema.API, error) {
	return r.ac.APIsQuery(ctx, environment, namespace, serviceName, hostname)
}

func (r *queryResolver) IDPPreset(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	return r.authentication.IDPPresetQuery(ctx, name)
}

func (r *queryResolver) IDPPresets(ctx context.Context, first *int, offset *int) ([]gqlschema.IDPPreset, error) {
	return r.authentication.IDPPresetsQuery(ctx, first, offset)
}

func (r *queryResolver) BackendModules(ctx context.Context) ([]gqlschema.BackendModule, error) {
	return r.ui.BackendModulesQuery(ctx)
}

// Subscriptions

type subscriptionResolver struct {
	*RootResolver
}

func (r *subscriptionResolver) ServiceInstanceEvent(ctx context.Context, environment string) (<-chan gqlschema.ServiceInstanceEvent, error) {
	return r.sc.Resolver.ServiceInstanceEventSubscription(ctx, environment)
}

func (r *subscriptionResolver) ServiceBindingUsageEvent(ctx context.Context, environment string) (<-chan gqlschema.ServiceBindingUsageEvent, error) {
	return r.sca.Resolver.ServiceBindingUsageEventSubscription(ctx, environment)
}

func (r *subscriptionResolver) ServiceBindingEvent(ctx context.Context, environment string) (<-chan gqlschema.ServiceBindingEvent, error) {
	return r.sc.Resolver.ServiceBindingEventSubscription(ctx, environment)
}

func (r *subscriptionResolver) ServiceBrokerEvent(ctx context.Context, environment string) (<-chan gqlschema.ServiceBrokerEvent, error) {
	return r.sc.Resolver.ServiceBrokerEventSubscription(ctx, environment)
}

func (r *subscriptionResolver) ClusterServiceBrokerEvent(ctx context.Context) (<-chan gqlschema.ClusterServiceBrokerEvent, error) {
	return r.sc.Resolver.ClusterServiceBrokerEventSubscription(ctx)
}

func (r *subscriptionResolver) ApplicationEvent(ctx context.Context) (<-chan gqlschema.ApplicationEvent, error) {
	return r.app.Resolver.ApplicationEventSubscription(ctx)
}

// Service Instance

type serviceInstanceResolver struct {
	sc  *servicecatalog.PluggableContainer
	sca *servicecatalogaddons.PluggableContainer
}

func (r *serviceInstanceResolver) ClusterServicePlan(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServicePlan, error) {
	return r.sc.Resolver.ServiceInstanceClusterServicePlanField(ctx, obj)
}

func (r *serviceInstanceResolver) ClusterServiceClass(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServiceClass, error) {
	return r.sc.Resolver.ServiceInstanceClusterServiceClassField(ctx, obj)
}

func (r *serviceInstanceResolver) ServicePlan(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServicePlan, error) {
	return r.sc.Resolver.ServiceInstanceServicePlanField(ctx, obj)
}

func (r *serviceInstanceResolver) ServiceClass(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error) {
	return r.sc.Resolver.ServiceInstanceServiceClassField(ctx, obj)
}

func (r *serviceInstanceResolver) Bindable(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error) {
	return r.sc.Resolver.ServiceInstanceBindableField(ctx, obj)
}

func (r *serviceInstanceResolver) ServiceBindings(ctx context.Context, obj *gqlschema.ServiceInstance) (gqlschema.ServiceBindings, error) {
	return r.sc.Resolver.ServiceBindingsToInstanceQuery(ctx, obj.Name, obj.Environment)
}

func (r *serviceInstanceResolver) ServiceBindingUsages(ctx context.Context, obj *gqlschema.ServiceInstance) ([]gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.ServiceBindingUsagesOfInstanceQuery(ctx, obj.Name, obj.Environment)
}

// Service Binding

type serviceBindingResolver struct {
	k8s *k8s.Resolver
}

func (r *serviceBindingResolver) Secret(ctx context.Context, serviceBinding *gqlschema.ServiceBinding) (*gqlschema.Secret, error) {
	return r.k8s.SecretQuery(ctx, serviceBinding.SecretName, serviceBinding.Environment)
}

// Service Binding Usage

type serviceBindingUsageResolver struct {
	sc *servicecatalog.PluggableContainer
}

func (r *serviceBindingUsageResolver) ServiceBinding(ctx context.Context, obj *gqlschema.ServiceBindingUsage) (*gqlschema.ServiceBinding, error) {
	return r.sc.Resolver.ServiceBindingQuery(ctx, obj.ServiceBindingName, obj.Environment)
}

// Application

type appResolver struct {
	app *application.PluggableContainer
}

func (r *appResolver) EnabledInEnvironments(ctx context.Context, obj *gqlschema.Application) ([]string, error) {
	return r.app.Resolver.ApplicationEnabledInEnvironmentsField(ctx, obj)
}

func (r *appResolver) Status(ctx context.Context, obj *gqlschema.Application) (gqlschema.ApplicationStatus, error) {
	return r.app.Resolver.ApplicationStatusField(ctx, obj)
}

// Deployment

type deploymentResolver struct {
	k8s *k8s.Resolver
}

func (r *deploymentResolver) BoundServiceInstanceNames(ctx context.Context, deployment *gqlschema.Deployment) ([]string, error) {
	return r.k8s.DeploymentBoundServiceInstanceNamesField(ctx, deployment)
}

// Event Activation

type eventActivationResolver struct {
	app *application.PluggableContainer
}

func (r *eventActivationResolver) Events(ctx context.Context, eventActivation *gqlschema.EventActivation) ([]gqlschema.EventActivationEvent, error) {
	return r.app.Resolver.EventActivationEventsField(ctx, eventActivation)
}

// Service Class

type serviceClassResolver struct {
	sc *servicecatalog.PluggableContainer
}

func (r *serviceClassResolver) Activated(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	return r.sc.Resolver.ServiceClassActivatedField(ctx, obj)
}

func (r *serviceClassResolver) Plans(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error) {
	return r.sc.Resolver.ServiceClassPlansField(ctx, obj)
}

func (r *serviceClassResolver) APISpec(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	return r.sc.Resolver.ServiceClassApiSpecField(ctx, obj)
}

func (r *serviceClassResolver) AsyncAPISpec(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	return r.sc.Resolver.ServiceClassAsyncApiSpecField(ctx, obj)
}

func (r *serviceClassResolver) Content(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	return r.sc.Resolver.ServiceClassContentField(ctx, obj)
}

// Cluster Service Class

type clusterServiceClassResolver struct {
	sc *servicecatalog.PluggableContainer
}

func (r *clusterServiceClassResolver) Activated(ctx context.Context, obj *gqlschema.ClusterServiceClass) (bool, error) {
	return r.sc.Resolver.ClusterServiceClassActivatedField(ctx, obj)
}

func (r *clusterServiceClassResolver) Plans(ctx context.Context, obj *gqlschema.ClusterServiceClass) ([]gqlschema.ClusterServicePlan, error) {
	return r.sc.Resolver.ClusterServiceClassPlansField(ctx, obj)
}

func (r *clusterServiceClassResolver) APISpec(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	return r.sc.Resolver.ClusterServiceClassApiSpecField(ctx, obj)
}

func (r *clusterServiceClassResolver) AsyncAPISpec(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	return r.sc.Resolver.ClusterServiceClassAsyncApiSpecField(ctx, obj)
}

func (r *clusterServiceClassResolver) Content(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	return r.sc.Resolver.ClusterServiceClassContentField(ctx, obj)
}

type environmentResolver struct {
	k8s *k8s.Resolver
}

func (r *environmentResolver) Applications(ctx context.Context, obj *gqlschema.Environment) ([]string, error) {
	return r.k8s.ApplicationsField(ctx, obj)
}
