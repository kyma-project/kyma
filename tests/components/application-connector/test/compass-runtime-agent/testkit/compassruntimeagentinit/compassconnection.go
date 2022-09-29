package compassruntimeagentinit

import (
	"context"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type compassConnectionCRConfiguration struct {
	compassConnectionInterface CompassConnectionInterface
}

//go:generate mockery --name=CompassConnectionInterface
type CompassConnectionInterface interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.CompassConnection, error)
	Create(ctx context.Context, compassConnection *v1alpha1.CompassConnection, opts v1.CreateOptions) (*v1alpha1.CompassConnection, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
}

func NewCompassConnectionCRConfiguration(compassConnectionInterface CompassConnectionInterface) compassConnectionCRConfiguration {
	return compassConnectionCRConfiguration{
		compassConnectionInterface: compassConnectionInterface,
	}
}

func (cc compassConnectionCRConfiguration) Do() (types.RollbackFunc, error) {

	compassConnectionCR, err := cc.compassConnectionInterface.Get(context.TODO(), "compass-connection", meta.GetOptions{})
	if err != nil {
		return nil, err
	}

	err = cc.compassConnectionInterface.Delete(context.TODO(), "compass-connection", meta.DeleteOptions{})
	if err != nil {
		return nil, err
	}
	compassConnectionCR.ResourceVersion = ""

	compassConnectionCRBackup := compassConnectionCR.DeepCopy()
	compassConnectionCRBackup.ObjectMeta.Name = "compass-connection-backup"
	_, err = cc.compassConnectionInterface.Create(context.TODO(), compassConnectionCRBackup, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return func() error {
		_, err = cc.compassConnectionInterface.Create(context.TODO(), compassConnectionCR, meta.CreateOptions{})
		if err != nil {
			return err
		}

		err = cc.compassConnectionInterface.Delete(context.TODO(), "compass-connection-backup", meta.DeleteOptions{})
		if err != nil {
			return err
		}

		return err
	}, nil
}
