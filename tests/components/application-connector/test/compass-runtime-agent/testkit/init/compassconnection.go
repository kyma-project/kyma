package init

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
	Update(ctx context.Context, compassConnection *v1alpha1.CompassConnection, opts v1.UpdateOptions) (*v1alpha1.CompassConnection, error)
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
		return nil, errors.Wrap(err, "failed to get Compass Connection CR")
	}

	err = cc.compassConnectionInterface.Delete(context.TODO(), "compass-connection", meta.DeleteOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete Compass Connection CR")
	}
	compassConnectionCR.ResourceVersion = ""

	compassConnectionCRBackup := compassConnectionCR.DeepCopy()
	compassConnectionCRBackup.ObjectMeta.Name = "compass-connection-backup"
	_, err = cc.compassConnectionInterface.Create(context.TODO(), compassConnectionCRBackup, meta.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Compass Connection CR")
	}

	return cc.getRollbackFunc(compassConnectionCRBackup), nil
}

func (cc compassConnectionCRConfiguration) getRollbackFunc(compassConnectionCRBackup *v1alpha1.CompassConnection) types.RollbackFunc {
	return func() error {
		var result *multierror.Error

		err := retry.Do(func() error {
			restoredCompassConnection, err := cc.compassConnectionInterface.Get(context.TODO(), "compass-connection", meta.GetOptions{})
			if err != nil {
				return err
			}

			restoredCompassConnection.Spec = compassConnectionCRBackup.Spec
			restoredCompassConnection.Status = compassConnectionCRBackup.Status

			_, err = cc.compassConnectionInterface.Update(context.TODO(), restoredCompassConnection, meta.UpdateOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to update Compass Connection CR")
			}
			return err
		})

		if err != nil {
			multierror.Append(result, err)
		}

		err = retry.Do(func() error {
			err = cc.compassConnectionInterface.Delete(context.TODO(), "compass-connection-backup", meta.DeleteOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return nil
				}
				return errors.Wrap(err, "failed to delete Compass Connection CR")
			}

			return nil
		})

		if err != nil {
			multierror.Append(result, err)
		}

		return result.ErrorOrNil()
	}
}
