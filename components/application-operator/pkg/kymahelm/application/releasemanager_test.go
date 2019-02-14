package application

import (
	"testing"

	helmmocks "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/proto/hapi/release"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

const (
	appName   = "default-app"
	namespace = "integration"
)

var (
	notEmptyListReleaseResponse = &rls.ListReleasesResponse{
		Count: 1,
		Releases: []*release.Release{
			{Name: appName},
		},
	}

	emptyListReleaseResponse = &rls.ListReleasesResponse{}
)

func TestReleaseManager_InstallNewAppChart(t *testing.T) {

	t.Run("should return status when release installed", func(t *testing.T) {
		// given
		installationResponse := &rls.InstallReleaseResponse{
			Release: &hapi_release5.Release{
				Info: &hapi_release5.Info{
					Status: &hapi_release5.Status{
						Code: hapi_release5.Status_DEPLOYED,
					},
					Description: "Installed",
				},
			},
		}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", applicationChartDirectory, namespace, appName, "overrides").Return(installationResponse, nil)

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

		// when
		status, description, err := releaseManager.InstallChart(appName, "overrides")

		// then
		assert.NoError(t, err)
		assert.Equal(t, hapi_release5.Status_DEPLOYED, status)
		assert.Equal(t, "Installed", description)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to install release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", applicationChartDirectory, namespace, appName, "overrides").Return(nil, errors.New("Error"))

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

		// when
		_, _, err := releaseManager.InstallChart(appName, "overrides")

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestReleaseManager_DeleteReleaseIfExists(t *testing.T) {

	t.Run("should delete release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", appName).Return(nil, nil)
		helmClient.On("ListReleases").Return(notEmptyListReleaseResponse, nil)

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

		// when
		err := releaseManager.DeleteReleaseIfExists(appName)

		// then
		assert.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("should succeed if release does not exists", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases").Return(emptyListReleaseResponse, nil)

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

		// when
		err := releaseManager.DeleteReleaseIfExists(appName)

		// then
		assert.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to delete release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", appName).Return(nil, errors.New("Error"))
		helmClient.On("ListReleases").Return(notEmptyListReleaseResponse, nil)

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

		// when
		err := releaseManager.DeleteReleaseIfExists(appName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to check existence", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases").Return(nil, errors.New("error"))

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

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
		helmClient.On("ListReleases").Return(notEmptyListReleaseResponse, nil)

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

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
		helmClient.On("ListReleases").Return(emptyListReleaseResponse, nil)

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

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
		helmClient.On("ListReleases").Return(nil, errors.New("Error"))

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

		// when
		_, err := releaseManager.CheckReleaseExistence(appName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestReleaseManager_CheckReleaseStatus(t *testing.T) {

	t.Run("should return status and desription", func(t *testing.T) {
		// given
		getResponse := &rls.GetReleaseStatusResponse{
			Info: &hapi_release5.Info{
				Status: &hapi_release5.Status{
					Code: hapi_release5.Status_DEPLOYED,
				},
				Description: "Installed",
			},
		}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ReleaseStatus", appName).Return(getResponse, nil)

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

		// when
		status, description, err := releaseManager.CheckReleaseStatus(appName)

		// then
		assert.NoError(t, err)
		assert.Equal(t, hapi_release5.Status_DEPLOYED, status)
		assert.Equal(t, "Installed", description)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to get release status", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ReleaseStatus", appName).Return(nil, errors.New("Error"))

		releaseManager := NewReleaseManager(helmClient, OverridesData{}, namespace)

		// when
		_, _, err := releaseManager.CheckReleaseStatus(appName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}
