package domain

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/experimental"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/authentication"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/kubeless"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type RootResolver struct {
	ui  *ui.Resolver
	k8s *k8s.Resolver

	sc             *servicecatalog.PluggableContainer
	sca            *servicecatalogaddons.PluggableContainer
	app            *application.PluggableContainer
	assetstore     *assetstore.PluggableContainer
	cms            *cms.PluggableContainer
	kubeless       *kubeless.PluggableResolver
	ac             *apicontroller.PluggableResolver
	authentication *authentication.PluggableResolver
}

func New(restConfig *rest.Config, appCfg application.Config, assetstoreCfg assetstore.Config, informerResyncPeriod time.Duration, featureToggles experimental.FeatureToggles) (*RootResolver, error) {
	uiContainer, err := ui.New(restConfig, informerResyncPeriod)
	makePluggable := module.MakePluggableFunc(uiContainer.BackendModuleInformer)

	assetStoreContainer, err := assetstore.New(restConfig, assetstoreCfg, informerResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing AssetStore resolver")
	}
	makePluggable(assetStoreContainer)

	cmsContainer, err := cms.New(restConfig, informerResyncPeriod, assetStoreContainer.AssetStoreRetriever)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing CMS resolver")
	}
	makePluggable(cmsContainer)

	scContainer, err := servicecatalog.New(restConfig, informerResyncPeriod, cmsContainer.CmsRetriever)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing ServiceCatalog container")
	}
	makePluggable(scContainer)

	scaContainer, err := servicecatalogaddons.New(restConfig, informerResyncPeriod, scContainer.ServiceCatalogRetriever)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing ServiceCatalog container")
	}
	makePluggable(scaContainer)

	appContainer, err := application.New(restConfig, appCfg, informerResyncPeriod, assetStoreContainer.AssetStoreRetriever)
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
		assetstore:     assetStoreContainer,
		cms:            cmsContainer,
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
	r.cms.StopCacheSyncOnClose(stopCh)
	r.assetstore.StopCacheSyncOnClose(stopCh)
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

func (r *RootResolver) ClusterDocsTopic() gqlschema.ClusterDocsTopicResolver {
	return &clusterDocsTopicResolver{r.cms}
}

func (r *RootResolver) DocsTopic() gqlschema.DocsTopicResolver {
	return &docsTopicResolver{r.cms}
}

func (r *RootResolver) ClusterAsset() gqlschema.ClusterAssetResolver {
	return &clusterAssetResolver{r.assetstore}
}

