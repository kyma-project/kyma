package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

func (r *applicationResolver) EnabledInNamespaces(ctx context.Context, obj *gqlschema.Application) ([]string, error) {
	return r.app.Resolver.ApplicationEnabledInNamespacesField(ctx, obj)
}

func (r *applicationResolver) EnabledMappingServices(ctx context.Context, obj *gqlschema.Application) ([]*gqlschema.EnabledMappingService, error) {
	return r.app.Resolver.ApplicationEnabledMappingServices(ctx, obj)
}

func (r *applicationResolver) Status(ctx context.Context, obj *gqlschema.Application) (gqlschema.ApplicationStatus, error) {
	return r.app.Resolver.ApplicationStatusField(ctx, obj)
}

func (r *assetResolver) Files(ctx context.Context, obj *gqlschema.Asset, filterExtensions []string) ([]*gqlschema.File, error) {
	return r.rafter.Resolver.AssetFilesField(ctx, obj, filterExtensions)
}

func (r *assetGroupResolver) Assets(ctx context.Context, obj *gqlschema.AssetGroup, types []string) ([]*gqlschema.Asset, error) {
	return r.rafter.Resolver.AssetGroupAssetsField(ctx, obj, types)
}

func (r *clusterAssetResolver) Files(ctx context.Context, obj *gqlschema.ClusterAsset, filterExtensions []string) ([]*gqlschema.File, error) {
	return r.rafter.Resolver.ClusterAssetFilesField(ctx, obj, filterExtensions)
}

func (r *clusterAssetGroupResolver) Assets(ctx context.Context, obj *gqlschema.ClusterAssetGroup, types []string) ([]*gqlschema.ClusterAsset, error) {
	return r.rafter.Resolver.ClusterAssetGroupAssetsField(ctx, obj, types)
}

func (r *clusterServiceClassResolver) Plans(ctx context.Context, obj *gqlschema.ClusterServiceClass) ([]*gqlschema.ClusterServicePlan, error) {
	return r.sc.Resolver.ClusterServiceClassPlansField(ctx, obj)
}

func (r *clusterServiceClassResolver) Activated(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) (bool, error) {
	return r.sc.Resolver.ClusterServiceClassActivatedField(ctx, obj, namespace)
}

func (r *clusterServiceClassResolver) Instances(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) ([]*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ClusterServiceClassInstancesField(ctx, obj, namespace)
}

func (r *clusterServiceClassResolver) ClusterAssetGroup(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.ClusterAssetGroup, error) {
	return r.sc.Resolver.ClusterServiceClassClusterAssetGroupField(ctx, obj)
}

func (r *clusterServicePlanResolver) ClusterAssetGroup(ctx context.Context, obj *gqlschema.ClusterServicePlan) (*gqlschema.ClusterAssetGroup, error) {
	return r.sc.Resolver.ClusterServicePlanClusterAssetGroupField(ctx, obj)
}

func (r *deploymentResolver) BoundServiceInstanceNames(ctx context.Context, obj *gqlschema.Deployment) ([]string, error) {
	return r.k8s.DeploymentBoundServiceInstanceNamesField(ctx, obj)
}

func (r *eventActivationResolver) Events(ctx context.Context, obj *gqlschema.EventActivation) ([]*gqlschema.EventActivationEvent, error) {
	return r.app.Resolver.EventActivationEventsField(ctx, obj)
}

func (r *mutationResolver) CreateResource(ctx context.Context, namespace string, resource gqlschema.JSON) (gqlschema.JSON, error) {
	return r.k8s.CreateResourceMutation(ctx, namespace, resource)
}

func (r *mutationResolver) CreateServiceInstance(ctx context.Context, namespace string, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.CreateServiceInstanceMutation(ctx, namespace, params)
}

func (r *mutationResolver) DeleteServiceInstance(ctx context.Context, name string, namespace string) (*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.DeleteServiceInstanceMutation(ctx, name, namespace)
}

func (r *mutationResolver) CreateServiceBinding(ctx context.Context, serviceBindingName *string, serviceInstanceName string, namespace string, parameters gqlschema.JSON) (*gqlschema.CreateServiceBindingOutput, error) {
	return r.sc.Resolver.CreateServiceBindingMutation(ctx, serviceBindingName, serviceInstanceName, namespace, parameters)
}

func (r *mutationResolver) DeleteServiceBinding(ctx context.Context, serviceBindingName string, namespace string) (*gqlschema.DeleteServiceBindingOutput, error) {
	return r.sc.Resolver.DeleteServiceBindingMutation(ctx, serviceBindingName, namespace)
}

