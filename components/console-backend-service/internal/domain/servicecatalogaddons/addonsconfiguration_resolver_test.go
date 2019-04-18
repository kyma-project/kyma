package servicecatalogaddons_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestAddonsConfigurationResolver_AddonsConfigurationsQuery(t *testing.T) {
	addonsCfgLister := automock.NewAddonsCfgLister()
	defer addonsCfgLister.AssertExpectations(t)
	addonsCfgLister.On("List", pager.PagingParams{nil, nil}).Return([]*v1.ConfigMap{fixAddonsConfigMap("test")}, nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, nil, addonsCfgLister)
	cfgs, err := resolver.AddonsConfigurationsQuery(context.Background(), nil, nil)

	require.NoError(t, err)
	assert.Equal(t, []gqlschema.AddonsConfiguration{*fixAddonsConfiguration("test")}, cfgs)
}

func TestAddonsConfigurationResolver_CreateAddonsConfiguration(t *testing.T) {
	addonsCfgMutation := automock.NewAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)
	expectedCfg := fixAddonsConfiguration("test")
	addonsCfgMutation.On("Create", "test", []string{}, &gqlschema.Labels{}).Return(fixAddonsConfigMap("test"), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil)
	cfgs, err := resolver.CreateAddonsConfiguration(context.Background(), "test", []string{}, &gqlschema.Labels{})

	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfgs)
}

func TestAddonsConfigurationResolver_UpdateAddonsConfiguration(t *testing.T) {
	addonsCfgMutation := automock.NewAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)
	expectedCfg := fixAddonsConfiguration("test")
	addonsCfgMutation.On("Update", "test", []string{}, &gqlschema.Labels{}).Return(fixAddonsConfigMap("test"), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil)
	cfgs, err := resolver.UpdateAddonsConfiguration(context.Background(), "test", []string{}, &gqlschema.Labels{})

	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfgs)
}

func TestAddonsConfigurationResolver_DeleteAddonsConfiguration(t *testing.T) {
	addonsCfgMutation := automock.NewAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)
	expectedCfg := fixAddonsConfiguration("test")
	addonsCfgMutation.On("Delete", "test").Return(fixAddonsConfigMap("test"), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil)
	cfgs, err := resolver.DeleteAddonsConfiguration(context.Background(), "test")

	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfgs)
}

func TestAddonsConfigurationResolver_AddAddonsConfigurationURLs(t *testing.T) {
	addonsCfgUpdater := automock.NewAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	expectedCfg := fixAddonsConfiguration("test")
	addonsCfgUpdater.On("AddRepos", "test", []string{"app.gg"}).Return(fixAddonsConfigMap("test"), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)
	cfgs, err := resolver.AddAddonsConfigurationURLs(context.Background(), "test", []string{"app.gg"})

	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfgs)
}

func TestAddonsConfigurationResolver_RemoveAddonsConfigurationURLs(t *testing.T) {
	addonsCfgUpdater := automock.NewAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	expectedCfg := fixAddonsConfiguration("test")
	addonsCfgUpdater.On("RemoveRepos", "test", []string{"www.piko.bo"}).Return(fixAddonsConfigMap("test"), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)
	cfgs, err := resolver.RemoveAddonsConfigurationURLs(context.Background(), "test", []string{"www.piko.bo"})

	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfgs)
}

func fixAddonsConfiguration(name string) *gqlschema.AddonsConfiguration {
	return &gqlschema.AddonsConfiguration{
		Name: name,
		Labels: gqlschema.Labels{
			addonsCfgLabelKey: addonsCfgLabelValue,
		},
		Urls: []string{
			"www.piko.bo",
		},
	}
}
