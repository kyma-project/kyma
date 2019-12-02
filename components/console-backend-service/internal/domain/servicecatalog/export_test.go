package servicecatalog

// ServiceInstance
import (
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
)

func NewServiceInstanceService(informer cache.SharedIndexInformer, client clientset.Interface) (*serviceInstanceService, error) {
	return newServiceInstanceService(informer, client)
}

func NewServiceInstanceResolver(serviceInstanceSvc serviceInstanceSvc, clusterServicePlanGetter clusterServicePlanGetter, clusterServiceClassGetter clusterServiceClassGetter, servicePlanGetter servicePlanGetter, serviceClassGetter serviceClassGetter) *serviceInstanceResolver {
	return newServiceInstanceResolver(serviceInstanceSvc, clusterServicePlanGetter, clusterServiceClassGetter, servicePlanGetter, serviceClassGetter)
}

func NewMockServiceInstanceConverter() *mockGqlServiceInstanceConverter {
	return new(mockGqlServiceInstanceConverter)
}

func NewMockServiceInstanceService() *mockServiceInstanceSvc {
	return new(mockServiceInstanceSvc)
}

func NewServiceInstanceCreateParameters(name, namespace string, labels []string, externalServicePlanName string, servicePlanClusterWide bool, externalServiceClassName string, serviceClassClusterWide bool, schema map[string]interface{}) *serviceInstanceCreateParameters {
	return &serviceInstanceCreateParameters{
		Name:      name,
		Namespace: namespace,
		Labels:    labels,
		PlanRef: instanceCreateResourceRef{
			ExternalName: externalServicePlanName,
			ClusterWide:  servicePlanClusterWide,
		},
		ClassRef: instanceCreateResourceRef{
			ExternalName: externalServiceClassName,
			ClusterWide:  serviceClassClusterWide,
		},
		Schema: schema,
	}
}

func (r *serviceInstanceResolver) SetInstanceConverter(converter gqlServiceInstanceConverter) {
	r.instanceConverter = converter
}

func (r *serviceInstanceResolver) SetClusterServiceClassConverter(converter gqlClusterServiceClassConverter) {
	r.clusterServiceClassConverter = converter
}

func (r *serviceInstanceResolver) SetClusterServicePlanConverter(converter gqlClusterServicePlanConverter) {
	r.clusterServicePlanConverter = converter
}

func (r *serviceInstanceResolver) SetServiceClassConverter(converter gqlServiceClassConverter) {
	r.serviceClassConverter = converter
}

func (r *serviceInstanceResolver) SetServicePlanConverter(converter gqlServicePlanConverter) {
	r.servicePlanConverter = converter
}

// ServiceClass

func NewServiceClassService(informer cache.SharedIndexInformer) (*serviceClassService, error) {
	return newServiceClassService(informer)
}

func NewServiceClassResolver(classLister serviceClassListGetter, planLister servicePlanLister, instanceLister instanceListerByServiceClass, cmsRetriever shared.CmsRetriever, rafterRetriever shared.RafterRetriever) *serviceClassResolver {
	return newServiceClassResolver(classLister, planLister, instanceLister, cmsRetriever, rafterRetriever)
}

func (r *serviceClassResolver) SetClassConverter(converter gqlServiceClassConverter) {
	r.classConverter = converter
}

func (r *serviceClassResolver) SetPlanConverter(converter gqlServicePlanConverter) {
	r.planConverter = converter
}

// ClusterServiceClass

func NewClusterServiceClassService(informer cache.SharedIndexInformer) (*clusterServiceClassService, error) {
	return newClusterServiceClassService(informer)
}

func NewClusterServiceClassResolver(classLister clusterServiceClassListGetter, planLister clusterServicePlanLister, instanceLister instanceListerByClusterServiceClass, cmsRetriever shared.CmsRetriever, rafterRetriever shared.RafterRetriever) *clusterServiceClassResolver {
	return newClusterServiceClassResolver(classLister, planLister, instanceLister, cmsRetriever, rafterRetriever)
}

func (r *clusterServiceClassResolver) SetClassConverter(converter gqlClusterServiceClassConverter) {
	r.classConverter = converter
}

func (r *clusterServiceClassResolver) SetPlanConverter(converter gqlClusterServicePlanConverter) {
	r.planConverter = converter
}

// ClusterServiceBroker

func NewClusterServiceBrokerResolver(brokerSvc clusterServiceBrokerSvc) *clusterServiceBrokerResolver {
	return newClusterServiceBrokerResolver(brokerSvc)
}

func (r *clusterServiceBrokerResolver) SetBrokerConverter(converter gqlClusterServiceBrokerConverter) {
	r.brokerConverter = converter
}

func NewClusterServiceBrokerService(informer cache.SharedIndexInformer) *clusterServiceBrokerService {
	return newClusterServiceBrokerService(informer)
}

// ServiceBroker

func NewServiceBrokerResolver(brokerSvc serviceBrokerSvc) *serviceBrokerResolver {
	return newServiceBrokerResolver(brokerSvc)
}

func (r *serviceBrokerResolver) SetBrokerConverter(converter gqlServiceBrokerConverter) {
	r.brokerConverter = converter
}

func NewServiceBrokerService(informer cache.SharedIndexInformer) *serviceBrokerService {
	return newServiceBrokerService(informer)
}

// ServicePlan

func NewServicePlanService(informer cache.SharedIndexInformer) (*servicePlanService, error) {
	return newServicePlanService(informer)
}

func NewClusterServicePlanService(informer cache.SharedIndexInformer) (*clusterServicePlanService, error) {
	return newClusterServicePlanService(informer)
}

// ServiceBinding

func NewServiceBindingResolver(sbService serviceBindingOperations) *serviceBindingResolver {
	return newServiceBindingResolver(sbService)
}
func NewServiceBindingService(client v1beta1.ServicecatalogV1beta1Interface, informer cache.SharedIndexInformer, sbName string) (*serviceBindingService, error) {
	return newServiceBindingService(client, informer, func() string {
		return sbName
	})
}

// Service Catalog Module

func (r *PluggableContainer) SetFakeClient() {
	r.cfg.scCli = fake.NewSimpleClientset()
}