func (r *mutationResolver) CreateServiceBindingUsage(ctx context.Context, namespace string, createServiceBindingUsageInput *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.CreateServiceBindingUsageMutation(ctx, namespace, createServiceBindingUsageInput)
}

func (r *mutationResolver) DeleteServiceBindingUsage(ctx context.Context, serviceBindingUsageName string, namespace string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	return r.sca.Resolver.DeleteServiceBindingUsageMutation(ctx, serviceBindingUsageName, namespace)
}

func (r *mutationResolver) DeleteServiceBindingUsages(ctx context.Context, serviceBindingUsageNames []string, namespace string) ([]*gqlschema.DeleteServiceBindingUsageOutput, error) {
	return r.sca.Resolver.DeleteServiceBindingUsagesMutation(ctx, serviceBindingUsageNames, namespace)
}

func (r *mutationResolver) CreateClusterAddonsConfiguration(ctx context.Context, name string, repositories []*gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.CreateClusterAddonsConfiguration(ctx, name, repositories, urls, labels)
}

func (r *mutationResolver) UpdateClusterAddonsConfiguration(ctx context.Context, name string, repositories []*gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.UpdateClusterAddonsConfiguration(ctx, name, repositories, urls, labels)
}

func (r *mutationResolver) DeleteClusterAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.DeleteClusterAddonsConfiguration(ctx, name)
}

func (r *mutationResolver) AddClusterAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.AddClusterAddonsConfigurationURLs(ctx, name, urls)
}

func (r *mutationResolver) RemoveClusterAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.RemoveClusterAddonsConfigurationURLs(ctx, name, urls)
}

func (r *mutationResolver) AddClusterAddonsConfigurationRepository(ctx context.Context, name string, repositories []*gqlschema.AddonsConfigurationRepositoryInput) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.AddClusterAddonsConfigurationRepositories(ctx, name, repositories)
}

func (r *mutationResolver) RemoveClusterAddonsConfigurationRepository(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.RemoveClusterAddonsConfigurationRepositories(ctx, name, urls)
}

func (r *mutationResolver) ResyncClusterAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.ResyncClusterAddonsConfiguration(ctx, name)
}

func (r *mutationResolver) CreateAddonsConfiguration(ctx context.Context, name string, namespace string, repositories []*gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.CreateAddonsConfiguration(ctx, name, namespace, repositories, urls, labels)
}

func (r *mutationResolver) UpdateAddonsConfiguration(ctx context.Context, name string, namespace string, repositories []*gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels gqlschema.Labels) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.UpdateAddonsConfiguration(ctx, name, namespace, repositories, urls, labels)
}

func (r *mutationResolver) DeleteAddonsConfiguration(ctx context.Context, name string, namespace string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.DeleteAddonsConfiguration(ctx, name, namespace)
}

func (r *mutationResolver) AddAddonsConfigurationURLs(ctx context.Context, name string, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.AddAddonsConfigurationURLs(ctx, name, namespace, urls)
}

func (r *mutationResolver) RemoveAddonsConfigurationURLs(ctx context.Context, name string, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.RemoveAddonsConfigurationURLs(ctx, name, namespace, urls)
}

func (r *mutationResolver) AddAddonsConfigurationRepository(ctx context.Context, name string, namespace string, repositories []*gqlschema.AddonsConfigurationRepositoryInput) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.AddAddonsConfigurationRepositories(ctx, name, namespace, repositories)
}

func (r *mutationResolver) RemoveAddonsConfigurationRepository(ctx context.Context, name string, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.RemoveAddonsConfigurationRepositories(ctx, name, namespace, urls)
}

func (r *mutationResolver) ResyncAddonsConfiguration(ctx context.Context, name string, namespace string) (*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.ResyncAddonsConfiguration(ctx, name, namespace)
}

func (r *mutationResolver) CreateApplication(ctx context.Context, name string, description *string, labels gqlschema.Labels) (*gqlschema.ApplicationMutationOutput, error) {
	return r.app.Resolver.CreateApplication(ctx, name, description, labels)
}

func (r *mutationResolver) UpdateApplication(ctx context.Context, name string, description *string, labels gqlschema.Labels) (*gqlschema.ApplicationMutationOutput, error) {
	return r.app.Resolver.UpdateApplication(ctx, name, description, labels)
}

