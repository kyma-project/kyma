package centralconnection

import (
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Manager interface {
	Create(*v1alpha1.CentralConnection) (*v1alpha1.CentralConnection, error)
	Update(*v1alpha1.CentralConnection) (*v1alpha1.CentralConnection, error)
	Get(name string, options v1.GetOptions) (*v1alpha1.CentralConnection, error)
}

type Client interface {
	Upsert(name string, spec v1alpha1.CentralConnectionSpec) (*v1alpha1.CentralConnection, error)
}

type centralConnectionClient struct {
	manager Manager
}

func NewCentralConnectionClient(manager Manager) Client {
	return &centralConnectionClient{
		manager: manager,
	}
}

func (c centralConnectionClient) Upsert(name string, spec v1alpha1.CentralConnectionSpec) (*v1alpha1.CentralConnection, error) {
	existingConnection, err := c.manager.Get(name, v1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}

		return c.create(name, spec)
	}

	existingConnection.Spec = spec
	existingConnection.Status = v1alpha1.CentralConnectionStatus{}

	return c.manager.Update(existingConnection)
}

func (c centralConnectionClient) create(name string, spec v1alpha1.CentralConnectionSpec) (*v1alpha1.CentralConnection, error) {
	centralConnection := &v1alpha1.CentralConnection{
		ObjectMeta: v1.ObjectMeta{Name: name},
		Spec:       spec,
	}

	return c.manager.Create(centralConnection)
}
