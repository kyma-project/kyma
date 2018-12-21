package domain

import (
	"context"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/ui"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/experimental"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"
	"time"

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
	k8s       *k8s.Resolver
	kubeless  *kubeless.Resolver
	sc        *servicecatalog.Resolver
	app       *application.Resolver
	content   *content.Resolver
	ac        *apicontroller.Resolver
	idpPreset *authentication.Resolver
}

func New(restConfig *rest.Config, contentCfg content.Config, appCfg application.Config, informerResyncPeriod time.Duration, featureToggles experimental.FeatureToggles) (*RootResolver, error) {

	uiContainer, err := ui.New(restConfig, informerResyncPeriod)


	makePluggable := module.MakePluggableFunc(uiContainer.BackendModuleInformer, featureToggles.ModulePluggability)

	contentContainer, err := content.New(contentCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Content resolver")
	}

	scContainer, err := servicecatalog.New(restConfig, informerResyncPeriod, contentContainer.AsyncApiSpecGetter, contentContainer.ApiSpecGetter, contentContainer.ContentGetter)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing ServiceCatalog container")
	}

	appContainer, err := application.New(restConfig, appCfg, informerResyncPeriod, contentContainer.AsyncApiSpecGetter)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Application resolver")
	}

	k8sResolver, err := k8s.New(restConfig, appContainer.AppLister, informerResyncPeriod, scContainer.ServiceBindingUsageLister, scContainer.ServiceBindingGetter)
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

	idpPresetResolver, err := authentication.New(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing idpPreset resolver")
	}

	return &RootResolver{
		k8s:       k8sResolver,
		kubeless:  kubelessResolver,
		app:       appContainer.Resolver,
		sc:        scContainer.Resolver,
		content:   contentContainer.Resolver,
		ac:        acResolver,
		idpPreset: idpPresetResolver,
	}, nil
}

// WaitForCacheSync waits for caches to populate. This is blocking operation.
func (r *RootResolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.app.WaitForCacheSync(stopCh)
	r.sc.WaitForCacheSync(stopCh)
	r.k8s.WaitForCacheSync(stopCh)
	r.kubeless.WaitForCacheSync(stopCh)
	r.ac.WaitForCacheSync(stopCh)
	r.idpPreset.WaitForCacheSync(stopCh)
	r.content.WaitForCacheSync(stopCh)
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
	return &serviceInstanceResolver{r.sc}
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
	return r.sc.CreateServiceInstanceMutation(ctx, params)
}

func (r *mutationResolver) DeleteServiceInstance(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	return r.sc.DeleteServiceInstanceMutation(ctx, name, environment)
}

func (r *mutationResolver) CreateServiceBinding(ctx context.Context, serviceBindingName *string, serviceInstanceName, env string, parameters *gqlschema.JSON) (*gqlschema.CreateServiceBindingOutput, error) {
	return r.sc.CreateServiceBindingMutation(ctx, serviceBindingName, serviceInstanceName, env, parameters)
}

func (r *mutationResolver) DeleteServiceBinding(ctx context.Context, serviceBindingName string, env string) (*gqlschema.DeleteServiceBindingOutput, error) {
	return r.sc.DeleteServiceBindingMutation(ctx, serviceBindingName, env)
}

func (r *mutationResolver) CreateServiceBindingUsage(ctx context.Context, createServiceBindingUsageInput *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error) {
	return r.sc.CreateServiceBindingUsageMutation(ctx, createServiceBindingUsageInput)
}

func (r *mutationResolver) DeleteServiceBindingUsage(ctx context.Context, serviceBindingUsageName string, env string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	return r.sc.DeleteServiceBindingUsageMutation(ctx, serviceBindingUsageName, env)
}

func (r *mutationResolver) EnableApplication(ctx context.Context, application string, environment string) (*gqlschema.ApplicationMapping, error) {
	return r.app.EnableApplicationMutation(ctx, application, environment)
}

