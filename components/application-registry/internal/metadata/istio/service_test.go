package istio

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/istio/mocks"
	"github.com/stretchr/testify/assert"
)

func TestService_Create(t *testing.T) {

	t.Run("should create authorization policy", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}
		repository.On("CreateAuthorizationPolicy", "app", applicationUID, "sid", "testsvc").Return(nil)

		service := NewService(repository)

		// when
		err := service.Create("app", "appUID", "sid", "testsvc")

		// then
		assert.NoError(t, err)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when creating authorization policy", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}
		repository.On("CreateAuthorizationPolicy", "app", applicationUID, "sid", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Create("app", "appUID", "sid", "testsvc")

		// then
		assert.Error(t, err)
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
		repository.AssertExpectations(t)
	})
}

func TestService_Upsert(t *testing.T) {

	t.Run("should upsert authorization policy", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}
		repository.On("UpsertAuthorizationPolicy", "app", applicationUID, "sid", "testsvc").Return(nil)

		service := NewService(repository)

		// when
		err := service.Upsert("app", "appUID", "sid", "testsvc")

		// then
		assert.NoError(t, err)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when upserting authorization policy", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}
		repository.On("UpsertAuthorizationPolicy", "app", applicationUID, "sid", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Upsert("app", "appUID", "sid", "testsvc")

		// then
		assert.Error(t, err)
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
		repository.AssertExpectations(t)
	})
}

func TestService_Delete(t *testing.T) {

	t.Run("should delete authorization policy", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}
		repository.On("DeleteAuthorizationPolicy", "testsvc").Return(nil)

		service := NewService(repository)

		// when
		err := service.Delete("testsvc")

		// then
		assert.NoError(t, err)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when deleting authorization policy", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}
		repository.On("DeleteAuthorizationPolicy", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Delete("testsvc")

		// then
		assert.Error(t, err)
		repository.AssertExpectations(t)
	})
}
