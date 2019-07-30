package applications

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	v1alpha12 "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type manager struct {
	applicationsInterface v1alpha12.ApplicationInterface
}

func NewManager(applicationsInterface v1alpha12.ApplicationInterface) Manager {
	return manager{
		applicationsInterface: applicationsInterface,
	}
}

func (m manager) Create(application *v1alpha1.Application) (*v1alpha1.Application, error) {
	return m.applicationsInterface.Create(application)
}

func (m manager) Update(application *v1alpha1.Application) (*v1alpha1.Application, error) {
	return m.applicationsInterface.Update(application)
}

func (m manager) Delete(name string, options *metav1.DeleteOptions) error {
	return m.applicationsInterface.Delete(name, options)
}

func (m manager) Get(name string, options metav1.GetOptions) (*v1alpha1.Application, error) {
	return m.applicationsInterface.Get(name, options)
}

func (m manager) List(opts metav1.ListOptions) (*v1alpha1.ApplicationList, error) {
	return m.applicationsInterface.List(opts)
}
