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

// Class

//go:generate mockery -name=classGetter -output=automock -outpkg=automock -case=underscore
type classGetter interface {
	Find(name string) (*v1beta1.ClusterServiceClass, error)
	FindByExternalName(externalName string) (*v1beta1.ClusterServiceClass, error)
}

//go:generate mockery -name=classListGetter -output=automock -outpkg=automock -case=underscore
type classListGetter interface {
	classGetter
	List(pagingParams pager.PagingParams) ([]*v1beta1.ClusterServiceClass, error)
}

//go:generate mockery -name=gqlClassConverter -output=automock -outpkg=automock -case=underscore
type gqlClassConverter interface {
	ToGQL(in *v1beta1.ClusterServiceClass) (*gqlschema.ServiceClass, error)
	ToGQLs(in []*v1beta1.ClusterServiceClass) ([]gqlschema.ServiceClass, error)
}

// Instance

//go:generate mockery -name=instanceGetter -output=automock -outpkg=automock -case=underscore
type instanceGetter interface {
	Find(name, environment string) (*v1beta1.ServiceInstance, error)
}

//go:generate mockery -name=instanceLister -output=automock -outpkg=automock -case=underscore
type instanceLister interface {
	List(environment string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceInstance, error)
	ListForStatus(environment string, pagingParams pager.PagingParams, status *status.ServiceInstanceStatusType) ([]*v1beta1.ServiceInstance, error)
}

//go:generate mockery -name=instanceSvc -inpkg -case=underscore
type instanceSvc interface {
	instanceGetter
	instanceLister
	Create(params instanceCreateParameters) (*v1beta1.ServiceInstance, error)
	Delete(name, namespace string) error
	IsBindable(relatedClass *v1beta1.ClusterServiceClass, relatedPlan *v1beta1.ClusterServicePlan) bool
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=classInstanceLister -output=automock -outpkg=automock -case=underscore
type classInstanceLister interface {
	ListForClass(className, externalClassName string) ([]*v1beta1.ServiceInstance, error)
}

//go:generate mockery -name=gqlInstanceConverter -inpkg -case=underscore
type gqlInstanceConverter interface {
	ToGQL(in *v1beta1.ServiceInstance) (*gqlschema.ServiceInstance, error)
	ToGQLs(in []*v1beta1.ServiceInstance) ([]gqlschema.ServiceInstance, error)
	GQLCreateInputToInstanceCreateParameters(in *gqlschema.ServiceInstanceCreateInput) *instanceCreateParameters
	ServiceStatusTypeToGQLStatusType(in status.ServiceInstanceStatusType) gqlschema.InstanceStatusType
	GQLStatusTypeToServiceStatusType(in gqlschema.InstanceStatusType) status.ServiceInstanceStatusType
	GQLStatusToServiceStatus(in *gqlschema.ServiceInstanceStatus) *status.ServiceInstanceStatus
	ServiceStatusToGQLStatus(in *status.ServiceInstanceStatus) *gqlschema.ServiceInstanceStatus
}

// Plan

//go:generate mockery -name=planGetter -output=automock -outpkg=automock -case=underscore
type planGetter interface {
	Find(name string) (*v1beta1.ClusterServicePlan, error)
	FindByExternalNameForClass(planExternalName, className string) (*v1beta1.ClusterServicePlan, error)
}

//go:generate mockery -name=planLister  -output=automock -outpkg=automock -case=underscore
type planLister interface {
	ListForClass(name string) ([]*v1beta1.ClusterServicePlan, error)
}

//go:generate mockery -name=gqlPlanConverter  -output=automock -outpkg=automock -case=underscore
type gqlPlanConverter interface {
	ToGQL(item *v1beta1.ClusterServicePlan) (*gqlschema.ServicePlan, error)
	ToGQLs(in []*v1beta1.ClusterServicePlan) ([]gqlschema.ServicePlan, error)
}

// Broker

//go:generate mockery -name=brokerGetter -output=automock -outpkg=automock -case=underscore
type brokerGetter interface {
	Find(name string) (*v1beta1.ClusterServiceBroker, error)
}

//go:generate mockery -name=brokerLister -output=automock -outpkg=automock -case=underscore
type brokerLister interface {
	List(pagingParams pager.PagingParams) ([]*v1beta1.ClusterServiceBroker, error)
}

//go:generate mockery -name=brokerListGetter -output=automock -outpkg=automock -case=underscore
type brokerListGetter interface {
	brokerGetter
	brokerLister
}

//go:generate mockery -name=gqlBrokerConverter -output=automock -outpkg=automock -case=underscore
type gqlBrokerConverter interface {
	ToGQL(in *v1beta1.ClusterServiceBroker) (*gqlschema.ServiceBroker, error)
	ToGQLs(in []*v1beta1.ClusterServiceBroker) ([]gqlschema.ServiceBroker, error)
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
}

// Binding usage

//go:generate mockery -name=serviceBindingUsageOperations -output=automock -outpkg=automock -case=underscore
type serviceBindingUsageOperations interface {
	Create(env string, sb *api.ServiceBindingUsage) (*api.ServiceBindingUsage, error)
	Delete(env string, name string) error
	Find(env string, name string) (*api.ServiceBindingUsage, error)
	ListForServiceInstance(env string, instanceName string) ([]*api.ServiceBindingUsage, error)
}
