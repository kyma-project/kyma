package authentication_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/idppreset/pkg/apis/authentication/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/authentication"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/authentication/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestIDPPresetResolver_CreateIDPPresetMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		var (
			fixName            = "testName"
			fixIssuer          = "testIssuer"
			fixJwksURI         = "testJwksURI"
			fixIDPPresetObj    = fixIDPPreset()
			fixGQLIDPPresetObj = fixIDPPresetGQL()
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("Create", fixName, fixIssuer, fixJwksURI).Return(fixIDPPresetObj, nil)

		converterMock := automock.NewGQLIDPPresetConverter()
		defer converterMock.AssertExpectations(t)
		converterMock.On("ToGQL", fixIDPPresetObj).Return(fixGQLIDPPresetObj)

		resolver := authentication.NewIDPPresetResolver(svc)
		resolver.SetIDPPresetConverter(converterMock)

		// when
		gotIDP, err := resolver.CreateIDPPresetMutation(context.Background(), fixName, fixIssuer, fixJwksURI)

		// then
		require.NoError(t, err)
		require.NotNil(t, gotIDP)
		assert.Equal(t, fixGQLIDPPresetObj, *gotIDP)
	})

	t.Run("Already exists", func(t *testing.T) {
		// given
		var (
			fixName    = "testName"
			fixIssuer  = "testIssuer"
			fixJwksURI = "testJwksURI"
			fixErr     = apiErrors.NewAlreadyExists(schema.GroupResource{}, "exists")
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("Create", fixName, fixIssuer, fixJwksURI).Return(nil, fixErr)

		resolver := authentication.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.CreateIDPPresetMutation(context.Background(), fixName, fixIssuer, fixJwksURI)

		// then
		require.Error(t, err)
		assert.True(t, gqlerror.IsAlreadyExists(err))
		require.Nil(t, gotIDP)
	})

	t.Run("Error", func(t *testing.T) {
		// given
		var (
			fixName    = "testName"
			fixIssuer  = "testIssuer"
			fixJwksURI = "testJwksURI"
			fixErr     = errors.New("something went wrong")
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("Create", fixName, fixIssuer, fixJwksURI).Return(nil, fixErr)

		resolver := authentication.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.CreateIDPPresetMutation(context.Background(), fixName, fixIssuer, fixJwksURI)

		// then
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		require.Nil(t, gotIDP)
	})
}

func TestIDPPresetResolver_DeleteIDPPresetMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		var (
			fixName            = "testName"
			fixIDPPresetObj    = fixIDPPreset()
			fixGQLIDPPresetObj = fixIDPPresetGQL()
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("Find", fixName).Return(fixIDPPresetObj, nil)
		svc.On("Delete", fixName).Return(nil, nil)

		converterMock := automock.NewGQLIDPPresetConverter()
		defer converterMock.AssertExpectations(t)
		converterMock.On("ToGQL", fixIDPPresetObj).Return(fixGQLIDPPresetObj)

		resolver := authentication.NewIDPPresetResolver(svc)
		resolver.SetIDPPresetConverter(converterMock)

		// when
		gotIDP, err := resolver.DeleteIDPPresetMutation(context.Background(), fixName)

		// then
		require.NoError(t, err)
		require.NotNil(t, gotIDP)
		assert.Equal(t, fixGQLIDPPresetObj, *gotIDP)
	})

	t.Run("Not found", func(t *testing.T) {
		// given
		fixName := "testName"

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("Find", fixName).Return(nil, nil)

		resolver := authentication.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.DeleteIDPPresetMutation(context.Background(), fixName)

		// then
		require.Error(t, err)
		assert.True(t, gqlerror.IsNotFound(err))
		require.Nil(t, gotIDP)
	})

	t.Run("Error", func(t *testing.T) {
		// given
		var (
			fixName         = "testName"
			fixIDPPresetObj = fixIDPPreset()
			fixErr          = errors.New("something went wrong")
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("Find", fixName).Return(fixIDPPresetObj, nil)
		svc.On("Delete", fixName).Return(fixErr)

		resolver := authentication.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.DeleteIDPPresetMutation(context.Background(), fixName)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		require.Nil(t, gotIDP)
	})
}

