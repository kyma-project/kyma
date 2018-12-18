package remoteenvironemnts

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
	reName    = "default-re"
	namespace = "integration"
)

func TestReleaseManager_InstallNewREChart(t *testing.T) {

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
		helmClient.On("InstallReleaseFromChart", reChartDirectory, namespace, reName, "overrides").Return(installationResponse, nil)

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		status, description, err := releaseManager.InstallNewREChart(reName)

		// then
		assert.NoError(t, err)
		assert.Equal(t, hapi_release5.Status_DEPLOYED, status)
		assert.Equal(t, "Installed", description)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to install release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", reChartDirectory, namespace, reName, "overrides").Return(nil, errors.New("Error"))

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		_, _, err := releaseManager.InstallNewREChart(reName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestReleaseManager_DeleteREChart(t *testing.T) {

	t.Run("should delete release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", reName).Return(nil, nil)

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		err := releaseManager.DeleteREChart(reName)

		// then
		assert.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to delete release", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", reName).Return(nil, errors.New("Error"))

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		err := releaseManager.DeleteREChart(reName)

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
				{Name: reName},
			},
		}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases").Return(listReleaseResponse, nil)

		releaseManager := NewReleaseManager(helmClient, "", namespace)

		// when
		releaseExists, err := releaseManager.CheckReleaseExistence(reName)

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
		releaseExists, err := releaseManager.CheckReleaseExistence(reName)

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
		_, err := releaseManager.CheckReleaseExistence(reName)

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
		helmClient.On("ReleaseStatus", reName).Return(getResponse, nil)

		releaseManager := NewReleaseManager(helmClient, "", namespace)

		// when
		status, description, err := releaseManager.CheckReleaseStatus(reName)

		// then
		assert.NoError(t, err)
		assert.Equal(t, hapi_release5.Status_DEPLOYED, status)
		assert.Equal(t, "Installed", description)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to get release status", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ReleaseStatus", reName).Return(nil, errors.New("Error"))

		releaseManager := NewReleaseManager(helmClient, "overrides", namespace)

		// when
		_, _, err := releaseManager.CheckReleaseStatus(reName)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}