func (r *RootResolver) Asset() gqlschema.AssetResolver {
	return &assetResolver{r.assetstore}
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

func (r *RootResolver) Namespace() gqlschema.NamespaceResolver {
	return &namespaceResolver{r.k8s}
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

func (r *mutationResolver) CreateResource(ctx context.Context, namespace string, resource gqlschema.JSON) (*gqlschema.JSON, error) {
	return r.k8s.CreateResourceMutation(ctx, namespace, resource)
}

func (r *mutationResolver) DeleteService(ctx context.Context, name string, namespace string) (*gqlschema.Service, error) {
	return r.k8s.DeleteServiceMutation(ctx, name, namespace)
}

func (r *mutationResolver) UpdateService(ctx context.Context, name string, namespace string, service gqlschema.JSON) (*gqlschema.Service, error) {
	return r.k8s.UpdateServiceMutation(ctx, name, namespace, service)
}

func (r *mutationResolver) UpdatePod(ctx context.Context, name string, namespace string, update gqlschema.JSON) (*gqlschema.Pod, error) {
	return r.k8s.UpdatePodMutation(ctx, name, namespace, update)
}

func (r *mutationResolver) DeletePod(ctx context.Context, name string, namespace string) (*gqlschema.Pod, error) {
	return r.k8s.DeletePodMutation(ctx, name, namespace)
}

func (r *mutationResolver) UpdateSecret(ctx context.Context, name string, namespace string, update gqlschema.JSON) (*gqlschema.Secret, error) {
	return r.k8s.UpdateSecretMutation(ctx, name, namespace, update)
}

func (r *mutationResolver) DeleteSecret(ctx context.Context, name string, namespace string) (*gqlschema.Secret, error) {
	return r.k8s.DeleteSecretMutation(ctx, name, namespace)
}

func (r *mutationResolver) UpdateReplicaSet(ctx context.Context, name string, namespace string, update gqlschema.JSON) (*gqlschema.ReplicaSet, error) {
	return r.k8s.UpdateReplicaSetMutation(ctx, name, namespace, update)
}

func (r *mutationResolver) DeleteReplicaSet(ctx context.Context, name string, namespace string) (*gqlschema.ReplicaSet, error) {
	return r.k8s.DeleteReplicaSetMutation(ctx, name, namespace)
}

func (r *mutationResolver) UpdateConfigMap(ctx context.Context, name string, namespace string, update gqlschema.JSON) (*gqlschema.ConfigMap, error) {
	return r.k8s.UpdateConfigMapMutation(ctx, name, namespace, update)
}

func (r *mutationResolver) DeleteConfigMap(ctx context.Context, name string, namespace string) (*gqlschema.ConfigMap, error) {
	return r.k8s.DeleteConfigMapMutation(ctx, name, namespace)
}

func (r *mutationResolver) CreateServiceInstance(ctx context.Context, namespace string, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.CreateServiceInstanceMutation(ctx, namespace, params)
}

func (r *mutationResolver) DeleteServiceInstance(ctx context.Context, name string, namespace string) (*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.DeleteServiceInstanceMutation(ctx, name, namespace)
}

func (r *mutationResolver) CreateServiceBinding(ctx context.Context, serviceBindingName *string, serviceInstanceName, ns string, parameters *gqlschema.JSON) (*gqlschema.CreateServiceBindingOutput, error) {
	return r.sc.Resolver.CreateServiceBindingMutation(ctx, serviceBindingName, serviceInstanceName, ns, parameters)
}

func (r *mutationResolver) DeleteServiceBinding(ctx context.Context, serviceBindingName string, ns string) (*gqlschema.DeleteServiceBindingOutput, error) {
	return r.sc.Resolver.DeleteServiceBindingMutation(ctx, serviceBindingName, ns)
}

func (r *mutationResolver) CreateServiceBindingUsage(ctx context.Context, namespace string, createServiceBindingUsageInput *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.CreateServiceBindingUsageMutation(ctx, namespace, createServiceBindingUsageInput)
}

func (r *mutationResolver) DeleteServiceBindingUsage(ctx context.Context, serviceBindingUsageName string, ns string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	return r.sca.Resolver.DeleteServiceBindingUsageMutation(ctx, serviceBindingUsageName, ns)
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

func (r *mutationResolver) CreateAddonsConfiguration(ctx context.Context, name string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.CreateAddonsConfiguration(ctx, name, urls, labels)
}

func (r *mutationResolver) UpdateAddonsConfiguration(ctx context.Context, name string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.UpdateAddonsConfiguration(ctx, name, urls, labels)
}

func (r *mutationResolver) DeleteAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.DeleteAddonsConfiguration(ctx, name)
}

func (r *mutationResolver) AddAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.AddAddonsConfigurationURLs(ctx, name, urls)
}

func (r *mutationResolver) RemoveAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.RemoveAddonsConfigurationURLs(ctx, name, urls)
}

// Queries

type queryResolver struct {
	*RootResolver
}

func (r *queryResolver) Namespaces(ctx context.Context, application *string) ([]gqlschema.Namespace, error) {
	return r.k8s.NamespacesQuery(ctx, application)
}

func (r *queryResolver) Deployments(ctx context.Context, namespace string, excludeFunctions *bool) ([]gqlschema.Deployment, error) {
	return r.k8s.DeploymentsQuery(ctx, namespace, excludeFunctions)
}

func (r *queryResolver) LimitRanges(ctx context.Context, ns string) ([]gqlschema.LimitRange, error) {
	return r.k8s.LimitRangesQuery(ctx, ns)
}

func (r *queryResolver) ResourceQuotas(ctx context.Context, namespace string) ([]gqlschema.ResourceQuota, error) {
	return r.k8s.ResourceQuotasQuery(ctx, namespace)
}

