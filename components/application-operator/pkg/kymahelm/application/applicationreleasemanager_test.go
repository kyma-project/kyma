package application

import (
	"context"
	"testing"

	hapi_release5 "helm.sh/helm/v3/pkg/release"

	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/application/mocks"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	helmmocks "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	appName      = "default-app"
	namespace    = "integration"
	group        = "group"
	tenant       = "tenant"
	emptyProfile = ""
)

var (
	notEmptyListReleaseResponse = []*hapi_release5.Release{
		{Name: appName},
	}
	emptyListReleaseResponse []*hapi_release5.Release

	emptyOverrides              = map[string]interface{}{"global": map[string]interface{}{}}
	overridesWithTenantAndGroup = map[string]interface{}{
		"global": map[string]interface{}{
			"tenant": tenant,
			"group":  group,
		},
	}
)

func TestReleaseManager_InstallNewAppChart(t *testing.T) {

	application := &v1alpha1.Application{
		ObjectMeta: v1.ObjectMeta{Name: appName},
	}

	t.Run("should install release with CN equal to app name", func(t *testing.T) {

		// given
		installationResponse := &hapi_release5.Release{
			Info: &hapi_release5.Info{
				Status:      hapi_release5.StatusDeployed,
				Description: "Installed",
			},
		}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", applicationChartDirectory, appName, namespace, emptyOverrides, emptyProfile).Return(installationResponse, nil)

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		status, description, err := releaseManager.InstallChart(application)

		// then
		assert.NoError(t, err)
		assert.Equal(t, hapi_release5.StatusDeployed, status)
		assert.Equal(t, "Installed", description)
		helmClient.AssertExpectations(t)
	})

	t.Run("should install release with CN equal to app name, O equal to tenant and OU equal to group", func(t *testing.T) {
		// given
		installationResponse := &hapi_release5.Release{
			Info: &hapi_release5.Info{
				Status:      hapi_release5.StatusDeployed,
				Description: "Installed",
			},
		}

		appWithGroupAndTenant := &v1alpha1.Application{
			ObjectMeta: v1.ObjectMeta{Name: appName},
			Spec:       v1alpha1.ApplicationSpec{Tenant: tenant, Group: group},
		}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", applicationChartDirectory, appName, namespace, overridesWithTenantAndGroup, emptyProfile).Return(installationResponse, nil)

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		status, description, err := releaseManager.InstallChart(appWithGroupAndTenant)

		// then
		assert.NoError(t, err)
		assert.Equal(t, hapi_release5.StatusDeployed, status)
		assert.Equal(t, "Installed", description)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to install release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", applicationChartDirectory, appName, namespace, emptyOverrides, emptyProfile).Return(nil, errors.New("Error"))

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		_, _, err := releaseManager.InstallChart(application)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestReleaseManager_DeleteReleaseIfExists(t *testing.T) {

	t.Run("should delete release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", appName, namespace).Return(nil, nil)
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil)

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		err := releaseManager.DeleteReleaseIfExists(appName)

		// then
		assert.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("should succeed if release does not exists", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(emptyListReleaseResponse, nil)

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		err := releaseManager.DeleteReleaseIfExists(appName)

		// then
		assert.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to delete release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", appName, namespace).Return(nil, errors.New("Error"))
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil)

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		err := releaseManager.DeleteReleaseIfExists(appName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to check existence", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(nil, errors.New("error"))

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		err := releaseManager.DeleteReleaseIfExists(appName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestReleaseManager_CheckReleaseExistence(t *testing.T) {

	t.Run("should return true when release exists", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil)

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		releaseExists, err := releaseManager.CheckReleaseExistence(appName)

		// then
		assert.NoError(t, err)
		assert.True(t, releaseExists)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return false when release does not exist", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(emptyListReleaseResponse, nil)

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		releaseExists, err := releaseManager.CheckReleaseExistence(appName)

		// then
		assert.NoError(t, err)
		assert.False(t, releaseExists)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error if failed to list releases", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(nil, errors.New("Error"))

		releaseManager := NewApplicationReleaseManager(helmClient, nil, OverridesData{}, namespace, emptyProfile)

		// when
		_, err := releaseManager.CheckReleaseExistence(appName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestReleaseManager_UpgradeReleases(t *testing.T) {

	t.Run("should upgrade existing releases", func(t *testing.T) {
		// given
		application := v1alpha1.Application{
			ObjectMeta: v1.ObjectMeta{Name: "app-1"},
		}

		otherApplication := v1alpha1.Application{
			ObjectMeta: v1.ObjectMeta{Name: "app-2"},
		}

		applicationList := &v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				application,
				otherApplication,
			},
		}

		updateResponse := &hapi_release5.Release{
			Info: &hapi_release5.Info{
				Status:      hapi_release5.StatusDeployed,
				Description: "Installed",
			},
		}

		helmListReleaseResponse := []*hapi_release5.Release{
			{Name: "app-1"},
			{Name: "app-2"},
		}

		appClient := &mocks.ApplicationClient{}
		appClient.On("List", context.Background(), mock.AnythingOfType("ListOptions")).Return(applicationList, nil)

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("UpdateReleaseFromChart", applicationChartDirectory, "app-1", namespace, emptyOverrides, emptyProfile).Return(updateResponse, nil)
		helmClient.On("UpdateReleaseFromChart", applicationChartDirectory, "app-2", namespace, emptyOverrides, emptyProfile).Return(updateResponse, nil)
		helmClient.On("ListReleases", namespace).Return(helmListReleaseResponse, nil)

		releaseManager := NewApplicationReleaseManager(helmClient, appClient, OverridesData{}, namespace, emptyProfile)

		// when
		err := releaseManager.UpgradeApplicationReleases()

		// then
		assert.NoError(t, err)
		appClient.AssertExpectations(t)
		helmClient.AssertExpectations(t)
	})

	t.Run("should set a proper status if upgrade failed", func(t *testing.T) {
		// given

		updateResponse := &hapi_release5.Release{
			Info: &hapi_release5.Info{
				Status:      hapi_release5.StatusFailed,
				Description: "Failed",
			},
		}

		applicationList := &v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				{
					ObjectMeta: v1.ObjectMeta{Name: "app-1"},
					Status: v1alpha1.ApplicationStatus{
						InstallationStatus: v1alpha1.InstallationStatus{},
					},
				},
			},
		}

		updatedApplication := &v1alpha1.Application{
			ObjectMeta: v1.ObjectMeta{Name: "app-1"},
			Status: v1alpha1.ApplicationStatus{
				InstallationStatus: v1alpha1.InstallationStatus{
					Status:      "failed",
					Description: emptyProfile,
				},
			},
		}

		helmListReleaseResponse := []*hapi_release5.Release{
			{Name: "app-1"},
		}

		appClient := &mocks.ApplicationClient{}
		appClient.On("List", context.Background(), mock.AnythingOfType("ListOptions")).Return(applicationList, nil)
		appClient.On("Update", context.Background(), updatedApplication, v1.UpdateOptions{}).Return(&applicationList.Items[0], nil)

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("UpdateReleaseFromChart", applicationChartDirectory, "app-1", namespace, emptyOverrides, emptyProfile).Return(updateResponse, errors.New("Error"))
		helmClient.On("ListReleases", namespace).Return(helmListReleaseResponse, nil)

		releaseManager := NewApplicationReleaseManager(helmClient, appClient, OverridesData{}, namespace, emptyProfile)

		// when
		err := releaseManager.UpgradeApplicationReleases()

		// then
		assert.NoError(t, err)
		appClient.AssertExpectations(t)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to fetch application list ", func(t *testing.T) {
		// given
		appClient := &mocks.ApplicationClient{}
		appClient.On("List", context.Background(), mock.AnythingOfType("ListOptions")).Return(nil, errors.New("Error"))

		releaseManager := NewApplicationReleaseManager(nil, appClient, OverridesData{}, namespace, emptyProfile)

		// when
		err := releaseManager.UpgradeApplicationReleases()

		// then
		assert.Error(t, err)
		appClient.AssertExpectations(t)
	})
}
