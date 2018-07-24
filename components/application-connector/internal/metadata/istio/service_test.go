package istio

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-connector/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/istio/mocks"
	"github.com/stretchr/testify/assert"
)

func TestService_Create(t *testing.T) {

	t.Run("should create denier, checknothing, rule", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("CreateDenier", "re", "sid", "testsvc").Return(nil)
		repository.On("CreateCheckNothing", "re", "sid", "testsvc").Return(nil)
		repository.On("CreateRule", "re", "sid", "testsvc").Return(nil)

		service := NewService(repository)

		// when
		err := service.Create("re", "sid", "testsvc")

		// then
		assert.NoError(t, err)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when creating denier", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("CreateDenier", "re", "sid", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Create("re", "sid", "testsvc")

		// then
		assert.Error(t, err)
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when creating checknothing", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("CreateDenier", "re", "sid", "testsvc").Return(nil)
		repository.On("CreateCheckNothing", "re", "sid", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Create("re", "sid", "testsvc")

		// then
		assert.Error(t, err)
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when creating rule", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("CreateDenier", "re", "sid", "testsvc").Return(nil)
		repository.On("CreateCheckNothing", "re", "sid", "testsvc").Return(nil)
		repository.On("CreateRule", "re", "sid", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Create("re", "sid", "testsvc")

		// then
		assert.Error(t, err)
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
		repository.AssertExpectations(t)
	})
}

func TestService_Upsert(t *testing.T) {

	t.Run("should upsert denier, checknothing, rule", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("UpsertDenier", "re", "sid", "testsvc").Return(nil)
		repository.On("UpsertCheckNothing", "re", "sid", "testsvc").Return(nil)
		repository.On("UpsertRule", "re", "sid", "testsvc").Return(nil)

		service := NewService(repository)

		// when
		err := service.Upsert("re", "sid", "testsvc")

		// then
		assert.NoError(t, err)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when upserting denier", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("UpsertDenier", "re", "sid", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Upsert("re", "sid", "testsvc")

		// then
		assert.Error(t, err)
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when upserting checknothing", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("UpsertDenier", "re", "sid", "testsvc").Return(nil)
		repository.On("UpsertCheckNothing", "re", "sid", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Upsert("re", "sid", "testsvc")

		// then
		assert.Error(t, err)
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when upserting rule", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("UpsertDenier", "re", "sid", "testsvc").Return(nil)
		repository.On("UpsertCheckNothing", "re", "sid", "testsvc").Return(nil)
		repository.On("UpsertRule", "re", "sid", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Upsert("re", "sid", "testsvc")

		// then
		assert.Error(t, err)
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
		repository.AssertExpectations(t)
	})
}

func TestService_Delete(t *testing.T) {

	t.Run("should delete denier, checknothing, rule", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("DeleteDenier", "testsvc").Return(nil)
		repository.On("DeleteCheckNothing", "testsvc").Return(nil)
		repository.On("DeleteRule", "testsvc").Return(nil)

		service := NewService(repository)

		// when
		err := service.Delete("testsvc")

		// then
		assert.NoError(t, err)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when deleting denier", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("DeleteDenier", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Delete("testsvc")

		// then
		assert.Error(t, err)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when deleting checknothing", func(t *testing.T) {

		// given
		repository := &mocks.Repository{}

		repository.On("DeleteDenier", "testsvc").Return(nil)
		repository.On("DeleteCheckNothing", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Delete("testsvc")

		// then
		assert.Error(t, err)
		repository.AssertExpectations(t)
	})

	t.Run("should handle errors when deleting rule", func(t *testing.T) {
		// given
		repository := &mocks.Repository{}

		repository.On("DeleteDenier", "testsvc").Return(nil)
		repository.On("DeleteCheckNothing", "testsvc").Return(nil)
		repository.On("DeleteRule", "testsvc").Return(apperrors.Internal(""))

		service := NewService(repository)

		// when
		err := service.Delete("testsvc")

		// then
		assert.Error(t, err)
		repository.AssertExpectations(t)
	})

}
