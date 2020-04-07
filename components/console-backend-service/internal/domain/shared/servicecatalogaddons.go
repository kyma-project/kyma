package shared

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	usageApi "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
)

//go:generate mockery -name=ServiceCatalogAddonsRetriever -output=automock -outpkg=automock -case=underscore
type ServiceCatalogAddonsRetriever interface {
	ServiceBindingUsage() ServiceBindingUsageLister
	ServiceBindingUsageConverter() GqlServiceBindingUsageConverter
}

//go:generate mockery -name=ServiceBindingUsageLister -output=automock -outpkg=automock -case=underscore
type ServiceBindingUsageLister interface {
	ListByUsageKind(ns, kind, resourceName string) ([]*usageApi.ServiceBindingUsage, error)
}

//go:generate mockery -name=GqlServiceBindingUsageConverter -output=automock -outpkg=automock -case=underscore
type GqlServiceBindingUsageConverter interface {
	ToGQL(item *usageApi.ServiceBindingUsage) (*gqlschema.ServiceBindingUsage, error)
	ToGQLs(in []*usageApi.ServiceBindingUsage) ([]gqlschema.ServiceBindingUsage, error)
}
