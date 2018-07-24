package ui_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/idppreset/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/idppreset/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/idppreset/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/ui"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testingErrors "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

func TestIDPPreset(t *testing.T) {
	t.Run("Create - with success", func(t *testing.T) {
		// given
		var (
			fixName    = "fixIDPPreset"
			fixIssuer  = "issuer"
			fixJwksURI = "uri"
		)

		fakeClient := fake.NewSimpleClientset()
		svc := ui.NewIDPPresetService(fakeClient.UiV1alpha1(), nil)

		// when
		idp, err := svc.Create(fixName, fixIssuer, fixJwksURI)

		// then
		require.NoError(t, err)
		require.NotNil(t, idp)
		assert.Equal(t, idp.Name, fixName)
		assert.Equal(t, idp.Kind, "IDPPreset")
		assert.Equal(t, idp.Spec.Name, fixName)
		assert.Equal(t, idp.Spec.Issuer, fixIssuer)
		assert.Equal(t, idp.Spec.JwksUri, fixJwksURI)
	})

	t.Run("Delete - with success", func(t *testing.T) {
		// given
		var (
			fixName         = "fixIDPPreset"
			fixIDPPresetObj = fixIDPPreset()
		)

		fakeClient := fake.NewSimpleClientset(fixIDPPresetObj)
		svc := ui.NewIDPPresetService(fakeClient.UiV1alpha1(), nil)

		// when
		errFromDelete := svc.Delete(fixName)
		_, err := fakeClient.UiV1alpha1().IDPPresets().Get(fixName, v1.GetOptions{})

		// then
		require.NoError(t, errFromDelete)
		assert.True(t, apiErrors.IsNotFound(err))
	})

	t.Run("Delete - with no success", func(t *testing.T) {
		// given
		var (
			errorMsg        = fmt.Errorf("Some error")
			fixName         = "fixIDPPreset"
			fixIDPPresetObj = fixIDPPreset()
		)

		fakeClient := fake.NewSimpleClientset(fixIDPPresetObj)
		failingReaction := func(action testingErrors.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, errorMsg
		}
		fakeClient.PrependReactor("delete", "idppresets", failingReaction)
		svc := ui.NewIDPPresetService(fakeClient.UiV1alpha1(), nil)

		// when
		err := svc.Delete(fixName)

		// then
		assert.Equal(t, errorMsg, err)
	})

	t.Run("Find - with success", func(t *testing.T) {
		// given
		var (
			fixName         = "fixIDPPreset"
			fixIssuer       = "issuer"
			fixJwksUri      = "uri"
			fixIDPPresetObj = fixIDPPreset()
		)

		fakeClient := fake.NewSimpleClientset(fixIDPPresetObj)
		informer := fixIDPPresetInformer(fakeClient)
		svc := ui.NewIDPPresetService(nil, informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		// when
		idp, err := svc.Find(fixName)

		// then
		require.NoError(t, err)
		require.NotNil(t, idp)
		assert.Equal(t, idp.Name, fixName)
		assert.Equal(t, idp.Spec.JwksUri, fixJwksUri)
		assert.Equal(t, idp.Spec.Issuer, fixIssuer)
	})

	t.Run("Find - with no success - returns nil", func(t *testing.T) {
		// given
		fixName := "fixIDPPreset"

		fakeClient := fake.NewSimpleClientset()
		informer := fixIDPPresetInformer(fakeClient)
		svc := ui.NewIDPPresetService(nil, informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		// when
		idp, err := svc.Find(fixName)

		// then
		require.NoError(t, err)
		assert.Nil(t, idp)
	})

	t.Run("List", func(t *testing.T) {
		// given
		fixIDPPresetObj := fixIDPPreset()

		fakeClient := fake.NewSimpleClientset(fixIDPPresetObj)
		informer := fixIDPPresetInformer(fakeClient)
		svc := ui.NewIDPPresetService(nil, informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		// when
		idps, err := svc.List(pager.PagingParams{})

		// then
		require.NoError(t, err)
		assert.Equal(t, []*v1alpha1.IDPPreset{
			fixIDPPresetObj,
		}, idps)
	})

	t.Run("List - with no success - returns empty array", func(t *testing.T) {
		// given
		emptyIDPPresetsArray := []*v1alpha1.IDPPreset{}

		fakeClient := fake.NewSimpleClientset()
		informer := fixIDPPresetInformer(fakeClient)
		svc := ui.NewIDPPresetService(nil, informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		// when
		idps, err := svc.List(pager.PagingParams{})

		// then
		require.NoError(t, err)
		assert.Equal(t, emptyIDPPresetsArray, idps)
	})
}

func fixIDPPresetInformer(fakeClient *fake.Clientset) cache.SharedIndexInformer {
	informerFactory := externalversions.NewSharedInformerFactory(fakeClient, 0)
	informer := informerFactory.Ui().V1alpha1().IDPPresets().Informer()

	return informer
}
