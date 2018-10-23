package istio

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/metadata-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/istio/mocks"
	"github.com/kyma-project/kyma/components/metadata-service/pkg/apis/istio/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var config = RepositoryConfig{Namespace: "testns"}

func TestRepository_Create(t *testing.T) {

	t.Run("should create denier", func(t *testing.T) {
		// given
		expected := &v1alpha2.Denier{
			ObjectMeta: v1.ObjectMeta{
				Name: "re-test-uuid1",
				Labels: map[string]string{
					k8sconsts.LabelRemoteEnvironment: "re",
					k8sconsts.LabelServiceId:         "sid",
				},
			},
			Spec: &v1alpha2.DenierSpec{
				Status: &v1alpha2.DenierStatus{
					Code:    7,
					Message: "Not allowed",
				},
			},
		}

		denierInterface := new(mocks.DenierInterface)
		denierInterface.On("Create", expected).Return(nil, nil)

		repository := NewRepository(nil, nil, denierInterface, config)

		// when
		err := repository.CreateDenier("re", "sid", "re-test-uuid1")

		// then
		assert.NoError(t, err)
		denierInterface.AssertExpectations(t)
	})

	t.Run("should handle error when creating denier", func(t *testing.T) {
		// given
		denierInterface := new(mocks.DenierInterface)
		denierInterface.On("Create", mock.AnythingOfType("*v1alpha2.Denier")).
			Return(nil, errors.New("some error"))

		repository := NewRepository(nil, nil, denierInterface, config)

		// when
		err := repository.CreateDenier("re", "sid", "re-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should create checknothing", func(t *testing.T) {
		// given
		expected := &v1alpha2.Checknothing{
			ObjectMeta: v1.ObjectMeta{
				Name: "re-test-uuid1",
				Labels: map[string]string{
					k8sconsts.LabelRemoteEnvironment: "re",
					k8sconsts.LabelServiceId:         "sid",
				},
			},
		}

		checknothingInterface := new(mocks.ChecknothingInterface)
		checknothingInterface.On("Create", expected).Return(nil, nil)

		repository := NewRepository(nil, checknothingInterface, nil, config)

		// when
		err := repository.CreateCheckNothing("re", "sid", "re-test-uuid1")

		// then
		assert.NoError(t, err)
		checknothingInterface.AssertExpectations(t)
	})

	t.Run("should handle error when creating checknothing", func(t *testing.T) {
		// given
		checknothingInterface := new(mocks.ChecknothingInterface)
		checknothingInterface.On("Create", mock.AnythingOfType("*v1alpha2.Checknothing")).
			Return(nil, errors.New("some error"))

		repository := NewRepository(nil, checknothingInterface, nil, config)

		// when
		err := repository.CreateCheckNothing("re", "sid", "re-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should create rule", func(t *testing.T) {
		// given
		expected := &v1alpha2.Rule{
			ObjectMeta: v1.ObjectMeta{
				Name: "re-test-uuid1",
				Labels: map[string]string{
					k8sconsts.LabelRemoteEnvironment: "re",
					k8sconsts.LabelServiceId:         "sid",
				},
			},
			Spec: &v1alpha2.RuleSpec{
				Match: `(destination.service.host == "re-test-uuid1.testns.svc.cluster.local") && (source.labels["re-test-uuid1"] != "true")`,
				Actions: []v1alpha2.RuleAction{{
					Handler:   "re-test-uuid1.denier",
					Instances: []string{"re-test-uuid1.checknothing"},
				}},
			},
		}

		ruleInterface := new(mocks.RuleInterface)
		ruleInterface.On("Create", expected).Return(nil, nil)

		repository := NewRepository(ruleInterface, nil, nil, config)

		// when
		err := repository.CreateRule("re", "sid", "re-test-uuid1")

		// then
		assert.NoError(t, err)
		ruleInterface.AssertExpectations(t)
	})

	t.Run("should handle error when creating rule", func(t *testing.T) {
		// given
		ruleInterface := new(mocks.RuleInterface)
		ruleInterface.On("Create", mock.AnythingOfType("*v1alpha2.Rule")).
			Return(nil, errors.New("some error"))

		repository := NewRepository(ruleInterface, nil, nil, config)

		// when
		err := repository.CreateRule("re", "sid", "re-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})
}

func TestRepository_Upsert(t *testing.T) {

	t.Run("should upsert denier", func(t *testing.T) {
		// given
		expected := &v1alpha2.Denier{
			ObjectMeta: v1.ObjectMeta{
				Name: "re-test-uuid1",
				Labels: map[string]string{
					k8sconsts.LabelRemoteEnvironment: "re",
					k8sconsts.LabelServiceId:         "sid",
				},
			},
			Spec: &v1alpha2.DenierSpec{
				Status: &v1alpha2.DenierStatus{
					Code:    7,
					Message: "Not allowed",
				},
			},
		}

		denierInterface := new(mocks.DenierInterface)
		denierInterface.On("Create", expected).Return(nil, nil)

		repository := NewRepository(nil, nil, denierInterface, config)

		// when
		err := repository.UpsertDenier("re", "sid", "re-test-uuid1")

		// then
		assert.NoError(t, err)
		denierInterface.AssertExpectations(t)
	})

	t.Run("should handle already exists error when upserting denier", func(t *testing.T) {
		// given
		denierInterface := new(mocks.DenierInterface)
		denierInterface.On("Create", mock.AnythingOfType("*v1alpha2.Denier")).
			Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{}, ""))

		repository := NewRepository(nil, nil, denierInterface, config)

		// when
		err := repository.UpsertDenier("re", "sid", "re-test-uuid1")

		// then
		assert.NoError(t, err)
		denierInterface.AssertExpectations(t)
	})

	t.Run("should handle error when upserting denier", func(t *testing.T) {
		// given
		denierInterface := new(mocks.DenierInterface)
		denierInterface.On("Create", mock.AnythingOfType("*v1alpha2.Denier")).
			Return(nil, errors.New("some error"))

		repository := NewRepository(nil, nil, denierInterface, config)

		// when
		err := repository.UpsertDenier("re", "sid", "re-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should upsert checknothing", func(t *testing.T) {
		// given
		expected := &v1alpha2.Checknothing{
			ObjectMeta: v1.ObjectMeta{
				Name: "re-test-uuid1",
				Labels: map[string]string{
					k8sconsts.LabelRemoteEnvironment: "re",
					k8sconsts.LabelServiceId:         "sid",
				},
			},
		}

		checknothingInterface := new(mocks.ChecknothingInterface)
		checknothingInterface.On("Create", expected).Return(nil, nil)

		repository := NewRepository(nil, checknothingInterface, nil, config)

		// when
		err := repository.UpsertCheckNothing("re", "sid", "re-test-uuid1")

		// then
		assert.NoError(t, err)
		checknothingInterface.AssertExpectations(t)
	})

	t.Run("should handle already exists error when upserting checknothing", func(t *testing.T) {
		// given
		checknothingInterface := new(mocks.ChecknothingInterface)
		checknothingInterface.On("Create", mock.AnythingOfType("*v1alpha2.Checknothing")).
			Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{}, ""))

		repository := NewRepository(nil, checknothingInterface, nil, config)

		// when
		err := repository.UpsertCheckNothing("re", "sid", "re-test-uuid1")

		// then
		assert.NoError(t, err)
		checknothingInterface.AssertExpectations(t)
	})

	t.Run("should handle error when upserting checknothing", func(t *testing.T) {
		// given
		checknothingInterface := new(mocks.ChecknothingInterface)
		checknothingInterface.On("Create", mock.AnythingOfType("*v1alpha2.Checknothing")).
			Return(nil, errors.New("some error"))

		repository := NewRepository(nil, checknothingInterface, nil, config)

		// when
		err := repository.UpsertCheckNothing("re", "sid", "re-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should upsert rule", func(t *testing.T) {
		// given
		expected := &v1alpha2.Rule{
			ObjectMeta: v1.ObjectMeta{
				Name: "re-test-uuid1",
				Labels: map[string]string{
					k8sconsts.LabelRemoteEnvironment: "re",
					k8sconsts.LabelServiceId:         "sid",
				},
			},
			Spec: &v1alpha2.RuleSpec{
				Match: `(destination.service.host == "re-test-uuid1.testns.svc.cluster.local") && (source.labels["re-test-uuid1"] != "true")`,
				Actions: []v1alpha2.RuleAction{{
					Handler:   "re-test-uuid1.denier",
					Instances: []string{"re-test-uuid1.checknothing"},
				}},
			},
		}

		ruleInterface := new(mocks.RuleInterface)
		ruleInterface.On("Create", expected).Return(nil, nil)

		repository := NewRepository(ruleInterface, nil, nil, config)

		// when
		err := repository.UpsertRule("re", "sid", "re-test-uuid1")

		// then
		assert.NoError(t, err)
		ruleInterface.AssertExpectations(t)
	})

	t.Run("should handle already exists error when upserting rule", func(t *testing.T) {
		// given
		ruleInterface := new(mocks.RuleInterface)
		ruleInterface.On("Create", mock.AnythingOfType("*v1alpha2.Rule")).
			Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{}, ""))

		repository := NewRepository(ruleInterface, nil, nil, config)

		// when
		err := repository.UpsertRule("re", "sid", "re-test-uuid1")

		// then
		assert.NoError(t, err)
		ruleInterface.AssertExpectations(t)
	})

	t.Run("should handle error when upserting rule", func(t *testing.T) {
		// given
		ruleInterface := new(mocks.RuleInterface)
		ruleInterface.On("Create", mock.AnythingOfType("*v1alpha2.Rule")).
			Return(nil, errors.New("some error"))

		repository := NewRepository(ruleInterface, nil, nil, config)

		// when
		err := repository.UpsertRule("re", "sid", "re-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})
}

func TestRepository_Delete(t *testing.T) {

	t.Run("should delete denier", func(t *testing.T) {
		// given
		denierInterface := new(mocks.DenierInterface)
		denierInterface.On("Delete", "re-test-uuid1", (*v1.DeleteOptions)(nil)).Return(nil)

		repository := NewRepository(nil, nil, denierInterface, config)

		// when
		err := repository.DeleteDenier("re-test-uuid1")

		// then
		assert.NoError(t, err)
		denierInterface.AssertExpectations(t)
	})

	t.Run("should handle error when deleting denier", func(t *testing.T) {
		// given
		denierInterface := new(mocks.DenierInterface)
		denierInterface.On("Delete", "re-test-uuid1", (*v1.DeleteOptions)(nil)).
			Return(errors.New("some error"))

		repository := NewRepository(nil, nil, denierInterface, config)

		// when
		err := repository.DeleteDenier("re-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should ignore not found error when deleting denier", func(t *testing.T) {
		// given
		denierInterface := new(mocks.DenierInterface)
		denierInterface.On("Delete", "re-test-uuid1", (*v1.DeleteOptions)(nil)).
			Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := NewRepository(nil, nil, denierInterface, config)

		// when
		err := repository.DeleteDenier("re-test-uuid1")

		// then
		assert.NoError(t, err)
	})

	t.Run("should delete checknothing", func(t *testing.T) {
		// given
		checknothingInterface := new(mocks.ChecknothingInterface)
		checknothingInterface.On("Delete", "re-test-uuid1", (*v1.DeleteOptions)(nil)).Return(nil)

		repository := NewRepository(nil, checknothingInterface, nil, config)

		// when
		err := repository.DeleteCheckNothing("re-test-uuid1")

		// then
		assert.NoError(t, err)
		checknothingInterface.AssertExpectations(t)
	})

	t.Run("should handle error when deleting checknothing", func(t *testing.T) {
		// given
		checknothingInterface := new(mocks.ChecknothingInterface)
		checknothingInterface.On("Delete", "re-test-uuid1", (*v1.DeleteOptions)(nil)).
			Return(errors.New("some error"))

		repository := NewRepository(nil, checknothingInterface, nil, config)

		// when
		err := repository.DeleteCheckNothing("re-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should ignore not found error when deleting checknothing", func(t *testing.T) {
		// given
		checknothingInterface := new(mocks.ChecknothingInterface)
		checknothingInterface.On("Delete", "re-test-uuid1", (*v1.DeleteOptions)(nil)).
			Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := NewRepository(nil, checknothingInterface, nil, config)

		// when
		err := repository.DeleteCheckNothing("re-test-uuid1")

		// then
		assert.NoError(t, err)
	})

	t.Run("should delete rule", func(t *testing.T) {
		// given
		ruleInterface := new(mocks.RuleInterface)
		ruleInterface.On("Delete", "re-test-uuid1", (*v1.DeleteOptions)(nil)).Return(nil)

		repository := NewRepository(ruleInterface, nil, nil, config)

		// when
		err := repository.DeleteRule("re-test-uuid1")

		// then
		assert.NoError(t, err)
		ruleInterface.AssertExpectations(t)
	})

	t.Run("should handle error when deleting rule", func(t *testing.T) {
		// given
		ruleInterface := new(mocks.RuleInterface)
		ruleInterface.On("Delete", "re-test-uuid1", (*v1.DeleteOptions)(nil)).
			Return(errors.New("some error"))

		repository := NewRepository(ruleInterface, nil, nil, config)

		// when
		err := repository.DeleteRule("re-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should ignore not found error when deleting rule", func(t *testing.T) {
		// given
		ruleInterface := new(mocks.RuleInterface)
		ruleInterface.On("Delete", "re-test-uuid1", (*v1.DeleteOptions)(nil)).
			Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := NewRepository(ruleInterface, nil, nil, config)

		// when
		err := repository.DeleteRule("re-test-uuid1")

		// then
		assert.NoError(t, err)
	})
}