func TestIDPPresetResolver_IDPPresetQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		var (
			fixName            = "testName"
			fixIDPPresetObj    = fixIDPPreset()
			fixGQLIDPPresetObj = fixIDPPresetGQL()
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("Find", fixName).Return(fixIDPPresetObj, nil)

		converterMock := automock.NewGQLIDPPresetConverter()
		defer converterMock.AssertExpectations(t)
		converterMock.On("ToGQL", fixIDPPresetObj).Return(fixGQLIDPPresetObj)

		resolver := authentication.NewIDPPresetResolver(svc)
		resolver.SetIDPPresetConverter(converterMock)

		// when
		gotIDP, err := resolver.IDPPresetQuery(context.Background(), fixName)

		// then
		require.NoError(t, err)
		require.NotNil(t, gotIDP)
		assert.Equal(t, fixGQLIDPPresetObj, *gotIDP)
	})

	t.Run("Not found", func(t *testing.T) {
		// given
		fixName := "testName"

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("Find", fixName).Return(nil, nil)

		resolver := authentication.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.IDPPresetQuery(context.Background(), fixName)

		// then
		require.NoError(t, err)
		require.Nil(t, gotIDP)
	})

	t.Run("Error", func(t *testing.T) {
		// given
		var (
			fixName = "testName"
			fixErr  = errors.New("something went wrong")
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("Find", fixName).Return(nil, fixErr)

		resolver := authentication.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.IDPPresetQuery(context.Background(), fixName)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		require.Nil(t, gotIDP)
	})
}

func TestIDPPresetResolver_IDPPresetsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		var (
			fixIDPPresets    = fixIDPPresets()
			fixIDPPreset     = fixIDPPreset()
			fixGQLIDPPreset  = fixIDPPresetGQL()
			fixGQLIDPPresets = fixIDPPresetsGQL()
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("List", pager.PagingParams{}).Return(fixIDPPresets, nil)

		converterMock := automock.NewGQLIDPPresetConverter()
		defer converterMock.AssertExpectations(t)
		converterMock.On("ToGQL", fixIDPPreset).Return(fixGQLIDPPreset)

		resolver := authentication.NewIDPPresetResolver(svc)
		resolver.SetIDPPresetConverter(converterMock)

		// when
		gotIDPs, err := resolver.IDPPresetsQuery(context.Background(), nil, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, fixGQLIDPPresets, gotIDPs)
	})

	t.Run("Not found", func(t *testing.T) {
		// given
		var (
			fixEmptyIDPPresetsArray    = []*v1alpha1.IDPPreset{}
			fixEmptyGQLIDPPresetsArray = []gqlschema.IDPPreset{}
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("List", pager.PagingParams{}).Return(fixEmptyIDPPresetsArray, nil)

		resolver := authentication.NewIDPPresetResolver(svc)

		// when
		gotIDPs, err := resolver.IDPPresetsQuery(context.Background(), nil, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, fixEmptyGQLIDPPresetsArray, gotIDPs)
	})

	t.Run("Error", func(t *testing.T) {
		// given
		var (
			fixEmptyIDPPresetsArray    = []*v1alpha1.IDPPreset{}
			fixEmptyGQLIDPPresetsArray = []gqlschema.IDPPreset{}
			fixErr                     = errors.New("something went wrong")
		)

		svc := automock.NewIDPPresetSvc()
		defer svc.AssertExpectations(t)
		svc.On("List", pager.PagingParams{}).Return(fixEmptyIDPPresetsArray, fixErr)

		resolver := authentication.NewIDPPresetResolver(svc)

		// when
		gotIDPs, err := resolver.IDPPresetsQuery(context.Background(), nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Equal(t, fixEmptyGQLIDPPresetsArray, gotIDPs)
		assert.Len(t, gotIDPs, 0)
	})
}
