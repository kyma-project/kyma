package k8s

import (
	bindingApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	usageApi "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	api "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
)

//go:generate mockery -name=deploymentLister -output=automock -outpkg=automock -case=underscore
type deploymentLister interface {
	List(environment string) ([]*api.Deployment, error)
	ListWithoutFunctions(environment string) ([]*api.Deployment, error)
}

//go:generate mockery -name=resourceQuotaLister -output=automock -outpkg=automock -case=underscore
type resourceQuotaLister interface {
	List(environment string) ([]*v1.ResourceQuota, error)
}

//go:generate mockery -name=ServiceBindingUsageLister -output=automock -outpkg=automock -case=underscore
type ServiceBindingUsageLister interface {
	ListForDeployment(environment, kind, deploymentName string) ([]*usageApi.ServiceBindingUsage, error)
}

//go:generate mockery -name=ServiceBindingGetter -output=automock -outpkg=automock -case=underscore
type ServiceBindingGetter interface {
	Find(env string, name string) (*bindingApi.ServiceBinding, error)
}
