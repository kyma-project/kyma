package init

import (
	"context"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCompassConnectionConfigurator(t *testing.T) {
	t.Run("should delete CompassConnection CR and restore it when RollbackFunction is called", func(t *testing.T) {
		// given
		compassConnectionCRFake := fake.NewSimpleClientset().CompassV1alpha1().CompassConnections()
		compassConnection := &v1alpha1.CompassConnection{
			ObjectMeta: meta.ObjectMeta{
				Name: "compass-connection",
			},
			TypeMeta: meta.TypeMeta{
				Kind:       "CompassConnection",
				APIVersion: "v1alpha",
			},
		}

		_, err := compassConnectionCRFake.Create(context.TODO(), compassConnection, meta.CreateOptions{})
		require.NoError(t, err)

		// when
		configurator := NewCompassConnectionCRConfiguration(compassConnectionCRFake)
		rollbackFunc, err := configurator.Do()

		// then
		require.NoError(t, err)
		_, err = compassConnectionCRFake.Get(context.TODO(), "compass-connection", meta.GetOptions{})
		require.Error(t, err)
		require.True(t, k8serrors.IsNotFound(err))

		_, err = compassConnectionCRFake.Get(context.TODO(), "compass-connection-backup", meta.GetOptions{})
		require.NoError(t, err)

		_, err = compassConnectionCRFake.Create(context.TODO(), compassConnection, meta.CreateOptions{})
		require.NoError(t, err)
		// when
		err = rollbackFunc()

		// then
		require.NoError(t, err)
		_, err = compassConnectionCRFake.Get(context.TODO(), "compass-connection", meta.GetOptions{})
		require.NoError(t, err)
	})

	t.Run("should not return error when Compass Connection CR doesn't exist", func(t *testing.T) {
		// given
		compassConnectionCRFake := fake.NewSimpleClientset().CompassV1alpha1().CompassConnections()
		compassConnection := &v1alpha1.CompassConnection{
			ObjectMeta: meta.ObjectMeta{
				Name: "compass-connection",
			},
			TypeMeta: meta.TypeMeta{
				Kind:       "CompassConnection",
				APIVersion: "v1alpha",
			},
		}

		_, err := compassConnectionCRFake.Create(context.TODO(), compassConnection, meta.CreateOptions{})
		require.NoError(t, err)

		// when
		configurator := NewCompassConnectionCRConfiguration(compassConnectionCRFake)
		rollbackFunc, err := configurator.Do()

		// then
		require.NoError(t, err)
		_, err = compassConnectionCRFake.Create(context.TODO(), compassConnection, meta.CreateOptions{})
		require.NoError(t, err)

		// when
		err = rollbackFunc()

		// then
		require.Error(t, err)
	})

	// TODO: consider a case when rollback function fails
}
