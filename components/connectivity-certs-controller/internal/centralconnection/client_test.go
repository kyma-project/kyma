package centralconnection

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/centralconnection/mocks"
)

const (
	connectionName = "my-connection"
)

func TestCentralConnectionClient_Upsert(t *testing.T) {

	existingConnection := &v1alpha1.CentralConnection{
		ObjectMeta: v1.ObjectMeta{Name: connectionName},
		Spec:       v1alpha1.CentralConnectionSpec{},
	}

	newConnectionSpec := v1alpha1.CentralConnectionSpec{
		ManagementInfoURL: managementInfoURL,
	}

	newConnection := &v1alpha1.CentralConnection{
		ObjectMeta: v1.ObjectMeta{Name: connectionName},
		Spec:       newConnectionSpec,
	}

	t.Run("should create Central Connection if it does not exist", func(t *testing.T) {
		// given
		ccManager := &mocks.Manager{}
		ccManager.On("Get", connectionName, v1.GetOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		ccManager.On("Create", newConnection).Return(newConnection, nil)

		ccClient := NewCentralConnectionClient(ccManager)

		// when
		connection, err := ccClient.Upsert(connectionName, newConnectionSpec)

		// then
		require.NoError(t, err)
		assert.Equal(t, newConnection, connection)

		ccManager.AssertExpectations(t)
	})

	t.Run("should update Central Connection if it exist", func(t *testing.T) {
		// given
		ccManager := &mocks.Manager{}
		ccManager.On("Get", connectionName, v1.GetOptions{}).Return(existingConnection, nil)
		ccManager.On("Update", newConnection).Return(newConnection, nil)

		ccClient := NewCentralConnectionClient(ccManager)

		// when
		connection, err := ccClient.Upsert(connectionName, newConnectionSpec)

		// then
		require.NoError(t, err)
		assert.Equal(t, newConnection, connection)

		ccManager.AssertExpectations(t)
	})

	t.Run("should return error when failed to update connection", func(t *testing.T) {
		// given
		ccManager := &mocks.Manager{}
		ccManager.On("Get", connectionName, v1.GetOptions{}).Return(existingConnection, nil)
		ccManager.On("Update", newConnection).Return(nil, k8serrors.NewInternalError(errors.New("error")))

		ccClient := NewCentralConnectionClient(ccManager)

		// when
		_, err := ccClient.Upsert(connectionName, newConnectionSpec)

		// then
		require.Error(t, err)

		ccManager.AssertExpectations(t)
	})

	t.Run("should return error when failed to create connection", func(t *testing.T) {
		// given
		ccManager := &mocks.Manager{}
		ccManager.On("Get", connectionName, v1.GetOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		ccManager.On("Create", newConnection).Return(nil, errors.New("error"))

		ccClient := NewCentralConnectionClient(ccManager)

		// when
		_, err := ccClient.Upsert(connectionName, newConnectionSpec)

		// then
		require.Error(t, err)

		ccManager.AssertExpectations(t)
	})

	t.Run("should return error when failed to get connection", func(t *testing.T) {
		// given
		ccManager := &mocks.Manager{}
		ccManager.On("Get", connectionName, v1.GetOptions{}).Return(nil, errors.New("error"))

		ccClient := NewCentralConnectionClient(ccManager)

		// when
		_, err := ccClient.Upsert(connectionName, newConnectionSpec)

		// then
		require.Error(t, err)

		ccManager.AssertExpectations(t)
	})

}
