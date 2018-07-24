package ui_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/idppreset/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/ui"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/ui/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
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

		resolver := ui.NewIDPPresetResolver(svc)
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

		resolver := ui.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.CreateIDPPresetMutation(context.Background(), fixName, fixIssuer, fixJwksURI)

		// then
		require.Error(t, err)
		require.Nil(t, gotIDP)
		assert.Equal(t, fmt.Sprintf("IDP Preset with the name `%s` already exists", fixName), err.Error())
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

		resolver := ui.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.CreateIDPPresetMutation(context.Background(), fixName, fixIssuer, fixJwksURI)

		// then
		require.Error(t, err)
		require.Nil(t, gotIDP)
		assert.Equal(t, fmt.Sprintf("Cannot create IDP Preset `%s`", fixName), err.Error())
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

		resolver := ui.NewIDPPresetResolver(svc)
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

		resolver := ui.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.DeleteIDPPresetMutation(context.Background(), fixName)

		// then
		require.Error(t, err)
		require.Nil(t, gotIDP)
		assert.Equal(t, fmt.Sprintf("Cannot find IDP Preset `%s`", fixName), err.Error())
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

		resolver := ui.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.DeleteIDPPresetMutation(context.Background(), fixName)

		require.Error(t, err)
		require.Nil(t, gotIDP)
		assert.EqualError(t, err, fmt.Sprintf("Cannot delete IDP Preset `%s`", fixName))
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

		resolver := ui.NewIDPPresetResolver(svc)
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

		resolver := ui.NewIDPPresetResolver(svc)

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

		resolver := ui.NewIDPPresetResolver(svc)

		// when
		gotIDP, err := resolver.IDPPresetQuery(context.Background(), fixName)

		require.Error(t, err)
		require.Nil(t, gotIDP)
		assert.Equal(t, fmt.Sprintf("Cannot query IDP Preset with name `%s`", fixName), err.Error())
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

		resolver := ui.NewIDPPresetResolver(svc)
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

		resolver := ui.NewIDPPresetResolver(svc)

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

		resolver := ui.NewIDPPresetResolver(svc)

		// when
		gotIDPs, err := resolver.IDPPresetsQuery(context.Background(), nil, nil)

		require.Error(t, err)
		assert.Equal(t, fixEmptyGQLIDPPresetsArray, gotIDPs)
		assert.Len(t, gotIDPs, 0)
		assert.Equal(t, fmt.Sprintf("Cannot query IDP Presets"), err.Error())
	})
}
