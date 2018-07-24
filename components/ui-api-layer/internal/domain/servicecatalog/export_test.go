package servicecatalog

// ServiceInstance
import (
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

func NewInstanceService(informer cache.SharedIndexInformer, client clientset.Interface) *instanceService {
	return newInstanceService(informer, client)
}

func NewInstanceResolver(iS instanceSvc, pG planGetter, cG classGetter) *instanceResolver {
	return newInstanceResolver(iS, pG, cG)
}

func NewMockInstanceConverter() *mockGqlInstanceConverter {
	return new(mockGqlInstanceConverter)
}

func NewMockInstanceService() *mockInstanceSvc {
	return new(mockInstanceSvc)
}

func NewInstanceListener(channel chan<- gqlschema.ServiceInstanceEvent, filter func(object interface{}) bool, instanceConverter gqlInstanceConverter) *instanceListener {
	return &instanceListener{
		channel:           channel,
		filter:            filter,
		instanceConverter: instanceConverter,
	}
}

func NewInstanceCreateParameters(name, namespace string, labels []string, externalServicePlanName, externalServiceClassName string, schema map[string]interface{}) *instanceCreateParameters {
	return &instanceCreateParameters{
		Name:                     name,
		Namespace:                namespace,
		Labels:                   labels,
		ExternalServicePlanName:  externalServicePlanName,
		ExternalServiceClassName: externalServiceClassName,
		Schema: schema,
	}
}

func (r *instanceResolver) SetInstanceConverter(converter gqlInstanceConverter) {
	r.instanceConverter = converter
}

func (r *instanceResolver) SetClassConverter(converter gqlClassConverter) {
	r.classConverter = converter
}

func (r *instanceResolver) SetPlanConverter(converter gqlPlanConverter) {
	r.planConverter = converter
}

// ServiceClass

func NewClassService(informer cache.SharedIndexInformer) *classService {
	return newClassService(informer)
}

func NewClassResolver(classLister classListGetter, planLister planLister, instanceLister classInstanceLister, asyncApiSpecGetter AsyncApiSpecGetter, apiSpecGetter ApiSpecGetter, contentGetter ContentGetter) *classResolver {
	return newClassResolver(classLister, planLister, instanceLister, asyncApiSpecGetter, apiSpecGetter, contentGetter)
}

func (r *classResolver) SetClassConverter(converter gqlClassConverter) {
	r.classConverter = converter
}

func (r *classResolver) SetPlanConverter(converter gqlPlanConverter) {
	r.planConverter = converter
}

// ServiceBroker

func NewBrokerService(informer cache.SharedIndexInformer) *brokerService {
	return newBrokerService(informer)
}

// ServicePlan

func NewPlanService(informer cache.SharedIndexInformer) *planService {
	return newPlanService(informer)
}

func NewServiceBindingResolver(sbService serviceBindingOperations) *serviceBindingResolver {
	return newServiceBindingResolver(sbService)
}
func NewServiceBindingService(client v1beta1.ServicecatalogV1beta1Interface, informer cache.SharedIndexInformer) *serviceBindingService {
	return newServiceBindingService(client, informer)
}

// Binding usage

func NewServiceBindingUsageService(buInterface v1alpha1.ServicecatalogV1alpha1Interface, informer cache.SharedIndexInformer, bindingOp serviceBindingOperations) *serviceBindingUsageService {
	return newServiceBindingUsageService(buInterface, informer, bindingOp)
}

func NewServiceBindingUsageResolver(op serviceBindingUsageOperations) *serviceBindingUsageResolver {
	return newServiceBindingUsageResolver(op)
}
