package apicontroller

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
)

//go:generate mockery -name=apiLister -output=automock -outpkg=automock -case=underscore
type apiLister interface {
	List(environment string, serviceName *string, hostname *string) ([]*v1alpha2.Api, error)
}
