package servicecatalog

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/resource"
)

// Content

//go:generate mockery -name=AsyncApiSpecGetter -output=automock -outpkg=automock -case=underscore
type AsyncApiSpecGetter interface {
	Find(kind, id string) (*storage.AsyncApiSpec, error)
}

//go:generate mockery -name=ApiSpecGetter -output=automock -outpkg=automock -case=underscore
type ApiSpecGetter interface {
	Find(kind, id string) (*storage.ApiSpec, error)
}

//go:generate mockery -name=ContentGetter -output=automock -outpkg=automock -case=underscore
type ContentGetter interface {
	Find(kind, id string) (*storage.Content, error)
}

// ServiceClass

//go:generate mockery -name=serviceClassGetter -output=automock -outpkg=automock -case=underscore
type serviceClassGetter interface {
	Find(name, environment string) (*v1beta1.ServiceClass, error)
	FindByExternalName(externalName, environment string) (*v1beta1.ServiceClass, error)
}

// ClusterServiceClass

//go:generate mockery -name=clusterServiceClassGetter -output=automock -outpkg=automock -case=underscore
type clusterServiceClassGetter interface {
	Find(name string) (*v1beta1.ClusterServiceClass, error)
	FindByExternalName(externalName string) (*v1beta1.ClusterServiceClass, error)
}

// ServiceInstance
//go:generate mockery -name=serviceInstanceLister -output=automock -outpkg=automock -case=underscore
type serviceInstanceLister interface {
	Find(name, environment string) (*v1beta1.ServiceInstance, error)
	List(environment string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceInstance, error)
	ListForStatus(environment string, pagingParams pager.PagingParams, status *status.ServiceInstanceStatusType) ([]*v1beta1.ServiceInstance, error)
}

//go:generate mockery -name=serviceInstanceSvc -inpkg -case=underscore
type serviceInstanceSvc interface {
	serviceInstanceLister
	Create(params serviceInstanceCreateParameters) (*v1beta1.ServiceInstance, error)
	Delete(name, namespace string) error
	IsBindableWithClusterRefs(relatedClass *v1beta1.ClusterServiceClass, relatedPlan *v1beta1.ClusterServicePlan) bool
	IsBindableWithLocalRefs(relatedClass *v1beta1.ServiceClass, relatedPlan *v1beta1.ServicePlan) bool
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=gqlServiceInstanceConverter -inpkg -case=underscore
type gqlServiceInstanceConverter interface {
	ToGQL(in *v1beta1.ServiceInstance) (*gqlschema.ServiceInstance, error)
	ToGQLs(in []*v1beta1.ServiceInstance) ([]gqlschema.ServiceInstance, error)
	GQLCreateInputToInstanceCreateParameters(in *gqlschema.ServiceInstanceCreateInput) *serviceInstanceCreateParameters
	ServiceStatusTypeToGQLStatusType(in status.ServiceInstanceStatusType) gqlschema.InstanceStatusType
	GQLStatusTypeToServiceStatusType(in gqlschema.InstanceStatusType) status.ServiceInstanceStatusType
	GQLStatusToServiceStatus(in *gqlschema.ServiceInstanceStatus) *status.ServiceInstanceStatus
	ServiceStatusToGQLStatus(in status.ServiceInstanceStatus) gqlschema.ServiceInstanceStatus
}

// ClusterServicePlan

//go:generate mockery -name=gqlClusterServicePlanConverter  -output=automock -outpkg=automock -case=underscore
type gqlClusterServicePlanConverter interface {
	ToGQL(item *v1beta1.ClusterServicePlan) (*gqlschema.ClusterServicePlan, error)
	ToGQLs(in []*v1beta1.ClusterServicePlan) ([]gqlschema.ClusterServicePlan, error)
}

// ServicePlan

//go:generate mockery -name=gqlServicePlanConverter  -output=automock -outpkg=automock -case=underscore
type gqlServicePlanConverter interface {
	ToGQL(item *v1beta1.ServicePlan) (*gqlschema.ServicePlan, error)
	ToGQLs(in []*v1beta1.ServicePlan) ([]gqlschema.ServicePlan, error)
}

// Notifier

type notifier interface {
	AddListener(observer resource.Listener)
	DeleteListener(observer resource.Listener)
}

// Binding

//go:generate mockery -name=serviceBindingOperations -output=automock -outpkg=automock -case=underscore
type serviceBindingOperations interface {
	Create(env string, sb *v1beta1.ServiceBinding) (*v1beta1.ServiceBinding, error)
	Delete(env string, name string) error
	Find(env string, name string) (*v1beta1.ServiceBinding, error)
	ListForServiceInstance(env string, instanceName string) ([]*v1beta1.ServiceBinding, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

// Binding usage

//go:generate mockery -name=serviceBindingUsageOperations -output=automock -outpkg=automock -case=underscore
type serviceBindingUsageOperations interface {
	Create(env string, sb *api.ServiceBindingUsage) (*api.ServiceBindingUsage, error)
	Delete(env string, name string) error
	Find(env string, name string) (*api.ServiceBindingUsage, error)
	ListForServiceInstance(env string, instanceName string) ([]*api.ServiceBindingUsage, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}
