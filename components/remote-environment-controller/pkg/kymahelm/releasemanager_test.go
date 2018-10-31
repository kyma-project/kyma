package kymahelm

import (
	"testing"
	"k8s.io/helm/pkg/proto/hapi/release"
	"github.com/stretchr/testify/assert"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	helmmocks "github.com/kyma-project/kyma/components/remote-environment-controller/pkg/kymahelm/mocks"
)

const (
	reName = "default-re"
	namespace = "integration"
)

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

	})

	t.Run("should return error if failed to list releases", func(t *testing.T) {

	})
}