package controller

import (
	"context"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-controller/pkg/controller/mocks"
	helmmocks "github.com/kyma-project/kyma/components/remote-environment-controller/pkg/kymahelm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

const (
	reName            = "re-name"
	releasesNamespace = "integration"
)

func TestRemoteEnvironmentReconciler_Reconcile(t *testing.T) {

	t.Run("should install chart when new RE is created", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		listReleaseResponse := &rls.ListReleasesResponse{
			Releases: []*release.Release{},
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREInstance).Return(nil)

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases").Return(listReleaseResponse, nil)
		helmClient.On("InstallReleaseFromChart", reChartDirectory, releasesNamespace, reName, "").Return(nil, nil)

		reReconciler := NewReconciler(managerClient, helmClient, "", releasesNamespace)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		helmClient.AssertExpectations(t)
	})

	t.Run("should delete chart when RE is deleted", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		uninstallResponse := &rls.UninstallReleaseResponse{
			Info: "uninstalled",
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Return(errors.NewNotFound(schema.GroupResource{}, reName))

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("DeleteRelease", reName).Return(uninstallResponse, nil)

		reReconciler := NewReconciler(managerClient, helmClient, "", releasesNamespace)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		helmClient.AssertExpectations(t)
	})

	t.Run("should not take action if RE is updated", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		listReleaseResponse := &rls.ListReleasesResponse{
			Count: 1,
			Releases: []*release.Release{
				{Name: reName},
			},
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREInstance).Return(nil)

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases").Return(listReleaseResponse, nil)

		reReconciler := NewReconciler(managerClient, helmClient, "", releasesNamespace)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error if error while Getting instance different than NotFound", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Return(errors.NewResourceExpired("error"))

		helmClient := &helmmocks.HelmClient{}

		reReconciler := NewReconciler(managerClient, helmClient, "", releasesNamespace)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.Error(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		helmClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to list releases", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREInstance).Return(nil)

		helmClient := &helmmocks.HelmClient{}
		helmClient.On("ListReleases").Return(nil, errors.NewBadRequest("error"))

		reReconciler := NewReconciler(managerClient, helmClient, "", releasesNamespace)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.Error(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		helmClient.AssertExpectations(t)
	})
}

func setupREInstance(args mock.Arguments) {
	reInstance := args.Get(2).(*v1alpha1.RemoteEnvironment)
	reInstance.Name = reName
}
