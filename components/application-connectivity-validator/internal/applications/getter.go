package applications

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	v1alpha12 "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=Getter
type Getter interface {
	Get(name string) (*v1alpha1.Application, error)
}

type applicationGetter struct {
	applicationInterface v1alpha12.ApplicationInterface
}

func NewGetter(applicationInterface v1alpha12.ApplicationInterface) Getter {
	return &applicationGetter{
		applicationInterface: applicationInterface,
	}
}

func (ai *applicationGetter) Get(applicationName string) (*v1alpha1.Application, error) {
	return ai.applicationInterface.Get(applicationName, metav1.GetOptions{})
}
