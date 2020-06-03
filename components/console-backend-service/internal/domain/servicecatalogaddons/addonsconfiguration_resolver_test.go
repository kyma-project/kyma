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

func TestAddonsConfigurationResolver_AddonsConfigurationsQuery(t *testing.T) {
	// given
	addonsCfgLister := automock.NewAddonsCfgLister()
	defer addonsCfgLister.AssertExpectations(t)

	fixNS := "test"
	listedAddons := []*v1alpha1.AddonsConfiguration{fixAddonsConfiguration(fixNS)}
	addonsCfgLister.On("List", fixNS, pager.PagingParams{}).
		Return(listedAddons, nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, nil, addonsCfgLister)

	// when
	res, err := resolver.AddonsConfigurationsQuery(context.Background(), fixNS, nil, nil)

	// then
	require.NoError(t, err)
	assert.Equal(t, []*gqlschema.AddonsConfiguration{fixGQLAddonsConfiguration(fixNS)}, res)
}

func TestAddonsConfigurationResolver_CreateAddonsConfiguration(t *testing.T) {
	// given
	fixNS := "test"
	addonsCfgMutation := automock.NewAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)

	addonsCfgMutation.On("Create", fixNS, fixNS, []*gqlschema.AddonsConfigurationRepositoryInput{}, gqlschema.Labels{}).
		Return(fixAddonsConfiguration(fixNS), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil)

	// when
	res, err := resolver.CreateAddonsConfiguration(context.Background(), fixNS, fixNS, nil, []string{}, gqlschema.Labels{})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(fixNS), res)
}

func TestAddonsConfigurationResolver_UpdateAddonsConfiguration(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgMutation := automock.NewAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)
	addonsCfgMutation.On("Update", addonName, addonName, []*gqlschema.AddonsConfigurationRepositoryInput{}, gqlschema.Labels{}).
		Return(fixAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil)

	// when
	cfgs, err := resolver.UpdateAddonsConfiguration(context.Background(), addonName, addonName, nil, []string{}, gqlschema.Labels{})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestAddonsConfigurationResolver_DeleteAddonsConfiguration(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgMutation := automock.NewAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)
	addonsCfgMutation.On("Delete", addonName, addonName).
		Return(fixAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil)

	// when
	cfgs, err := resolver.DeleteAddonsConfiguration(context.Background(), addonName, addonName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestAddonsConfigurationResolver_AddAddonsConfigurationURLs(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("AddRepos", addonName, addonName, []*gqlschema.AddonsConfigurationRepositoryInput{{URL: "app.gg"}}).
		Return(fixAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.AddAddonsConfigurationURLs(context.Background(), addonName, addonName, []string{"app.gg"})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestAddonsConfigurationResolver_RemoveAddonsConfigurationURLs(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("RemoveRepos", addonName, addonName, []string{"www.piko.bo"}).
		Return(fixAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.RemoveAddonsConfigurationURLs(context.Background(), addonName, addonName, []string{"www.piko.bo"})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestAddonsConfigurationResolver_AddAddonsConfigurationRepository(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	url := "app.gg"
	addonsCfgUpdater.On("AddRepos", addonName, addonName, []*gqlschema.AddonsConfigurationRepositoryInput{{URL: url}}).
		Return(fixAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.AddAddonsConfigurationRepositories(context.Background(), addonName, addonName, []*gqlschema.AddonsConfigurationRepositoryInput{{URL: url}})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestAddonsConfigurationResolver_RemoveAddonsConfigurationRepository(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	url := "www.piko.bo"
	addonsCfgUpdater.On("RemoveRepos", addonName, addonName, []string{url}).
		Return(fixAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.RemoveAddonsConfigurationRepositories(context.Background(), addonName, addonName, []string{url})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestAddonsConfigurationResolver_ResyncAddonsConfigurationURLs(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("Resync", addonName, addonName).
		Return(fixAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.ResyncAddonsConfiguration(context.Background(), addonName, addonName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func fixGQLAddonsConfiguration(name string) *gqlschema.AddonsConfiguration {
	url := "www.piko.bello"
	return &gqlschema.AddonsConfiguration{
		Name: name,
		Urls: []string{
			url,
		},
		Repositories: []*gqlschema.AddonsConfigurationRepository{
			{
				URL: url,
			},
		},
		Labels: gqlschema.Labels{},
		Status: &gqlschema.AddonsConfigurationStatus{
			Repositories: []*gqlschema.AddonsConfigurationStatusRepository{},
		},
	}
}
