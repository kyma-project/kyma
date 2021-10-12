package gateway

import (
	"context"
	"errors"
	"testing"

	"helm.sh/helm/v3/pkg/release"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/gateway/mocks"
	helmmocks "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace    = "test"
	emptyProfile = ""
)

var (
	gatewayName = getGatewayReleaseName(namespace)

	notEmptyListReleaseResponse = []*release.Release{
		{
			Name: gatewayName,
			Info: &release.Info{
				Status: release.StatusDeployed,
			}},
	}

	emptyListReleaseResponse []*release.Release
	expectedOverrides        = map[string]interface{}{"global": map[string]interface{}{}}
)

func TestGatewayManager_InstallGateway(t *testing.T) {
	t.Run("Should install Gateway", func(t *testing.T) {
		//given
		installationResponse := &release.Release{}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", gatewayChartDirectory, gatewayName, namespace, expectedOverrides, emptyProfile).Return(installationResponse, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil, emptyProfile)

		//when
		err := gatewayManager.InstallGateway(namespace)

		//then
		require.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should fail when Helm fails to install release", func(t *testing.T) {
		//given
		installationResponse := &release.Release{}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("InstallReleaseFromChart", gatewayChartDirectory, gatewayName, namespace, expectedOverrides, emptyProfile).
			Return(installationResponse, errors.New("all your base are belong to us"))

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil, emptyProfile)

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
		helmClient.On("DeleteRelease", gatewayName, namespace).Return(nil, nil)
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil, emptyProfile)

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

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil, emptyProfile)

		// when
		err := gatewayManager.DeleteGateway(namespace)

		// then
		assert.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should return error when deleteting gateway fails", func(t *testing.T) {
		// given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", gatewayName, namespace).Return(nil, errors.New("oh no"))
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil, emptyProfile)

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

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil, emptyProfile)

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

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil, emptyProfile)

		//when
		exists, status, err := gatewayManager.GatewayExists(namespace)

		//then
		require.NoError(t, err)
		require.Equal(t, status, release.StatusDeployed)
		assert.True(t, exists)
	})

	t.Run("Should return false when Gateway does not exist", func(t *testing.T) {
		//given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(emptyListReleaseResponse, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil, emptyProfile)

		//when
		exists, status, err := gatewayManager.GatewayExists(namespace)

		//then
		require.NoError(t, err)
		require.Equal(t, status, release.StatusUnknown)
		assert.False(t, exists)
	})

	t.Run("Should return error when listing Gateways fails", func(t *testing.T) {
		//given
		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(emptyListReleaseResponse, errors.New("dam, son"))

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, nil, emptyProfile)

		//when
		_, _, err := gatewayManager.GatewayExists(namespace)

		//then
		require.Error(t, err)
	})
}

func TestGatewayManager_UpgradeGateways(t *testing.T) {
	t.Run("Should update Gateway", func(t *testing.T) {
		//given
		serviceInstanceList := &v1beta1.ServiceInstanceList{
			Items: []v1beta1.ServiceInstance{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
					},
				}},
		}

		response := &release.Release{}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil).Once()
		helmClient.On("UpdateReleaseFromChart", gatewayChartDirectory, gatewayName, namespace, expectedOverrides, emptyProfile).Return(response, nil).Once()

		scClient := &mocks.ServiceInstanceClient{}
		scClient.On("List", context.Background(), metav1.ListOptions{}).Return(serviceInstanceList, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, scClient, emptyProfile)

		//when
		err := gatewayManager.UpgradeGateways()

		//then
		require.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should update two Gateways", func(t *testing.T) {
		// given
		secondNamespace := "secondnamespace"

		serviceInstanceList := &v1beta1.ServiceInstanceList{
			Items: []v1beta1.ServiceInstance{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: secondNamespace,
					},
				},
			},
		}

		secondNotEmptyListReleaseResponse := []*release.Release{
			{
				Name: getGatewayReleaseName(secondNamespace),
				Info: &release.Info{Status: release.StatusDeployed},
			},
		}

		response := &release.Release{}

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases", namespace).Return(notEmptyListReleaseResponse, nil).Once()
		helmClient.On("ListReleases", secondNamespace).Return(secondNotEmptyListReleaseResponse, nil).Once()
		helmClient.On("UpdateReleaseFromChart", gatewayChartDirectory, mock.AnythingOfType("string"), mock.AnythingOfType("string"), expectedOverrides, emptyProfile).Return(response, nil).Twice()

		scClient := &mocks.ServiceInstanceClient{}
		scClient.On("List", context.Background(), metav1.ListOptions{}).Return(serviceInstanceList, nil)

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, scClient, emptyProfile)

		//when
		err := gatewayManager.UpgradeGateways()

		//then
		require.NoError(t, err)
		helmClient.AssertExpectations(t)
	})

	t.Run("Should return error when listing Service Instances fails", func(t *testing.T) {
		//given
		helmClient := &helmmocks.HelmClient{}

		scClient := &mocks.ServiceInstanceClient{}
		scClient.On("List", context.Background(), metav1.ListOptions{}).Return(nil, errors.New("some error"))

		gatewayManager := NewGatewayManager(helmClient, OverridesData{}, scClient, emptyProfile)

		//when

		err := gatewayManager.UpgradeGateways()

		//then
		require.Error(t, err)
	})
}
