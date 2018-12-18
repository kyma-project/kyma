package application

import (
	helmmocks "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/proto/hapi/release"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"testing"
)

const (
	appName   = "default-app"
	namespace = "integration"
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

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		status, description, err := releaseManager.InstallChart(appName)

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

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		_, _, err := releaseManager.InstallChart(appName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestReleaseManager_DeleteAppChart(t *testing.T) {

	t.Run("should delete release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", appName).Return(nil, nil)

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		err := releaseManager.DeleteChart(appName)

		// then
		assert.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to delete release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", appName).Return(nil, errors.New("Error"))

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		err := releaseManager.DeleteChart(appName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestReleaseManager_CheckReleaseExistence(t *testing.T) {

	t.Run("should return true when release exists", func(t *testing.T) {
		// given
		listReleaseResponse := &rls.ListReleasesResponse{
			Count: 1,
			Releases: []*release.Release{
				{Name: appName},
			},
		}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases").Return(listReleaseResponse, nil)

		releaseManager := NewReleaseManager(helmClient, "", namespace)

		// when
		releaseExists, err := releaseManager.CheckReleaseExistence(appName)

		// then
		assert.NoError(t, err)
		assert.True(t, releaseExists)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return false when release does not exist", func(t *testing.T) {
		// given
		listReleaseResponse := &rls.ListReleasesResponse{}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases").Return(listReleaseResponse, nil)

		releaseManager := NewReleaseManager(helmClient, "", namespace)

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

		releaseManager := NewReleaseManager(helmClient, "", namespace)

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

		releaseManager := NewReleaseManager(helmClient, "", namespace)

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

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		_, _, err := releaseManager.CheckReleaseStatus(appName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}
