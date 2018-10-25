package secrets

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/secrets/mocks"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestRepository_Get(t *testing.T) {
	t.Run("should get given secret", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock, "default-ec")

		secret := makeSecret("new-secret", "CLIENT_ID", "CLIENT_SECRET", "secretId", "default-ec")
		secretsManagerMock.On("Get", "new-secret", metav1.GetOptions{}).Return(secret, nil)

		// when
		clientId, clientSecret, err := repository.Get("new-secret")

		// then
		assert.NoError(t, err)
		assert.Equal(t, "CLIENT_ID", clientId)
		assert.Equal(t, "CLIENT_SECRET", clientSecret)

		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return an error in case fetching fails", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock, "default-ec")

		secretsManagerMock.On("Get", "secret-name", metav1.GetOptions{}).Return(
			nil,
			errors.New("some error"))

		// when
		clientId, clientSecret, err := repository.Get("secret-name")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		assert.Equal(t, "", clientId)
		assert.Equal(t, "", clientSecret)

		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return not found if secret does not exist", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock, "default-ec")

		secretsManagerMock.On("Get", "secret-name", metav1.GetOptions{}).Return(
			nil,
			k8serrors.NewNotFound(schema.GroupResource{},
				""))

		// when
		clientId, clientSecret, err := repository.Get("secret-name")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
		assert.NotEmpty(t, err.Error())

		assert.Equal(t, "", clientId)
		assert.Equal(t, "", clientSecret)

		secretsManagerMock.AssertExpectations(t)
	})
}

func makeSecret(name, clientID, clientSecret, serviceID, remoteEnvironment string) *v1.Secret {
	secretMap := make(map[string][]byte)
	secretMap[ClientIDKey] = []byte(clientID)
	secretMap[ClientSecretKey] = []byte(clientSecret)

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelRemoteEnvironment: remoteEnvironment,
				k8sconsts.LabelServiceId:         serviceID,
			},
		},
		Data: secretMap,
	}
}
