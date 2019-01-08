package shared

import "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

//go:generate mockery -name=ApplicationRetriever -output=automock -outpkg=automock -case=underscore
type ApplicationRetriever interface {
	Application() ApplicationLister
}

//go:generate mockery -name=ApplicationLister -output=automock -outpkg=automock -case=underscore
type ApplicationLister interface {
	ListInEnvironment(environment string) ([]*v1alpha1.Application, error)
	ListNamespacesFor(appName string) ([]string, error)
}
