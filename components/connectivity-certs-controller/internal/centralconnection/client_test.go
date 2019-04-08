package centralconnection

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/centralconnection/mocks"
)

const (
	connectionName = "my-connection"
)

func TestCentralConnectionClient_Upsert(t *testing.T) {

	newConnection := &v1alpha1.CentralConnection{
		ObjectMeta: v1.ObjectMeta{
			Name: connectionName,
		},
		Spec: v1alpha1.CentralConnectionSpec{
			ManagementInfoURL: managementInfoURL,
		},
	}

	t.Run("should create Central Connection if it does not exist", func(t *testing.T) {
		// given
		ccManager := &mocks.Manager{}
		ccManager.On("Update", newConnection).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		ccManager.On("Create", newConnection).Return(newConnection, nil)

		ccClient := NewCentralConnectionClient(ccManager)

		// when
		connection, err := ccClient.Upsert(newConnection)

		// then
		require.NoError(t, err)
		assert.Equal(t, newConnection, connection)

		ccManager.AssertExpectations(t)
	})

	t.Run("should update Central Connection if it exist", func(t *testing.T) {
		// given
		ccManager := &mocks.Manager{}
		ccManager.On("Update", newConnection).Return(newConnection, nil)

		ccClient := NewCentralConnectionClient(ccManager)

		// when
		connection, err := ccClient.Upsert(newConnection)

		// then
		require.NoError(t, err)
		assert.Equal(t, newConnection, connection)

		ccManager.AssertExpectations(t)
	})

	t.Run("should return error when failed to update connection", func(t *testing.T) {
		// given
		ccManager := &mocks.Manager{}
		ccManager.On("Update", newConnection).Return(nil, k8serrors.NewInternalError(errors.New("error")))

		ccClient := NewCentralConnectionClient(ccManager)

		// when
		_, err := ccClient.Upsert(newConnection)

		// then
		require.Error(t, err)

		ccManager.AssertExpectations(t)
	})

	t.Run("should return error when failed to create connection", func(t *testing.T) {
		// given
		ccManager := &mocks.Manager{}
		ccManager.On("Update", newConnection).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		ccManager.On("Create", newConnection).Return(nil, k8serrors.NewInternalError(errors.New("error")))

		ccClient := NewCentralConnectionClient(ccManager)

		// when
		_, err := ccClient.Upsert(newConnection)

		// then
		require.Error(t, err)

		ccManager.AssertExpectations(t)
	})

}
