package shared

import (
	usageApi "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
)

//go:generate mockery -name=ServiceCatalogAddonsRetriever -output=automock -outpkg=automock -case=underscore
type ServiceCatalogAddonsRetriever interface {
	ServiceBindingUsage() ServiceBindingUsageLister
}

//go:generate mockery -name=ServiceBindingUsageLister -output=automock -outpkg=automock -case=underscore
type ServiceBindingUsageLister interface {
	ListForDeployment(environment, kind, deploymentName string) ([]*usageApi.ServiceBindingUsage, error)
}
