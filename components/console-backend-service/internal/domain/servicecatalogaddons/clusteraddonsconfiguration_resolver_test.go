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
)

func TestAddonsConfigurationResolver_AddonsConfigurationsQuery(t *testing.T) {
	// given
	addonsCfgLister := automock.NewAddonsCfgLister()
	defer addonsCfgLister.AssertExpectations(t)

	listedAddons := []*v1alpha1.ClusterAddonsConfiguration{fixClusterAddonsConfiguration("test")}
	addonsCfgLister.On("List", pager.PagingParams{}).
		Return(listedAddons, nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, nil, addonsCfgLister)

	// when
	res, err := resolver.AddonsConfigurationsQuery(context.Background(), nil, nil)

	// then
	require.NoError(t, err)
	assert.Equal(t, []gqlschema.AddonsConfiguration{*fixGQLAddonsConfiguration("test")}, res)
}

func TestAddonsConfigurationResolver_CreateAddonsConfiguration(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgMutation := automock.NewAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)

	addonsCfgMutation.On("Create", addonName, []string{}, &gqlschema.Labels{}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil)

	// when
	res, err := resolver.CreateAddonsConfiguration(context.Background(), addonName, []string{}, &gqlschema.Labels{})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), res)
}

func TestAddonsConfigurationResolver_UpdateAddonsConfiguration(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgMutation := automock.NewAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)
	addonsCfgMutation.On("Update", addonName, []string{}, &gqlschema.Labels{}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil)

	// when
	cfgs, err := resolver.UpdateAddonsConfiguration(context.Background(), addonName, []string{}, &gqlschema.Labels{})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestAddonsConfigurationResolver_DeleteAddonsConfiguration(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgMutation := automock.NewAddonsCfgMutations()
	defer addonsCfgMutation.AssertExpectations(t)
	addonsCfgMutation.On("Delete", addonName).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(nil, addonsCfgMutation, nil)

	// when
	cfgs, err := resolver.DeleteAddonsConfiguration(context.Background(), addonName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestAddonsConfigurationResolver_AddAddonsConfigurationURLs(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("AddRepos", addonName, []string{"app.gg"}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.AddAddonsConfigurationURLs(context.Background(), addonName, []string{"app.gg"})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func TestAddonsConfigurationResolver_RemoveAddonsConfigurationURLs(t *testing.T) {
	// given
	const addonName = "test"
	addonsCfgUpdater := automock.NewAddonsCfgUpdater()
	defer addonsCfgUpdater.AssertExpectations(t)
	addonsCfgUpdater.On("RemoveRepos", addonName, []string{"www.piko.bo"}).
		Return(fixClusterAddonsConfiguration(addonName), nil).Once()

	resolver := servicecatalogaddons.NewAddonsConfigurationResolver(addonsCfgUpdater, nil, nil)

	// when
	cfgs, err := resolver.RemoveAddonsConfigurationURLs(context.Background(), addonName, []string{"www.piko.bo"})

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLAddonsConfiguration(addonName), cfgs)
}

func fixGQLAddonsConfiguration(name string) *gqlschema.AddonsConfiguration {
	return &gqlschema.AddonsConfiguration{
		Name: name,
		Urls: []string{
			"www.piko.bello",
		},
	}
}
