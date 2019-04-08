package centralconnection

import (
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

type Manager interface {
	Create(*v1alpha1.CentralConnection) (*v1alpha1.CentralConnection, error)
	Update(*v1alpha1.CentralConnection) (*v1alpha1.CentralConnection, error)
}

type Client interface {
	Upsert(*v1alpha1.CentralConnection) (*v1alpha1.CentralConnection, error)
}

type centralConnectionClient struct {
	manager Manager
}

func NewCentralConnectionClient(manager Manager) Client {
	return &centralConnectionClient{
		manager: manager,
	}
}

func (c centralConnectionClient) Upsert(connection *v1alpha1.CentralConnection) (*v1alpha1.CentralConnection, error) {
	newConnection, err := c.manager.Update(connection)
	if err == nil {
		return newConnection, nil
	}

	if !k8serrors.IsNotFound(err) {
		return nil, err
	}

	newConnection, err = c.manager.Create(connection)
	if err != nil {
		return nil, err
	}

	return newConnection, nil
}
