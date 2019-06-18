package apicontroller

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
)

//go:generate mockery -name=apiLister -output=automock -outpkg=automock -case=underscore
type apiLister interface {
	List(namespace string, serviceName *string, hostname *string) ([]*v1alpha2.Api, error)
	Find(name string, namespace string) (*v1alpha2.Api, error)
	Create(name string, namespace string, params gqlschema.APICreateInput) (*v1alpha2.Api, error)
	Update(name string, namespace string, params gqlschema.APICreateInput) (*v1alpha2.Api, error)
	Delete(name string, namespace string) error
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}
