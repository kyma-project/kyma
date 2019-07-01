package apicontroller

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
)

//go:generate mockery -name=apiSvc -output=automock -outpkg=automock -case=underscore
type apiSvc interface {
	List(namespace string, serviceName *string, hostname *string) ([]*v1alpha2.Api, error)
	Find(name string, namespace string) (*v1alpha2.Api, error)
	Create(api *v1alpha2.Api) (*v1alpha2.Api, error)
	Update(api *v1alpha2.Api) (*v1alpha2.Api, error)
	Delete(name string, namespace string) error
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=apiConv -output=automock -outpkg=automock -case=underscore
type apiConv interface {
	ToGQL(in *v1alpha2.Api) *gqlschema.API
	ToGQLs(in []*v1alpha2.Api) []gqlschema.API
	ToApi(name string, namespace string, in gqlschema.APIInput) *v1alpha2.Api
}