func (r *mutationResolver) DeleteApplication(ctx context.Context, name string) (*gqlschema.DeleteApplicationOutput, error) {
	return r.app.Resolver.DeleteApplication(ctx, name)
}

func (r *mutationResolver) EnableApplication(ctx context.Context, application string, namespace string, allServices *bool, services []*gqlschema.ApplicationMappingService) (*gqlschema.ApplicationMapping, error) {
	return r.app.Resolver.EnableApplicationMutation(ctx, application, namespace, allServices, services)
}

func (r *mutationResolver) OverloadApplication(ctx context.Context, application string, namespace string, allServices *bool, services []*gqlschema.ApplicationMappingService) (*gqlschema.ApplicationMapping, error) {
	return r.app.Resolver.OverloadApplicationMutation(ctx, application, namespace, allServices, services)
}

func (r *mutationResolver) DisableApplication(ctx context.Context, application string, namespace string) (*gqlschema.ApplicationMapping, error) {
	return r.app.Resolver.DisableApplicationMutation(ctx, application, namespace)
}

func (r *mutationResolver) UpdatePod(ctx context.Context, name string, namespace string, pod gqlschema.JSON) (*gqlschema.Pod, error) {
	return r.k8s.UpdatePodMutation(ctx, name, namespace, pod)
}

func (r *mutationResolver) DeletePod(ctx context.Context, name string, namespace string) (*gqlschema.Pod, error) {
	return r.k8s.DeletePodMutation(ctx, name, namespace)
}

func (r *mutationResolver) UpdateSecret(ctx context.Context, name string, namespace string, secret gqlschema.JSON) (*gqlschema.Secret, error) {
	return r.k8s.UpdateSecretMutation(ctx, name, namespace, secret)
}

func (r *mutationResolver) DeleteSecret(ctx context.Context, name string, namespace string) (*gqlschema.Secret, error) {
	return r.k8s.DeleteSecretMutation(ctx, name, namespace)
}

func (r *mutationResolver) UpdateReplicaSet(ctx context.Context, name string, namespace string, replicaSet gqlschema.JSON) (*gqlschema.ReplicaSet, error) {
	return r.k8s.UpdateReplicaSetMutation(ctx, name, namespace, replicaSet)
}

func (r *mutationResolver) DeleteReplicaSet(ctx context.Context, name string, namespace string) (*gqlschema.ReplicaSet, error) {
	return r.k8s.DeleteReplicaSetMutation(ctx, name, namespace)
}

func (r *mutationResolver) UpdateConfigMap(ctx context.Context, name string, namespace string, configMap gqlschema.JSON) (*gqlschema.ConfigMap, error) {
	return r.k8s.UpdateConfigMapMutation(ctx, name, namespace, configMap)
}

func (r *mutationResolver) DeleteConfigMap(ctx context.Context, name string, namespace string) (*gqlschema.ConfigMap, error) {
	return r.k8s.DeleteConfigMapMutation(ctx, name, namespace)
}

func (r *mutationResolver) UpdateService(ctx context.Context, name string, namespace string, service gqlschema.JSON) (*gqlschema.Service, error) {
	return r.k8s.UpdateServiceMutation(ctx, name, namespace, service)
}

func (r *mutationResolver) DeleteService(ctx context.Context, name string, namespace string) (*gqlschema.Service, error) {
	return r.k8s.DeleteServiceMutation(ctx, name, namespace)
}

func (r *mutationResolver) CreateNamespace(ctx context.Context, name string, labels gqlschema.Labels) (*gqlschema.NamespaceMutationOutput, error) {
	return r.k8s.CreateNamespace(ctx, name, labels)
}

func (r *mutationResolver) UpdateNamespace(ctx context.Context, name string, labels gqlschema.Labels) (*gqlschema.NamespaceMutationOutput, error) {
	return r.k8s.UpdateNamespace(ctx, name, labels)
}

func (r *mutationResolver) DeleteNamespace(ctx context.Context, name string) (*gqlschema.Namespace, error) {
	return r.k8s.DeleteNamespace(ctx, name)
}

func (r *mutationResolver) CreateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionMutationInput) (*gqlschema.Function, error) {
	return r.serverless.Resolver.CreateFunction(ctx, name, namespace, params)
}

func (r *mutationResolver) UpdateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionMutationInput) (*gqlschema.Function, error) {
	return r.serverless.Resolver.UpdateFunction(ctx, name, namespace, params)
}

