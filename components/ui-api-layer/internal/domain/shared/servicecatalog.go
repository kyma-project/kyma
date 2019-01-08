package shared

import (
	bindingApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	usageApi "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
)

//go:generate mockery -name=ServiceCatalogRetriever -output=automock -outpkg=automock -case=underscore
type ServiceCatalogRetriever interface {
	ServiceBinding() ServiceBindingGetter
	ServiceBindingUsage() ServiceBindingUsageLister
}

//go:generate mockery -name=ServiceBindingUsageLister -output=automock -outpkg=automock -case=underscore
type ServiceBindingUsageLister interface {
	ListForDeployment(environment, kind, deploymentName string) ([]*usageApi.ServiceBindingUsage, error)
}

//go:generate mockery -name=ServiceBindingGetter -output=automock -outpkg=automock -case=underscore
type ServiceBindingGetter interface {
	Find(env string, name string) (*bindingApi.ServiceBinding, error)
}
