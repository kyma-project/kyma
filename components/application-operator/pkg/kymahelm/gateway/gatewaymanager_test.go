package gateway

import (
	"errors"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/gateway/mocks"
	helmmocks "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"testing"
)

const (
	namespace         = "test"
	expectedOverrides = `global:
	applicationGatewayImage: 
	applicationGatewayTestsImage: `
)

var (
	gatewayName                 = getGatewayName(namespace)
	notEmptyListReleaseResponse = &rls.ListReleasesResponse{
		Count: 1,
		Releases: []*release.Release{
			{Name: gatewayName},
		},
	}

	emptyListReleaseResponse = &rls.ListReleasesResponse{}
)

func TestGatewayManager_InstallGateway(t *testing.T) {
	t.Run("Should install Gateway", func(t *testing.T) {
		//given
		installationResponse := &rls.InstallReleaseResponse{}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", gatewayChartDirectory, namespace, gatewayName, expectedOverrides).Return(installationResponse, nil)

		appMappingClient := &mocks.ApplicationMappingClient{}

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, appMappingClient)

		//when
		err := gatewayManager.InstallGateway(namespace)

		//then
		require.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should fail when Helm fails to install release", func(t *testing.T) {
		//given
		installationResponse := &rls.InstallReleaseResponse{}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", gatewayChartDirectory, namespace, gatewayName, expectedOverrides).
			Return(installationResponse, errors.New("all your base are belong to us"))

		appMappingClient := &mocks.ApplicationMappingClient{}

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, appMappingClient)

		//when
		err := gatewayManager.InstallGateway(namespace)

		//then
		require.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestGatewayManager_DeleteGateway(t *testing.T) {
	t.Run("Should delete gateway", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", gatewayName).Return(nil, nil)
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil)

		// when
		err := gatewayManager.DeleteGateway(namespace)

		// then
		assert.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should succeed if release does not exists", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(emptyListReleaseResponse, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil)

		// when
		err := gatewayManager.DeleteGateway(namespace)

		// then
		assert.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should return error when deleteting gateway fails", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", gatewayName).Return(nil, errors.New("oh no"))
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil)

		// when
		err := gatewayManager.DeleteGateway(namespace)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should return error when checking if gateway exists fails", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(emptyListReleaseResponse, errors.New("uh, me failed"))

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil)

		// when
		err := gatewayManager.DeleteGateway(namespace)

		// then
		assert.Error(t, err)
		helmClient.AssertExpectations(t)
	})
}

func TestGatewayManager_GatewayExists(t *testing.T) {
	t.Run("Should return true when Gateway exists", func(t *testing.T) {
		//given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil)

		//when
		exists, err := gatewayManager.GatewayExists(namespace)

		//then
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Should return false when Gateway does not exist", func(t *testing.T) {
		//given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(emptyListReleaseResponse, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil)

		//when
		exists, err := gatewayManager.GatewayExists(namespace)

		//then
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Should return error when listing Gateways fails", func(t *testing.T) {
		//given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(emptyListReleaseResponse, errors.New("dam, son"))

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil)

		//when
		_, err := gatewayManager.GatewayExists(namespace)

		//then
		require.Error(t, err)
	})
}

func TestGatewayManager_UpgradeGateways(t *testing.T) {
	t.Run("Should update Gateway", func(t *testing.T) {
		//given
		appMappingList := &v1alpha1.ApplicationMappingList{
			Items: []v1alpha1.ApplicationMapping{
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: namespace,
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: namespace,
					},
				}},
		}

		response := &rls.UpdateReleaseResponse{}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil).Once()
		helmClient.On("UpdateReleaseFromChart", gatewayChartDirectory, gatewayName, expectedOverrides).Return(response, nil).Once()

		appMappingClient := &mocks.ApplicationMappingClient{}
		appMappingClient.On("List", v1.ListOptions{}).Return(appMappingList, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, appMappingClient)

		//when
		err := gatewayManager.UpgradeGateways()

		//then
		require.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should update two Gateways", func(t *testing.T) {
		// given
		secondNamespace := "secondnamespace"

		appMappingList := &v1alpha1.ApplicationMappingList{
			Items: []v1alpha1.ApplicationMapping{
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: namespace,
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: secondNamespace,
					},
				}},
		}

		secondNotEmptyListReleaseResponse := &rls.ListReleasesResponse{
			Count: 1,
			Releases: []*release.Release{
				{Name: getGatewayName(secondNamespace)},
			},
		}

		response := &rls.UpdateReleaseResponse{}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil).Once()
		helmClient.On("ListReleases", secondNamespace).Return(secondNotEmptyListReleaseResponse, nil).Once()
		helmClient.On("UpdateReleaseFromChart", gatewayChartDirectory, mock.AnythingOfType("string"), expectedOverrides).Return(response, nil).Twice()

		appMappingClient := &mocks.ApplicationMappingClient{}
		appMappingClient.On("List", v1.ListOptions{}).Return(appMappingList, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, appMappingClient)

		//when
		err := gatewayManager.UpgradeGateways()

		//then
		require.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should return error when listing Application Mappings fails", func(t *testing.T) {
		//given
		helmClient := &helmmocks.HelmClient{}

		appMappingClient := &mocks.ApplicationMappingClient{}
		appMappingClient.On("List", v1.ListOptions{}).Return(nil, errors.New("app mapping list not ok"))

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, appMappingClient)

		//when

		err := gatewayManager.UpgradeGateways()

		//then
		require.Error(t, err)
	})

	t.Run("Should return error when listing releases fails", func(t *testing.T) {
		//given
		appMappingList := &v1alpha1.ApplicationMappingList{
			Items: []v1alpha1.ApplicationMapping{
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: namespace,
					},
				},
			},
		}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(nil, errors.New("woah, error"))

		appMappingClient := &mocks.ApplicationMappingClient{}
		appMappingClient.On("List", v1.ListOptions{}).Return(appMappingList, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, appMappingClient)

		//when

		err := gatewayManager.UpgradeGateways()

		//then
		require.Error(t, err)
	})

	t.Run("Should return error, when updating Gateway fails", func(t *testing.T) {
		//given
		appMappingList := &v1alpha1.ApplicationMappingList{
			Items: []v1alpha1.ApplicationMapping{
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: namespace,
					},
				},
			},
		}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil)

		appMappingClient := &mocks.ApplicationMappingClient{}
		appMappingClient.On("List", v1.ListOptions{}).Return(appMappingList, nil)
		helmClient.On("UpdateReleaseFromChart", gatewayChartDirectory, mock.AnythingOfType("string"), expectedOverrides).Return(nil, errors.New("yikes"))

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, appMappingClient)

		//when

		err := gatewayManager.UpgradeGateways()

		//then
		require.Error(t, err)
	})
}
