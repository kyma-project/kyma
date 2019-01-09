package shared

import (
	bindingApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

//go:generate mockery -name=ServiceCatalogRetriever -output=automock -outpkg=automock -case=underscore
type ServiceCatalogRetriever interface {
	ServiceBinding() ServiceBindingFinderLister
}

//go:generate mockery -name=ServiceBindingFinderLister -output=automock -outpkg=automock -case=underscore
type ServiceBindingFinderLister interface {
	Find(env string, name string) (*bindingApi.ServiceBinding, error)
	ListForServiceInstance(env string, instanceName string) ([]*bindingApi.ServiceBinding, error)
}