func (r *mutationResolver) DeleteFunction(ctx context.Context, namespace string, function gqlschema.FunctionMetadataInput) (*gqlschema.FunctionMetadata, error) {
	return r.serverless.Resolver.DeleteFunction(ctx, namespace, function)
}

func (r *mutationResolver) DeleteManyFunctions(ctx context.Context, namespace string, functions []*gqlschema.FunctionMetadataInput) ([]*gqlschema.FunctionMetadata, error) {
	return r.serverless.Resolver.DeleteManyFunctions(ctx, namespace, functions)
}

func (r *namespaceResolver) Pods(ctx context.Context, obj *gqlschema.Namespace) ([]*gqlschema.Pod, error) {
	return r.k8s.PodsQuery(ctx, obj.Name, nil, nil)
}

func (r *namespaceResolver) Deployments(ctx context.Context, obj *gqlschema.Namespace, excludeFunctions *bool) ([]*gqlschema.Deployment, error) {
	return r.k8s.DeploymentsQuery(ctx, obj.Name, excludeFunctions)
}

func (r *namespaceResolver) Applications(ctx context.Context, obj *gqlschema.Namespace) ([]string, error) {
	return r.k8s.ApplicationsField(ctx, obj)
}

func (r *namespaceListItemResolver) PodsCount(ctx context.Context, obj *gqlschema.NamespaceListItem) (int, error) {
	return r.k8s.PodsCountField(ctx, obj)
}

func (r *namespaceListItemResolver) HealthyPodsCount(ctx context.Context, obj *gqlschema.NamespaceListItem) (int, error) {
	return r.k8s.HealthyPodsCountField(ctx, obj)
}

func (r *namespaceListItemResolver) ApplicationsCount(ctx context.Context, obj *gqlschema.NamespaceListItem) (*int, error) {
	return r.k8s.ApplicationsCountField(ctx, obj)
}

func (r *queryResolver) ClusterAssetGroups(ctx context.Context, viewContext *string, groupName *string) ([]*gqlschema.ClusterAssetGroup, error) {
	return r.rafter.Resolver.ClusterAssetGroupsQuery(ctx, viewContext, groupName)
}

func (r *queryResolver) ServiceInstance(ctx context.Context, name string, namespace string) (*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ServiceInstanceQuery(ctx, name, namespace)
}

func (r *queryResolver) ServiceInstances(ctx context.Context, namespace string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ServiceInstancesQuery(ctx, namespace, first, offset, status)
}

func (r *queryResolver) ClusterServiceClasses(ctx context.Context, first *int, offset *int) ([]*gqlschema.ClusterServiceClass, error) {
	return r.sc.Resolver.ClusterServiceClassesQuery(ctx, first, offset)
}

func (r *queryResolver) ClusterServiceClass(ctx context.Context, name string) (*gqlschema.ClusterServiceClass, error) {
	return r.sc.Resolver.ClusterServiceClassQuery(ctx, name)
}

