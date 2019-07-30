package applications

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=Manager
// Manager contains operations for managing Application CRD
type Manager interface {
	Create(*v1alpha1.Application) (*v1alpha1.Application, error)
	Update(*v1alpha1.Application) (*v1alpha1.Application, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string, options metav1.GetOptions) (*v1alpha1.Application, error)
	List(opts metav1.ListOptions) (*v1alpha1.ApplicationList, error)
}

type Credentials struct {
	Type              string
	SecretName        string
	AuthenticationUrl string
	CSRFInfo          *CSRFInfo
	Headers           *map[string][]string
	QueryParameters   *map[string][]string
}

type CSRFInfo struct {
	TokenEndpointURL string
}
