package servicecatalogaddons_test

import (
	"context"
	"testing"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterAddonsConfigurationResolver_AddonsConfigurationsQuery(t *testing.T) {
	// given
	addonsCfgLister := automock.NewClusterAddonsCfgLister()
	defer addonsCfgLister.AssertExpectations(t)

	listedAddons := []*v1alpha1.ClusterAddonsConfiguration{fixClusterAddonsConfiguration("test")}
	addonsCfgLister.On("List", pager.PagingParams{}).
		Return(listedAddons, nil).Once()

	resolver := servicecatalogaddons.NewClusterAddonsConfigurationResolver(nil, nil, addonsCfgLister)

	// when
	res, err := resolver.ClusterAddonsConfigurationsQuery(context.Background(), nil, nil)

	// then
	require.NoError(t, err)
	assert.Equal(t, []*gqlschema.AddonsConfiguration{fixGQLAddonsConfiguration("test")}, res)
}

func TestClusterAddonsConfigurationResolver_CreateAddonsConfiguration(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgMutation := automock.NewClusterAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)

	addonsCfgMutation.On("Create", addonName, []*gqlschema.AddonsConfigurationRepositoryInput{}, gqlschema.Labels{}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewClusterAddonsConfigurationResolver(nil, addonsCfgMutation, nil)

	// when
	res, err := resolver.CreateClusterAddonsConfiguration(context.Background(), addonName, nil, []string{}, gqlschema.Labels{})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), res)
}

func TestClusterAddonsConfigurationResolver_UpdateAddonsConfiguration(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgMutation := automock.NewClusterAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)
	addonsCfgMutation.On("Update", addonName, []*gqlschema.AddonsConfigurationRepositoryInput{}, gqlschema.Labels{}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewClusterAddonsConfigurationResolver(nil, addonsCfgMutation, nil)

	// when
	cfgs, err := resolver.UpdateClusterAddonsConfiguration(context.Background(), addonName, nil, []string{}, gqlschema.Labels{})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestClusterAddonsConfigurationResolver_DeleteAddonsConfiguration(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgMutation := automock.NewClusterAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)
	addonsCfgMutation.On("Delete", addonName).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewClusterAddonsConfigurationResolver(nil, addonsCfgMutation, nil)

	// when
	cfgs, err := resolver.DeleteClusterAddonsConfiguration(context.Background(), addonName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

// DEPRECATED: Remove
func TestClusterAddonsConfigurationResolver_AddAddonsConfigurationURLs(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewClusterAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("AddRepos", addonName, []*gqlschema.AddonsConfigurationRepositoryInput{{URL: "app.gg"}}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewClusterAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.AddClusterAddonsConfigurationURLs(context.Background(), addonName, []string{"app.gg"})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestClusterAddonsConfigurationResolver_RemoveAddonsConfigurationURLs(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewClusterAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("RemoveRepos", addonName, []string{"www.piko.bo"}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewClusterAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.RemoveClusterAddonsConfigurationURLs(context.Background(), addonName, []string{"www.piko.bo"})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestClusterAddonsConfigurationResolver_AddAddonsConfigurationRepository(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewClusterAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("AddRepos", addonName, []*gqlschema.AddonsConfigurationRepositoryInput{{URL: "app.gg"}}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewClusterAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.AddClusterAddonsConfigurationRepositories(context.Background(), addonName, []*gqlschema.AddonsConfigurationRepositoryInput{{URL: "app.gg"}})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestClusterAddonsConfigurationResolver_RemoveAddonsConfigurationRepository(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewClusterAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("RemoveRepos", addonName, []string{"www.piko.bo"}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewClusterAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.RemoveClusterAddonsConfigurationRepositories(context.Background(), addonName, []string{"www.piko.bo"})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestClusterAddonsConfigurationResolver_ResyncAddonsConfiguration(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewClusterAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("Resync", addonName).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewClusterAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.ResyncClusterAddonsConfiguration(context.Background(), addonName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}
