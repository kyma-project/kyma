package servicecatalogaddons_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestAddonsConfigurationResolver_AddonsConfigurationsQuery(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
		// given
		addonsCfgLister := automock.NewAddonsCfgLister()
		defer addonsCfgLister.AssertExpectations(t)

		listedAddons := []*v1alpha1.ClusterAddonsConfiguration{fixClusterAddonsConfiguration("test")}
		addonsCfgLister.On("List", pager.PagingParams{}).
			Return(listedAddons, nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, nil, addonsCfgLister, true)

		// when
		res, err := resolver.AddonsConfigurationsQuery(context.Background(), nil, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, []gqlschema.AddonsConfiguration{*fixGQLAddonsConfiguration("test")}, res)
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		// given
		addonsCfgLister := automock.NewAddonsCfgLister()
		defer addonsCfgLister.AssertExpectations(t)
		addonsCfgLister.On("ListConfigMaps", pager.PagingParams{nil, nil}).Return([]*v1.ConfigMap{fixAddonsConfigMap("test")}, nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, nil, addonsCfgLister, false)

		// when
		cfgs, err := resolver.AddonsConfigurationsQuery(context.Background(), nil, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, []gqlschema.AddonsConfiguration{*fixAddonsConfigurationFromCM("test")}, cfgs)
	})
}

func TestAddonsConfigurationResolver_CreateAddonsConfiguration(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
		// given
		const addonName = "test"
		addonsCfgMutation := automock.NewAddonsCfgMutations()
		defer addonsCfgMutation.AssertExpectations(t)

		addonsCfgMutation.On("Create", addonName, []string{}, &gqlschema.Labels{}).
			Return(fixClusterAddonsConfiguration(addonName), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil, true)

		// when
		res, err := resolver.CreateAddonsConfiguration(context.Background(), addonName, []string{}, &gqlschema.Labels{})

		// then
		require.NoError(t, err)
		assert.Equal(t, fixGQLAddonsConfiguration(addonName), res)
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		// given
		addonsCfgMutation := automock.NewAddonsCfgMutations()
		defer addonsCfgMutation.AssertExpectations(t)
		expectedCfg := fixAddonsConfigurationFromCM("test")
		addonsCfgMutation.On("CreateConfigMap", "test", []string{}, &gqlschema.Labels{}).Return(fixAddonsConfigMap("test"), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil, false)

		// when
		cfgs, err := resolver.CreateAddonsConfiguration(context.Background(), "test", []string{}, &gqlschema.Labels{})

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedCfg, cfgs)
	})
}

func TestAddonsConfigurationResolver_UpdateAddonsConfiguration(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
		// given
		const addonName = "test"
		addonsCfgMutation := automock.NewAddonsCfgMutations()
		defer addonsCfgMutation.AssertExpectations(t)
		addonsCfgMutation.On("Update", addonName, []string{}, &gqlschema.Labels{}).
			Return(fixClusterAddonsConfiguration(addonName), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil, true)

		// when
		cfgs, err := resolver.UpdateAddonsConfiguration(context.Background(), addonName, []string{}, &gqlschema.Labels{})

		// then
		require.NoError(t, err)
		assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		addonsCfgMutation := automock.NewAddonsCfgMutations()
		defer addonsCfgMutation.AssertExpectations(t)
		expectedCfg := fixAddonsConfigurationFromCM("test")
		addonsCfgMutation.On("UpdateConfigMap", "test", []string{}, &gqlschema.Labels{}).Return(fixAddonsConfigMap("test"), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil, false)
		cfgs, err := resolver.UpdateAddonsConfiguration(context.Background(), "test", []string{}, &gqlschema.Labels{})

		require.NoError(t, err)
		assert.Equal(t, expectedCfg, cfgs)
	})
}

func TestAddonsConfigurationResolver_DeleteAddonsConfiguration(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
		// given
		const addonName = "test"
		addonsCfgMutation := automock.NewAddonsCfgMutations()
		defer addonsCfgMutation.AssertExpectations(t)
		addonsCfgMutation.On("Delete", addonName).
			Return(fixClusterAddonsConfiguration(addonName), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil, true)

		// when
		cfgs, err := resolver.DeleteAddonsConfiguration(context.Background(), addonName)

		// then
		require.NoError(t, err)
		assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		addonsCfgMutation := automock.NewAddonsCfgMutations()
		defer addonsCfgMutation.AssertExpectations(t)
		expectedCfg := fixAddonsConfigurationFromCM("test")
		addonsCfgMutation.On("DeleteConfigMap", "test").Return(fixAddonsConfigMap("test"), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil, false)
		cfgs, err := resolver.DeleteAddonsConfiguration(context.Background(), "test")

		require.NoError(t, err)
		assert.Equal(t, expectedCfg, cfgs)
	})
}

func TestAddonsConfigurationResolver_AddAddonsConfigurationURLs(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
		// given
		const addonName = "test"
		addonsCfgUpdater := automock.NewAddonsCfgUpdater()
		defer addonsCfgUpdater.AssertExpectations(t)
		addonsCfgUpdater.On("AddRepos", addonName, []string{"app.gg"}).
			Return(fixClusterAddonsConfiguration(addonName), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil, true)

		// when
		cfgs, err := resolver.AddAddonsConfigurationURLs(context.Background(), addonName, []string{"app.gg"})

		// then
		require.NoError(t, err)
		assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		addonsCfgUpdater := automock.NewAddonsCfgUpdater()
		defer addonsCfgUpdater.AssertExpectations(t)
		expectedCfg := fixAddonsConfigurationFromCM("test")
		addonsCfgUpdater.On("AddReposToConfigMap", "test", []string{"app.gg"}).Return(fixAddonsConfigMap("test"), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil, false)
		cfgs, err := resolver.AddAddonsConfigurationURLs(context.Background(), "test", []string{"app.gg"})

		require.NoError(t, err)
		assert.Equal(t, expectedCfg, cfgs)
	})
}

func TestAddonsConfigurationResolver_RemoveAddonsConfigurationURLs(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
		// given
		const addonName = "test"
		addonsCfgUpdater := automock.NewAddonsCfgUpdater()
		defer addonsCfgUpdater.AssertExpectations(t)
		addonsCfgUpdater.On("RemoveRepos", addonName, []string{"www.piko.bo"}).
			Return(fixClusterAddonsConfiguration(addonName), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil, true)

		// when
		cfgs, err := resolver.RemoveAddonsConfigurationURLs(context.Background(), addonName, []string{"www.piko.bo"})

		// then
		require.NoError(t, err)
		assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)

	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		addonsCfgUpdater := automock.NewAddonsCfgUpdater()
		defer addonsCfgUpdater.AssertExpectations(t)
		expectedCfg := fixAddonsConfigurationFromCM("test")
		addonsCfgUpdater.On("RemoveReposFromConfigMap", "test", []string{"www.piko.bo"}).Return(fixAddonsConfigMap("test"), nil).Once()

		resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil, false)
		cfgs, err := resolver.RemoveAddonsConfigurationURLs(context.Background(), "test", []string{"www.piko.bo"})

		require.NoError(t, err)
		assert.Equal(t, expectedCfg, cfgs)
	})
}

func fixAddonsConfigurationFromCM(name string) *gqlschema.AddonsConfiguration {
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

func fixGQLAddonsConfiguration(name string) *gqlschema.AddonsConfiguration {
	return &gqlschema.AddonsConfiguration{
		Name: name,
		Urls: []string{
			"www.piko.bello",
		},
	}
}
