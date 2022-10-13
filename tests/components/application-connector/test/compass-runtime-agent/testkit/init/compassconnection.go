package init

import (
	"context"
	"github.com/avast/retry-go"
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

const (
	ConnectionCRName       = "compass-connection"
	ConnectionBackupCRName = "compass-connection-backup"
)

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

	backupRollbackFunc, err := cc.backup()
	if err != nil {
		return nil, err
	}

	deleteRollbackFunc, err := cc.delete()
	if err != nil {
		return backupRollbackFunc, err
	}

	return newRollbackFunc(deleteRollbackFunc, backupRollbackFunc), nil
}
func (cc compassConnectionCRConfiguration) backup() (types.RollbackFunc, error) {
	compassConnectionCR, err := cc.compassConnectionInterface.Get(context.TODO(), ConnectionCRName, meta.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get Compass Connection CR")
	}

	compassConnectionCR.ResourceVersion = ""

	compassConnectionCRBackup := compassConnectionCR.DeepCopy()
	compassConnectionCRBackup.ObjectMeta.Name = "compass-connection-backup"
	_, err = cc.compassConnectionInterface.Create(context.TODO(), compassConnectionCRBackup, meta.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Compass Connection CR")
	}

	rollbackFunc := func() error {
		return retry.Do(func() error {
			err = cc.compassConnectionInterface.Delete(context.TODO(), "compass-connection-backup", meta.DeleteOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return nil
				}
				return errors.Wrap(err, "failed to delete Compass Connection CR")
			}

			return nil
		})
	}

	return rollbackFunc, nil
}

func (cc compassConnectionCRConfiguration) delete() (types.RollbackFunc, error) {
	err := cc.compassConnectionInterface.Delete(context.TODO(), ConnectionCRName, meta.DeleteOptions{})

	if err != nil {
		return nil, errors.Wrap(err, "failed to delete Compass Connection CR")
	}

	rollbackFunc := func() error {
		return retry.Do(func() error {
			restoredCompassConnection, err := cc.compassConnectionInterface.Get(context.TODO(), ConnectionCRName, meta.GetOptions{})
			if err != nil {
				return err
			}

			compassConnectionCRBackup, err := cc.compassConnectionInterface.Get(context.TODO(), ConnectionBackupCRName, meta.GetOptions{})
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
	}

	return rollbackFunc, nil
}
