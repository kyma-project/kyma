package apicontroller

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
)

//go:generate mockery -name=apiLister -output=automock -outpkg=automock -case=underscore
type apiLister interface {
	List(namespace string, serviceName *string, hostname *string) ([]*v1alpha2.Api, error)
	Find(name string, namespace string) (*v1alpha2.Api, error)
	Create(name string, namespace string, hostname string, serviceName string, servicePort int, authenticationType string, jwksUri string, issuer string, disableIstioAuthPolicyMTLS *bool, authenticationEnabled *bool) (*v1alpha2.Api, error)
	Delete(name string, namespace string) error
}