func (r *queryResolver) ServiceClasses(ctx context.Context, namespace string, first *int, offset *int) ([]*gqlschema.ServiceClass, error) {
	return r.sc.Resolver.ServiceClassesQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ServiceClass(ctx context.Context, namespace string, name string) (*gqlschema.ServiceClass, error) {
	return r.sc.Resolver.ServiceClassQuery(ctx, name, namespace)
}

func (r *queryResolver) ClusterServiceBrokers(ctx context.Context, first *int, offset *int) ([]*gqlschema.ClusterServiceBroker, error) {
	return r.sc.Resolver.ClusterServiceBrokersQuery(ctx, first, offset)
}

func (r *queryResolver) ClusterServiceBroker(ctx context.Context, name string) (*gqlschema.ClusterServiceBroker, error) {
	return r.sc.Resolver.ClusterServiceBrokerQuery(ctx, name)
}

func (r *queryResolver) ServiceBrokers(ctx context.Context, namespace string, first *int, offset *int) ([]*gqlschema.ServiceBroker, error) {
	return r.sc.Resolver.ServiceBrokersQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ServiceBroker(ctx context.Context, name string, namespace string) (*gqlschema.ServiceBroker, error) {
	return r.sc.Resolver.ServiceBrokerQuery(ctx, name, namespace)
}

func (r *queryResolver) ServiceBindingUsage(ctx context.Context, name string, namespace string) (*gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.ServiceBindingUsageQuery(ctx, name, namespace)
}

func (r *queryResolver) ServiceBindingUsages(ctx context.Context, namespace string, resourceKind *string, resourceName *string) ([]*gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.ServiceBindingUsagesQuery(ctx, namespace, resourceKind, resourceName)
}

func (r *queryResolver) ServiceBinding(ctx context.Context, name string, namespace string) (*gqlschema.ServiceBinding, error) {
	return r.sc.Resolver.ServiceBindingQuery(ctx, name, namespace)
}

func (r *queryResolver) UsageKinds(ctx context.Context, first *int, offset *int) ([]*gqlschema.UsageKind, error) {
	return r.sca.Resolver.ListUsageKinds(ctx, first, offset)
}

func (r *queryResolver) ClusterAddonsConfigurations(ctx context.Context, first *int, offset *int) ([]*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.ClusterAddonsConfigurationsQuery(ctx, first, offset)
}

func (r *queryResolver) AddonsConfigurations(ctx context.Context, namespace string, first *int, offset *int) ([]*gqlschema.AddonsConfiguration, error) {
	return r.sca.Resolver.AddonsConfigurationsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) BindableResources(ctx context.Context, namespace string) ([]*gqlschema.BindableResourcesOutputItem, error) {
	return r.sca.Resolver.ListBindableResources(ctx, namespace)
}

func (r *queryResolver) Application(ctx context.Context, name string) (*gqlschema.Application, error) {
	return r.app.Resolver.ApplicationQuery(ctx, name)
}

func (r *queryResolver) Applications(ctx context.Context, namespace *string, first *int, offset *int) ([]*gqlschema.Application, error) {
	return r.app.Resolver.ApplicationsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ConnectorService(ctx context.Context, application string) (*gqlschema.ConnectorService, error) {
	return r.app.Resolver.ConnectorServiceQuery(ctx, application)
}

func (r *queryResolver) Namespaces(ctx context.Context, withSystemNamespaces *bool, withInactiveStatus *bool) ([]*gqlschema.NamespaceListItem, error) {
	return r.k8s.NamespacesQuery(ctx, withSystemNamespaces, withInactiveStatus)
}

func (r *queryResolver) Namespace(ctx context.Context, name string) (*gqlschema.Namespace, error) {
	return r.k8s.NamespaceQuery(ctx, name)
}

func (r *queryResolver) Deployments(ctx context.Context, namespace string, excludeFunctions *bool) ([]*gqlschema.Deployment, error) {
	return r.k8s.DeploymentsQuery(ctx, namespace, excludeFunctions)
}

func (r *queryResolver) VersionInfo(ctx context.Context) (*gqlschema.VersionInfo, error) {
	return r.k8s.VersionInfoQuery(ctx)
}

func (r *queryResolver) Pod(ctx context.Context, name string, namespace string) (*gqlschema.Pod, error) {
	return r.k8s.PodQuery(ctx, name, namespace)
}

func (r *queryResolver) Pods(ctx context.Context, namespace string, first *int, offset *int) ([]*gqlschema.Pod, error) {
	return r.k8s.PodsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) Service(ctx context.Context, name string, namespace string) (*gqlschema.Service, error) {
	return r.k8s.ServiceQuery(ctx, name, namespace)
}

func (r *queryResolver) Services(ctx context.Context, namespace string, excludedLabels []string, first *int, offset *int) ([]*gqlschema.Service, error) {
	return r.k8s.ServicesQuery(ctx, namespace, excludedLabels, first, offset)
}

func (r *queryResolver) ConfigMap(ctx context.Context, name string, namespace string) (*gqlschema.ConfigMap, error) {
	return r.k8s.ConfigMapQuery(ctx, name, namespace)
}

func (r *queryResolver) ConfigMaps(ctx context.Context, namespace string, first *int, offset *int) ([]*gqlschema.ConfigMap, error) {
	return r.k8s.ConfigMapsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) ReplicaSet(ctx context.Context, name string, namespace string) (*gqlschema.ReplicaSet, error) {
	return r.k8s.ReplicaSetQuery(ctx, name, namespace)
}

func (r *queryResolver) ReplicaSets(ctx context.Context, namespace string, first *int, offset *int) ([]*gqlschema.ReplicaSet, error) {
	return r.k8s.ReplicaSetsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) EventActivations(ctx context.Context, namespace string) ([]*gqlschema.EventActivation, error) {
	return r.app.Resolver.EventActivationsQuery(ctx, namespace)
}

func (r *queryResolver) BackendModules(ctx context.Context) ([]*gqlschema.BackendModule, error) {
	return r.ui.BackendModulesQuery(ctx)
}

func (r *queryResolver) Secret(ctx context.Context, name string, namespace string) (*gqlschema.Secret, error) {
	return r.k8s.SecretQuery(ctx, name, namespace)
}

func (r *queryResolver) Secrets(ctx context.Context, namespace string, first *int, offset *int) ([]*gqlschema.Secret, error) {
	return r.k8s.SecretsQuery(ctx, namespace, first, offset)
}

func (r *queryResolver) MicroFrontends(ctx context.Context, namespace string) ([]*gqlschema.MicroFrontend, error) {
	return r.ui.MicroFrontendsQuery(ctx, namespace)
}

func (r *queryResolver) ClusterMicroFrontends(ctx context.Context) ([]*gqlschema.ClusterMicroFrontend, error) {
	return r.ui.ClusterMicroFrontendsQuery(ctx)
}

func (r *queryResolver) SelfSubjectRules(ctx context.Context, namespace *string) ([]*gqlschema.ResourceRule, error) {
	return r.k8s.SelfSubjectRulesQuery(ctx, namespace)
}

func (r *queryResolver) Function(ctx context.Context, name string, namespace string) (*gqlschema.Function, error) {
	return r.serverless.FunctionQuery(ctx, name, namespace)
}

func (r *queryResolver) Functions(ctx context.Context, namespace string) ([]*gqlschema.Function, error) {
	return r.serverless.FunctionsQuery(ctx, namespace)
}

func (r *serviceBindingResolver) Secret(ctx context.Context, obj *gqlschema.ServiceBinding) (*gqlschema.Secret, error) {
	return r.k8s.SecretQuery(ctx, obj.SecretName, obj.Namespace)
}

func (r *serviceBindingUsageResolver) ServiceBinding(ctx context.Context, obj *gqlschema.ServiceBindingUsage) (*gqlschema.ServiceBinding, error) {
	return r.sc.Resolver.ServiceBindingQuery(ctx, obj.ServiceBindingName, obj.Namespace)
}

func (r *serviceClassResolver) Plans(ctx context.Context, obj *gqlschema.ServiceClass) ([]*gqlschema.ServicePlan, error) {
	return r.sc.Resolver.ServiceClassPlansField(ctx, obj)
}

func (r *serviceClassResolver) Activated(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	return r.sc.Resolver.ServiceClassActivatedField(ctx, obj)
}

func (r *serviceClassResolver) Instances(ctx context.Context, obj *gqlschema.ServiceClass) ([]*gqlschema.ServiceInstance, error) {
	return r.sc.Resolver.ServiceClassInstancesField(ctx, obj)
}

func (r *serviceClassResolver) ClusterAssetGroup(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.ClusterAssetGroup, error) {
	return r.sc.Resolver.ServiceClassClusterAssetGroupField(ctx, obj)
}

func (r *serviceClassResolver) AssetGroup(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.AssetGroup, error) {
	return r.sc.Resolver.ServiceClassAssetGroupField(ctx, obj)
}

func (r *serviceInstanceResolver) ServiceClass(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error) {
	return r.sc.Resolver.ServiceInstanceServiceClassField(ctx, obj)
}

func (r *serviceInstanceResolver) ClusterServiceClass(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServiceClass, error) {
	return r.sc.Resolver.ServiceInstanceClusterServiceClassField(ctx, obj)
}

func (r *serviceInstanceResolver) ServicePlan(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServicePlan, error) {
	return r.sc.Resolver.ServiceInstanceServicePlanField(ctx, obj)
}

func (r *serviceInstanceResolver) ClusterServicePlan(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServicePlan, error) {
	return r.sc.Resolver.ServiceInstanceClusterServicePlanField(ctx, obj)
}

func (r *serviceInstanceResolver) Bindable(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error) {
	return r.sc.Resolver.ServiceInstanceBindableField(ctx, obj)
}

func (r *serviceInstanceResolver) ServiceBindings(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceBindings, error) {
	return r.sc.Resolver.ServiceBindingsToInstanceQuery(ctx, obj.Name, obj.Namespace)
}

func (r *serviceInstanceResolver) ServiceBindingUsages(ctx context.Context, obj *gqlschema.ServiceInstance) ([]*gqlschema.ServiceBindingUsage, error) {
	return r.sca.Resolver.ServiceBindingUsagesOfInstanceQuery(ctx, obj.Name, obj.Namespace)
}

func (r *servicePlanResolver) ClusterAssetGroup(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.ClusterAssetGroup, error) {
	return r.sc.Resolver.ServicePlanClusterAssetGroupField(ctx, obj)
}

func (r *servicePlanResolver) AssetGroup(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.AssetGroup, error) {
	return r.sc.Resolver.ServicePlanAssetGroupField(ctx, obj)
}

func (r *subscriptionResolver) ClusterAssetEvent(ctx context.Context) (<-chan *gqlschema.ClusterAssetEvent, error) {
	return r.rafter.Resolver.ClusterAssetEventSubscription(ctx)
}

func (r *subscriptionResolver) AssetEvent(ctx context.Context, namespace string) (<-chan *gqlschema.AssetEvent, error) {
	return r.rafter.Resolver.AssetEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ClusterAssetGroupEvent(ctx context.Context) (<-chan *gqlschema.ClusterAssetGroupEvent, error) {
	return r.rafter.Resolver.ClusterAssetGroupEventSubscription(ctx)
}

func (r *subscriptionResolver) AssetGroupEvent(ctx context.Context, namespace string) (<-chan *gqlschema.AssetGroupEvent, error) {
	return r.rafter.Resolver.AssetGroupEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ServiceInstanceEvent(ctx context.Context, namespace string) (<-chan *gqlschema.ServiceInstanceEvent, error) {
	return r.sc.Resolver.ServiceInstanceEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ServiceBindingEvent(ctx context.Context, namespace string) (<-chan *gqlschema.ServiceBindingEvent, error) {
	return r.sc.Resolver.ServiceBindingEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ServiceBindingUsageEvent(ctx context.Context, namespace string, resourceKind *string, resourceName *string) (<-chan *gqlschema.ServiceBindingUsageEvent, error) {
	return r.sca.Resolver.ServiceBindingUsageEventSubscription(ctx, namespace, resourceKind, resourceName)
}

func (r *subscriptionResolver) ServiceBrokerEvent(ctx context.Context, namespace string) (<-chan *gqlschema.ServiceBrokerEvent, error) {
	return r.sc.Resolver.ServiceBrokerEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ClusterServiceBrokerEvent(ctx context.Context) (<-chan *gqlschema.ClusterServiceBrokerEvent, error) {
	return r.sc.Resolver.ClusterServiceBrokerEventSubscription(ctx)
}

func (r *subscriptionResolver) ApplicationEvent(ctx context.Context) (<-chan *gqlschema.ApplicationEvent, error) {
	return r.app.Resolver.ApplicationEventSubscription(ctx)
}

func (r *subscriptionResolver) PodEvent(ctx context.Context, namespace string) (<-chan *gqlschema.PodEvent, error) {
	return r.k8s.PodEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) DeploymentEvent(ctx context.Context, namespace string) (<-chan *gqlschema.DeploymentEvent, error) {
	return r.k8s.DeploymentEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ServiceEvent(ctx context.Context, namespace string) (<-chan *gqlschema.ServiceEvent, error) {
	return r.k8s.ServiceEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ConfigMapEvent(ctx context.Context, namespace string) (<-chan *gqlschema.ConfigMapEvent, error) {
	return r.k8s.ConfigMapEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) SecretEvent(ctx context.Context, namespace string) (<-chan *gqlschema.SecretEvent, error) {
	return r.k8s.SecretEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ClusterAddonsConfigurationEvent(ctx context.Context) (<-chan *gqlschema.ClusterAddonsConfigurationEvent, error) {
	return r.sca.Resolver.ClusterAddonsConfigurationEventSubscription(ctx)
}

func (r *subscriptionResolver) AddonsConfigurationEvent(ctx context.Context, namespace string) (<-chan *gqlschema.AddonsConfigurationEvent, error) {
	return r.sca.Resolver.AddonsConfigurationEventSubscription(ctx, namespace)
}

func (r *subscriptionResolver) NamespaceEvent(ctx context.Context, withSystemNamespaces *bool) (<-chan *gqlschema.NamespaceEvent, error) {
	return r.k8s.NamespaceEventSubscription(ctx, withSystemNamespaces)
}

func (r *subscriptionResolver) FunctionEvent(ctx context.Context, namespace string, functionName *string) (<-chan *gqlschema.FunctionEvent, error) {
	return r.serverless.FunctionEventSubscription(ctx, namespace, functionName)
}

// Application returns gqlschema.ApplicationResolver implementation.
func (r *Resolver) Application() gqlschema.ApplicationResolver { return &applicationResolver{r} }

// Asset returns gqlschema.AssetResolver implementation.
func (r *Resolver) Asset() gqlschema.AssetResolver { return &assetResolver{r} }

// AssetGroup returns gqlschema.AssetGroupResolver implementation.
func (r *Resolver) AssetGroup() gqlschema.AssetGroupResolver { return &assetGroupResolver{r} }

// ClusterAsset returns gqlschema.ClusterAssetResolver implementation.
func (r *Resolver) ClusterAsset() gqlschema.ClusterAssetResolver { return &clusterAssetResolver{r} }

// ClusterAssetGroup returns gqlschema.ClusterAssetGroupResolver implementation.
func (r *Resolver) ClusterAssetGroup() gqlschema.ClusterAssetGroupResolver {
	return &clusterAssetGroupResolver{r}
}

// ClusterServiceClass returns gqlschema.ClusterServiceClassResolver implementation.
func (r *Resolver) ClusterServiceClass() gqlschema.ClusterServiceClassResolver {
	return &clusterServiceClassResolver{r}
}

// ClusterServicePlan returns gqlschema.ClusterServicePlanResolver implementation.
func (r *Resolver) ClusterServicePlan() gqlschema.ClusterServicePlanResolver {
	return &clusterServicePlanResolver{r}
}

// Deployment returns gqlschema.DeploymentResolver implementation.
func (r *Resolver) Deployment() gqlschema.DeploymentResolver { return &deploymentResolver{r} }

// EventActivation returns gqlschema.EventActivationResolver implementation.
func (r *Resolver) EventActivation() gqlschema.EventActivationResolver {
	return &eventActivationResolver{r}
}

// Mutation returns gqlschema.MutationResolver implementation.
func (r *Resolver) Mutation() gqlschema.MutationResolver { return &mutationResolver{r} }

// Namespace returns gqlschema.NamespaceResolver implementation.
func (r *Resolver) Namespace() gqlschema.NamespaceResolver { return &namespaceResolver{r} }

// NamespaceListItem returns gqlschema.NamespaceListItemResolver implementation.
func (r *Resolver) NamespaceListItem() gqlschema.NamespaceListItemResolver {
	return &namespaceListItemResolver{r}
}

// Query returns gqlschema.QueryResolver implementation.
func (r *Resolver) Query() gqlschema.QueryResolver { return &queryResolver{r} }

// ServiceBinding returns gqlschema.ServiceBindingResolver implementation.
func (r *Resolver) ServiceBinding() gqlschema.ServiceBindingResolver {
	return &serviceBindingResolver{r}
}

// ServiceBindingUsage returns gqlschema.ServiceBindingUsageResolver implementation.
func (r *Resolver) ServiceBindingUsage() gqlschema.ServiceBindingUsageResolver {
	return &serviceBindingUsageResolver{r}
}

// ServiceClass returns gqlschema.ServiceClassResolver implementation.
func (r *Resolver) ServiceClass() gqlschema.ServiceClassResolver { return &serviceClassResolver{r} }

// ServiceInstance returns gqlschema.ServiceInstanceResolver implementation.
func (r *Resolver) ServiceInstance() gqlschema.ServiceInstanceResolver {
	return &serviceInstanceResolver{r}
}

// ServicePlan returns gqlschema.ServicePlanResolver implementation.
func (r *Resolver) ServicePlan() gqlschema.ServicePlanResolver { return &servicePlanResolver{r} }

// Subscription returns gqlschema.SubscriptionResolver implementation.
func (r *Resolver) Subscription() gqlschema.SubscriptionResolver { return &subscriptionResolver{r} }

type applicationResolver struct{ *Resolver }
type assetResolver struct{ *Resolver }
type assetGroupResolver struct{ *Resolver }
type clusterAssetResolver struct{ *Resolver }
type clusterAssetGroupResolver struct{ *Resolver }
type clusterServiceClassResolver struct{ *Resolver }
type clusterServicePlanResolver struct{ *Resolver }
type deploymentResolver struct{ *Resolver }
type eventActivationResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type namespaceResolver struct{ *Resolver }
type namespaceListItemResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type serviceBindingResolver struct{ *Resolver }
type serviceBindingUsageResolver struct{ *Resolver }
type serviceClassResolver struct{ *Resolver }
type serviceInstanceResolver struct{ *Resolver }
type servicePlanResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