func (r *RootResolver) ResourceQuotasStatus(ctx context.Context, namespace string) (gqlschema.ResourceQuotasStatus, error) {
	return r.k8s.ResourceQuotasStatus(ctx, namespace)
}

func (r *queryResolver) Service(ctx context.Context, name string, namespace string) (*gqlschema.Service, error) {
	return r.k8s.ServiceQuery(ctx, name, namespace)
}

func (r *queryResolver) Services(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.Service, error) {
	return r.k8s.ServicesQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) Pod(ctx context.Context, name string, namespace string) (*gqlschema.Pod, error) {
	return r.k8s.PodQuery(ctx, name, namespace)
}

func (r *queryResolver) Pods(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.Pod, error) {
	return r.k8s.PodsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ReplicaSet(ctx context.Context, name string, namespace string) (*gqlschema.ReplicaSet, error) {
	return r.k8s.ReplicaSetQuery(ctx, name, namespace)
}

func (r *queryResolver) ReplicaSets(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ReplicaSet, error) {
	return r.k8s.ReplicaSetsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ConfigMap(ctx context.Context, name string, namespace string) (*gqlschema.ConfigMap, error) {
	return r.k8s.ConfigMapQuery(ctx, name, namespace)
}

func (r *queryResolver) ConfigMaps(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ConfigMap, error) {
	return r.k8s.ConfigMapsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) Functions(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.Function, error) {
	return r.kubeless.FunctionsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ServiceInstance(ctx context.Context, name string, namespace string) (*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ServiceInstanceQuery(ctx, name, namespace)
}

func (r *queryResolver) ServiceInstances(ctx context.Context, namespace string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ServiceInstancesQuery(ctx, namespace, first, offset, status)
}

func (r *queryResolver) ServiceClasses(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
	return r.sc.Resolver.ServiceClassesQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ServiceClass(ctx context.Context, namespace string, name string) (*gqlschema.ServiceClass, error) {
	return r.sc.Resolver.ServiceClassQuery(ctx, name, namespace)
}

func (r *queryResolver) ClusterServiceClasses(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceClass, error) {
	return r.sc.Resolver.ClusterServiceClassesQuery(ctx, first, offset)
}

func (r *queryResolver) ClusterServiceClass(ctx context.Context, name string) (*gqlschema.ClusterServiceClass, error) {
	return r.sc.Resolver.ClusterServiceClassQuery(ctx, name)
}

func (r *queryResolver) ServiceBrokers(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ServiceBroker, error) {
	return r.sc.Resolver.ServiceBrokersQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ServiceBroker(ctx context.Context, namespace string, name string) (*gqlschema.ServiceBroker, error) {
	return r.sc.Resolver.ServiceBrokerQuery(ctx, namespace, name)
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

func (r *queryResolver) AddonsConfigurations(ctx context.Context, first *int, offset *int) ([]gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.AddonsConfigurationsQuery(ctx, first, offset)
}

func (r *queryResolver) BindableResources(ctx context.Context, namespace string) ([]gqlschema.BindableResourcesOutputItem, error) {
	return r.sca.Resolver.ListBindableResources(ctx, namespace)
}

func (r *queryResolver) ServiceBinding(ctx context.Context, name string, namespace string) (*gqlschema.ServiceBinding, error) {
	return r.sc.Resolver.ServiceBindingQuery(ctx, name, namespace)
}

func (r *queryResolver) ServiceBindingUsage(ctx context.Context, name, namespace string) (*gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.ServiceBindingUsageQuery(ctx, name, namespace)
}

func (r *queryResolver) ClusterDocsTopics(ctx context.Context, viewContext *string, groupName *string) ([]gqlschema.ClusterDocsTopic, error) {
	return r.cms.Resolver.ClusterDocsTopicsQuery(ctx, viewContext, groupName)
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

func (r *queryResolver) Apis(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]gqlschema.API, error) {
	return r.ac.APIsQuery(ctx, namespace, serviceName, hostname)
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

func (r *queryResolver) Secret(ctx context.Context, name, namespace string) (*gqlschema.Secret, error) {
	return r.k8s.SecretQuery(ctx, name, namespace)
}

func (r *queryResolver) Secrets(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.Secret, error) {
	return r.k8s.SecretsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) MicroFrontends(ctx context.Context, namespace string) ([]gqlschema.MicroFrontend, error) {
	return r.ui.MicroFrontendsQuery(ctx, namespace)
}

func (r *queryResolver) ClusterMicroFrontends(ctx context.Context) ([]gqlschema.ClusterMicroFrontend, error) {
	return r.ui.ClusterMicroFrontendsQuery(ctx)
}

func (r *queryResolver) SelfSubjectRules(ctx context.Context, namespace *string) ([]gqlschema.ResourceRule, error) {
	return r.k8s.SelfSubjectRulesQuery(ctx, namespace)
}

// Subscriptions

type subscriptionResolver struct {
	*RootResolver
}

func (r *subscriptionResolver) ClusterAssetEvent(ctx context.Context) (<-chan gqlschema.ClusterAssetEvent, error) {
	return r.assetstore.Resolver.ClusterAssetEventSubscription(ctx)
}

func (r *subscriptionResolver) AssetEvent(ctx context.Context, namespace string) (<-chan gqlschema.AssetEvent, error) {
	return r.assetstore.Resolver.AssetEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ClusterDocsTopicEvent(ctx context.Context) (<-chan gqlschema.ClusterDocsTopicEvent, error) {
	return r.cms.Resolver.ClusterDocsTopicEventSubscription(ctx)
}

func (r *subscriptionResolver) DocsTopicEvent(ctx context.Context, namespace string) (<-chan gqlschema.DocsTopicEvent, error) {
	return r.cms.Resolver.DocsTopicEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ServiceInstanceEvent(ctx context.Context, namespace string) (<-chan gqlschema.ServiceInstanceEvent, error) {
	return r.sc.Resolver.ServiceInstanceEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ServiceBindingUsageEvent(ctx context.Context, namespace string) (<-chan gqlschema.ServiceBindingUsageEvent, error) {
	return r.sca.Resolver.ServiceBindingUsageEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ServiceBindingEvent(ctx context.Context, namespace string) (<-chan gqlschema.ServiceBindingEvent, error) {
	return r.sc.Resolver.ServiceBindingEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ServiceBrokerEvent(ctx context.Context, namespace string) (<-chan gqlschema.ServiceBrokerEvent, error) {
	return r.sc.Resolver.ServiceBrokerEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ClusterServiceBrokerEvent(ctx context.Context) (<-chan gqlschema.ClusterServiceBrokerEvent, error) {
	return r.sc.Resolver.ClusterServiceBrokerEventSubscription(ctx)
}

func (r *subscriptionResolver) ApplicationEvent(ctx context.Context) (<-chan gqlschema.ApplicationEvent, error) {
	return r.app.Resolver.ApplicationEventSubscription(ctx)
}

func (r *subscriptionResolver) PodEvent(ctx context.Context, namespace string) (<-chan gqlschema.PodEvent, error) {
	return r.k8s.PodEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ServiceEvent(ctx context.Context, namespace string) (<-chan gqlschema.ServiceEvent, error) {
	return r.k8s.ServiceEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ConfigMapEvent(ctx context.Context, namespace string) (<-chan gqlschema.ConfigMapEvent, error) {
	return r.k8s.ConfigMapEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) SecretEvent(ctx context.Context, namespace string) (<-chan gqlschema.SecretEvent, error) {
	return r.k8s.SecretEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) AddonsConfigurationEvent(ctx context.Context) (<-chan gqlschema.AddonsConfigurationEvent, error) {
	return r.sca.Resolver.AddonsConfigurationEventSubscription(ctx)
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

func (r *serviceInstanceResolver) ServiceBindings(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceBindings, error) {
	return r.sc.Resolver.ServiceBindingsToInstanceQuery(ctx, obj.Name, obj.Namespace)
}

func (r *serviceInstanceResolver) ServiceBindingUsages(ctx context.Context, obj *gqlschema.ServiceInstance) ([]gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.ServiceBindingUsagesOfInstanceQuery(ctx, obj.Name, obj.Namespace)
}

// Service Binding

type serviceBindingResolver struct {
	k8s *k8s.Resolver
}

func (r *serviceBindingResolver) Secret(ctx context.Context, serviceBinding *gqlschema.ServiceBinding) (*gqlschema.Secret, error) {
	return r.k8s.SecretQuery(ctx, serviceBinding.SecretName, serviceBinding.Namespace)
}

// Service Binding Usage

type serviceBindingUsageResolver struct {
	sc *servicecatalog.PluggableContainer
}

func (r *serviceBindingUsageResolver) ServiceBinding(ctx context.Context, obj *gqlschema.ServiceBindingUsage) (*gqlschema.ServiceBinding, error) {
	return r.sc.Resolver.ServiceBindingQuery(ctx, obj.ServiceBindingName, obj.Namespace)
}

// Application

type appResolver struct {
	app *application.PluggableContainer
}

func (r *appResolver) EnabledInNamespaces(ctx context.Context, obj *gqlschema.Application) ([]string, error) {
	return r.app.Resolver.ApplicationEnabledInNamespacesField(ctx, obj)
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

func (r *serviceClassResolver) Instances(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ServiceClassInstancesField(ctx, obj)
}

func (r *serviceClassResolver) Activated(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	return r.sc.Resolver.ServiceClassActivatedField(ctx, obj)
}

func (r *serviceClassResolver) Plans(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error) {
	return r.sc.Resolver.ServiceClassPlansField(ctx, obj)
}

func (r *serviceClassResolver) ClusterDocsTopic(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.ClusterDocsTopic, error) {
	return r.sc.Resolver.ServiceClassClusterDocsTopicField(ctx, obj)
}

func (r *serviceClassResolver) DocsTopic(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.DocsTopic, error) {
	return r.sc.Resolver.ServiceClassDocsTopicField(ctx, obj)
}

// Cluster Service Class

type clusterServiceClassResolver struct {
	sc *servicecatalog.PluggableContainer
}

func (r *clusterServiceClassResolver) Instances(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) ([]gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ClusterServiceClassInstancesField(ctx, obj, namespace)
}

func (r *clusterServiceClassResolver) Activated(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) (bool, error) {
	return r.sc.Resolver.ClusterServiceClassActivatedField(ctx, obj, namespace)
}

func (r *clusterServiceClassResolver) Plans(ctx context.Context, obj *gqlschema.ClusterServiceClass) ([]gqlschema.ClusterServicePlan, error) {
	return r.sc.Resolver.ClusterServiceClassPlansField(ctx, obj)
}

func (r *clusterServiceClassResolver) ClusterDocsTopic(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.ClusterDocsTopic, error) {
	return r.sc.Resolver.ClusterServiceClassClusterDocsTopicField(ctx, obj)
}

// Namespace

type namespaceResolver struct {
	k8s *k8s.Resolver
}

func (r *namespaceResolver) Applications(ctx context.Context, obj *gqlschema.Namespace) ([]string, error) {
	return r.k8s.ApplicationsField(ctx, obj)
}

// CMS

type clusterDocsTopicResolver struct {
	cms *cms.PluggableContainer
}

func (r *clusterDocsTopicResolver) Assets(ctx context.Context, obj *gqlschema.ClusterDocsTopic, types []string) ([]gqlschema.ClusterAsset, error) {
	return r.cms.Resolver.ClusterDocsTopicAssetsField(ctx, obj, types)
}

type docsTopicResolver struct {
	cms *cms.PluggableContainer
}

func (r *docsTopicResolver) Assets(ctx context.Context, obj *gqlschema.DocsTopic, types []string) ([]gqlschema.Asset, error) {
	return r.cms.Resolver.DocsTopicAssetsField(ctx, obj, types)
}

// Asset Store

type clusterAssetResolver struct {
	assetstore *assetstore.PluggableContainer
}

func (r *clusterAssetResolver) Files(ctx context.Context, obj *gqlschema.ClusterAsset, filterExtensions []string) ([]gqlschema.File, error) {
	return r.assetstore.Resolver.ClusterAssetFilesField(ctx, obj, filterExtensions)
}

type assetResolver struct {
	assetstore *assetstore.PluggableContainer
}

func (r *assetResolver) Files(ctx context.Context, obj *gqlschema.Asset, filterExtensions []string) ([]gqlschema.File, error) {
	return r.assetstore.Resolver.AssetFilesField(ctx, obj, filterExtensions)
}
