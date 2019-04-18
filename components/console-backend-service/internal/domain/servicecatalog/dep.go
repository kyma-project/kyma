package servicecatalog

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
)

// ServiceClass

//go:generate mockery -name=serviceClassGetter -output=automock -outpkg=automock -case=underscore
type serviceClassGetter interface {
	Find(name, namespace string) (*v1beta1.ServiceClass, error)
	FindByExternalName(externalName, namespace string) (*v1beta1.ServiceClass, error)
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
	Find(name, namespace string) (*v1beta1.ServiceInstance, error)
	List(namespace string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceInstance, error)
	ListForStatus(namespace string, pagingParams pager.PagingParams, status *status.ServiceInstanceStatusType) ([]*v1beta1.ServiceInstance, error)
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
	GQLCreateInputToInstanceCreateParameters(in *gqlschema.ServiceInstanceCreateInput, namespace string) *serviceInstanceCreateParameters
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

// Binding

//go:generate mockery -name=serviceBindingOperations -output=automock -outpkg=automock -case=underscore
type serviceBindingOperations interface {
	Create(namespace string, sb *v1beta1.ServiceBinding) (*v1beta1.ServiceBinding, error)
	Delete(namespace string, name string) error
	Find(namespace string, name string) (*v1beta1.ServiceBinding, error)
	ListForServiceInstance(namespace string, instanceName string) ([]*v1beta1.ServiceBinding, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type notifier interface {
	AddListener(observer resource.Listener)
	DeleteListener(observer resource.Listener)
}