func (r *mutationResolver) DisableApplication(ctx context.Context, application string, environment string) (*gqlschema.ApplicationMapping, error) {
	return r.app.DisableApplicationMutation(ctx, application, environment)
}

func (r *mutationResolver) CreateIDPPreset(ctx context.Context, name string, issuer string, jwksURI string) (*gqlschema.IDPPreset, error) {
	return r.idpPreset.CreateIDPPresetMutation(ctx, name, issuer, jwksURI)
}

func (r *mutationResolver) DeleteIDPPreset(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	return r.idpPreset.DeleteIDPPresetMutation(ctx, name)
}

func (r *mutationResolver) CreateApplication(ctx context.Context, name string, description *string, labels *gqlschema.Labels) (gqlschema.ApplicationMutationOutput, error) {
	return r.app.CreateApplication(ctx, name, description, labels)
}

func (r *mutationResolver) UpdateApplication(ctx context.Context, name string, description *string, labels *gqlschema.Labels) (gqlschema.ApplicationMutationOutput, error) {
	return r.app.UpdateApplication(ctx, name, description, labels)
}

func (r *mutationResolver) DeleteApplication(ctx context.Context, name string) (gqlschema.DeleteApplicationOutput, error) {
	return r.app.DeleteApplication(ctx, name)
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

func (r *queryResolver) Functions(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.Function, error) {
	return r.kubeless.FunctionsQuery(ctx, environment, first, offset)
}

func (r *queryResolver) ServiceInstance(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	return r.sc.ServiceInstanceQuery(ctx, name, environment)
}

func (r *queryResolver) ServiceInstances(ctx context.Context, environment string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error) {
	return r.sc.ServiceInstancesQuery(ctx, environment, first, offset, status)
}

func (r *queryResolver) ServiceClasses(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
	return r.sc.ServiceClassesQuery(ctx, environment, first, offset)
}

func (r *queryResolver) ServiceClass(ctx context.Context, environment string, name string) (*gqlschema.ServiceClass, error) {
	return r.sc.ServiceClassQuery(ctx, name, environment)
}

func (r *queryResolver) ClusterServiceClasses(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceClass, error) {
	return r.sc.ClusterServiceClassesQuery(ctx, first, offset)
}

func (r *queryResolver) ClusterServiceClass(ctx context.Context, name string) (*gqlschema.ClusterServiceClass, error) {
	return r.sc.ClusterServiceClassQuery(ctx, name)
}

func (r *queryResolver) ServiceBrokers(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceBroker, error) {
	return r.sc.ServiceBrokersQuery(ctx, environment, first, offset)
}

func (r *queryResolver) ServiceBroker(ctx context.Context, environment string, name string) (*gqlschema.ServiceBroker, error) {
	return r.sc.ServiceBrokerQuery(ctx, environment, name)
}

func (r *queryResolver) ClusterServiceBrokers(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceBroker, error) {
	return r.sc.ClusterServiceBrokersQuery(ctx, first, offset)
}

func (r *queryResolver) ClusterServiceBroker(ctx context.Context, name string) (*gqlschema.ClusterServiceBroker, error) {
	return r.sc.ClusterServiceBrokerQuery(ctx, name)
}

func (r *queryResolver) UsageKinds(ctx context.Context, first *int, offset *int) ([]gqlschema.UsageKind, error) {
	return r.sc.ListUsageKinds(ctx, first, offset)
}

func (r *queryResolver) UsageKindResources(ctx context.Context, usageKind string, environment string) ([]gqlschema.UsageKindResource, error) {
	return r.sc.ListServiceUsageKindResources(ctx, usageKind, environment)
}

func (r *queryResolver) BindableResources(ctx context.Context, environment string) ([]gqlschema.BindableResourcesOutputItem, error) {
	return r.sc.ListBindableResources(ctx, environment)
}

func (r *queryResolver) ServiceBinding(ctx context.Context, name string, environment string) (*gqlschema.ServiceBinding, error) {
	return r.sc.ServiceBindingQuery(ctx, name, environment)
}

func (r *queryResolver) ServiceBindingUsage(ctx context.Context, name, environment string) (*gqlschema.ServiceBindingUsage, error) {
	return r.sc.ServiceBindingUsageQuery(ctx, name, environment)
}

func (r *queryResolver) Content(ctx context.Context, contentType, id string) (*gqlschema.JSON, error) {
	return r.content.ContentQuery(ctx, contentType, id)
}

func (r *queryResolver) Topics(ctx context.Context, input []gqlschema.InputTopic, internal *bool) ([]gqlschema.TopicEntry, error) {
	return r.content.TopicsQuery(ctx, input, internal)
}

func (r *queryResolver) Application(ctx context.Context, name string) (*gqlschema.Application, error) {
	return r.app.ApplicationQuery(ctx, name)
}

func (r *queryResolver) Applications(ctx context.Context, environment *string, first *int, offset *int) ([]gqlschema.Application, error) {
	return r.app.ApplicationsQuery(ctx, environment, first, offset)
}

func (r *queryResolver) ConnectorService(ctx context.Context, application string) (gqlschema.ConnectorService, error) {
	return r.app.ConnectorServiceQuery(ctx, application)
}

func (r *queryResolver) EventActivations(ctx context.Context, environment string) ([]gqlschema.EventActivation, error) {
	return r.app.EventActivationsQuery(ctx, environment)
}

func (r *queryResolver) Apis(ctx context.Context, environment string, serviceName *string, hostname *string) ([]gqlschema.API, error) {
	return r.ac.APIsQuery(ctx, environment, serviceName, hostname)
}

func (r *queryResolver) IDPPreset(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	return r.idpPreset.IDPPresetQuery(ctx, name)
}

func (r *queryResolver) IDPPresets(ctx context.Context, first *int, offset *int) ([]gqlschema.IDPPreset, error) {
	return r.idpPreset.IDPPresetsQuery(ctx, first, offset)
}

// Subscriptions

type subscriptionResolver struct {
	*RootResolver
}

func (r *subscriptionResolver) ServiceInstanceEvent(ctx context.Context, environment string) (<-chan gqlschema.ServiceInstanceEvent, error) {
	return r.sc.ServiceInstanceEventSubscription(ctx, environment)
}

func (r *subscriptionResolver) ServiceBindingUsageEvent(ctx context.Context, environment string) (<-chan gqlschema.ServiceBindingUsageEvent, error) {
	return r.sc.ServiceBindingUsageEventSubscription(ctx, environment)
}

func (r *subscriptionResolver) ServiceBindingEvent(ctx context.Context, environment string) (<-chan gqlschema.ServiceBindingEvent, error) {
	return r.sc.ServiceBindingEventSubscription(ctx, environment)
}

func (r *subscriptionResolver) ServiceBrokerEvent(ctx context.Context, environment string) (<-chan gqlschema.ServiceBrokerEvent, error) {
	return r.sc.ServiceBrokerEventSubscription(ctx, environment)
}

func (r *subscriptionResolver) ClusterServiceBrokerEvent(ctx context.Context) (<-chan gqlschema.ClusterServiceBrokerEvent, error) {
	return r.sc.ClusterServiceBrokerEventSubscription(ctx)
}

func (r *subscriptionResolver) ApplicationEvent(ctx context.Context) (<-chan gqlschema.ApplicationEvent, error) {
	return r.app.ApplicationEventSubscription(ctx)
}

// Service Instance

type serviceInstanceResolver struct {
	sc *servicecatalog.Resolver
}

func (r *serviceInstanceResolver) ClusterServicePlan(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServicePlan, error) {
	return r.sc.ServiceInstanceClusterServicePlanField(ctx, obj)
}

func (r *serviceInstanceResolver) ClusterServiceClass(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServiceClass, error) {
	return r.sc.ServiceInstanceClusterServiceClassField(ctx, obj)
}

func (r *serviceInstanceResolver) ServicePlan(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServicePlan, error) {
	return r.sc.ServiceInstanceServicePlanField(ctx, obj)
}

func (r *serviceInstanceResolver) ServiceClass(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error) {
	return r.sc.ServiceInstanceServiceClassField(ctx, obj)
}

func (r *serviceInstanceResolver) Bindable(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error) {
	return r.sc.ServiceInstanceBindableField(ctx, obj)
}

func (r *serviceInstanceResolver) ServiceBindings(ctx context.Context, obj *gqlschema.ServiceInstance) (gqlschema.ServiceBindings, error) {
	return r.sc.ServiceBindingsToInstanceQuery(ctx, obj.Name, obj.Environment)
}

func (r *serviceInstanceResolver) ServiceBindingUsages(ctx context.Context, obj *gqlschema.ServiceInstance) ([]gqlschema.ServiceBindingUsage, error) {
	return r.sc.ServiceBindingUsagesOfInstanceQuery(ctx, obj.Name, obj.Environment)
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
	sc *servicecatalog.Resolver
}

func (r *serviceBindingUsageResolver) ServiceBinding(ctx context.Context, obj *gqlschema.ServiceBindingUsage) (*gqlschema.ServiceBinding, error) {
	return r.sc.ServiceBindingQuery(ctx, obj.ServiceBindingName, obj.Environment)
}

// Application

type appResolver struct {
	app *application.Resolver
}

func (r *appResolver) EnabledInEnvironments(ctx context.Context, obj *gqlschema.Application) ([]string, error) {
	return r.app.ApplicationEnabledInEnvironmentsField(ctx, obj)
}

func (r *appResolver) Status(ctx context.Context, obj *gqlschema.Application) (gqlschema.ApplicationStatus, error) {
	return r.app.ApplicationStatusField(ctx, obj)
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
	re *application.Resolver
}

func (r *eventActivationResolver) Events(ctx context.Context, eventActivation *gqlschema.EventActivation) ([]gqlschema.EventActivationEvent, error) {
	return r.re.EventActivationEventsField(ctx, eventActivation)
}

// Service Class

type serviceClassResolver struct {
	sc *servicecatalog.Resolver
}

func (r *serviceClassResolver) Activated(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	return r.sc.ServiceClassActivatedField(ctx, obj)
}

func (r *serviceClassResolver) Plans(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error) {
	return r.sc.ServiceClassPlansField(ctx, obj)
}

func (r *serviceClassResolver) APISpec(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	return r.sc.ServiceClassApiSpecField(ctx, obj)
}

func (r *serviceClassResolver) AsyncAPISpec(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	return r.sc.ServiceClassAsyncApiSpecField(ctx, obj)
}

func (r *serviceClassResolver) Content(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	return r.sc.ServiceClassContentField(ctx, obj)
}

// Cluster Service Class

type clusterServiceClassResolver struct {
	sc *servicecatalog.Resolver
}

func (r *clusterServiceClassResolver) Activated(ctx context.Context, obj *gqlschema.ClusterServiceClass) (bool, error) {
	return r.sc.ClusterServiceClassActivatedField(ctx, obj)
}

func (r *clusterServiceClassResolver) Plans(ctx context.Context, obj *gqlschema.ClusterServiceClass) ([]gqlschema.ClusterServicePlan, error) {
	return r.sc.ClusterServiceClassPlansField(ctx, obj)
}

func (r *clusterServiceClassResolver) APISpec(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	return r.sc.ClusterServiceClassApiSpecField(ctx, obj)
}

func (r *clusterServiceClassResolver) AsyncAPISpec(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	return r.sc.ClusterServiceClassAsyncApiSpecField(ctx, obj)
}

func (r *clusterServiceClassResolver) Content(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	return r.sc.ClusterServiceClassContentField(ctx, obj)
}
