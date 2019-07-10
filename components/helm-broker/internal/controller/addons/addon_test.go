package addons

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestAddonController_IsReady(t *testing.T) {
	// Given
	ta := testAddon()

	// Then
	assert.Equal(t, "default-addon", ta.Addon.Name)
	assert.Equal(t, "1.0", ta.Addon.Version)
	assert.True(t, ta.IsReady())
}

func TestAddonController_IsComplete(t *testing.T) {
	// Given
	ta := testAddon()

	// When
	ta.ID = "7929c146-bf8d-4b65-8eba-8348ac956546"

	// Then
	assert.True(t, ta.IsComplete())
}

func TestAddonController_FetchingError(t *testing.T) {
	// Given
	ta := testAddon()

	// When
	ta.FetchingError(fmt.Errorf("some error:a:b:c:d:e:f:g"))

	// Then
	assert.False(t, ta.IsReady())
	assert.Equal(t, v1alpha1.AddonStatusFailed, ta.Addon.Status)
	assert.Equal(t, v1alpha1.AddonFetchingError, ta.Addon.Reason)
	assert.Equal(t, "Fetching failed due to error: 'some error:a:b:c'", ta.Addon.Message)
}

func TestAddonController_LoadingError(t *testing.T) {
	// Given
	ta := testAddon()

	// When
	ta.LoadingError(fmt.Errorf("loading error"))

	// Then
	assert.False(t, ta.IsReady())
	assert.Equal(t, v1alpha1.AddonStatusFailed, ta.Addon.Status)
	assert.Equal(t, v1alpha1.AddonLoadingError, ta.Addon.Reason)
	assert.Equal(t, "Loading failed due to error: 'loading error'", ta.Addon.Message)
}

func TestAddonController_ConflictInSpecifiedRepositories(t *testing.T) {
	// Given
	ta := testAddon()

	// When
	ta.ConflictInSpecifiedRepositories(fmt.Errorf("id exist in repositories"))

	// Then
	assert.False(t, ta.IsReady())
	assert.Equal(t, v1alpha1.AddonStatusFailed, ta.Addon.Status)
	assert.Equal(t, v1alpha1.AddonConflictInSpecifiedRepositories, ta.Addon.Reason)
	assert.Equal(t, "Specified repositories have addons with the same ID: id exist in repositories", ta.Addon.Message)
}

func TestAddonController_ConflictWithAlreadyRegisteredAddons(t *testing.T) {
	// Given
	ta := testAddon()

	// When
	ta.ConflictWithAlreadyRegisteredAddons(fmt.Errorf("id exist in storage"))

	// Then
	assert.False(t, ta.IsReady())
	assert.Equal(t, v1alpha1.AddonStatusFailed, ta.Addon.Status)
	assert.Equal(t, v1alpha1.AddonConflictWithAlreadyRegisteredAddons, ta.Addon.Reason)
	assert.Equal(t, "An addon with the same ID is already registered: id exist in storage", ta.Addon.Message)
}

func TestAddonController_RegisteringError(t *testing.T) {
	// Given
	ta := testAddon()

	// When
	ta.RegisteringError(fmt.Errorf("cannot register"))

	// Then
	assert.False(t, ta.IsReady())
	assert.Equal(t, v1alpha1.AddonStatusFailed, ta.Addon.Status)
	assert.Equal(t, v1alpha1.AddonRegisteringError, ta.Addon.Reason)
	assert.Equal(t, "Registering failed due to error: 'cannot register'", ta.Addon.Message)
}

func testAddon() *AddonController {
	return NewAddon("default-addon", "1.0", "https://example.com")
}
