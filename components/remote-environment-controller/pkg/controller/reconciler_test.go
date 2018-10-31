package controller

import (
	"context"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-controller/pkg/controller/mocks"
	helmmocks "github.com/kyma-project/kyma/components/remote-environment-controller/pkg/kymahelm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"

)

const (
	reName            = "re-name"
)

func TestRemoteEnvironmentReconciler_Reconcile(t *testing.T) {

	t.Run("should install chart when new RE is created", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREInstance).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(false, nil)
		releaseManager.On("InstallNewREChart", reName).Return(hapi_4.Status_DEPLOYED, "Deployed", nil)

		reClient := &mocks.RemoteEnvironmentClient{}
		reClient.On("Update", mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).Return(nil, nil)

		reReconciler := NewReconciler(managerClient, releaseManager, reClient)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		releaseManager.AssertExpectations(t)
	})

	t.Run("should set access-label when new RE is created", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREWithoutAccessLabel).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(false, nil)
		releaseManager.On("InstallNewREChart", reName).Return(hapi_4.Status_DEPLOYED, "Deployed", nil)

		reClient := &mocks.RemoteEnvironmentClient{}
		reClient.On("Update", mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).Return(nil, nil)

		reReconciler := NewReconciler(managerClient, releaseManager, reClient)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		releaseManager.AssertExpectations(t)
		reClient.AssertExpectations(t)
	})

	t.Run("should delete chart when RE is deleted", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Return(errors.NewNotFound(schema.GroupResource{}, reName))

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("DeleteREChart", reName).Return(nil)

		reClient := &mocks.RemoteEnvironmentClient{}

		reReconciler := NewReconciler(managerClient, releaseManager, reClient)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		releaseManager.AssertExpectations(t)
	})

	t.Run("should update status if RE is updated", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREInstance).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(true, nil)
		releaseManager.On("CheckReleaseStatus", reName).Return(hapi_4.Status_DEPLOYED, "Installed", nil)

		reClient := &mocks.RemoteEnvironmentClient{}
		reClient.On("Update", mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).Return(nil, nil)

		reReconciler := NewReconciler(managerClient, releaseManager, reClient)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		releaseManager.AssertExpectations(t)
	})

	t.Run("should correct access-label if updated with wrong value", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREWithWrongAccessLabel).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(true, nil)
		releaseManager.On("CheckReleaseStatus", reName).Return(hapi_4.Status_DEPLOYED, "Installed", nil)

		reClient := &mocks.RemoteEnvironmentClient{}
		reClient.On("Update", mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).Return(nil, nil)

		reReconciler := NewReconciler(managerClient, releaseManager, reClient)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		releaseManager.AssertExpectations(t)
		reClient.AssertExpectations(t)
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

		releaseManager := &helmmocks.ReleaseManager{}

		reClient := &mocks.RemoteEnvironmentClient{}

		reReconciler := NewReconciler(managerClient, releaseManager, reClient)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.Error(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		releaseManager.AssertExpectations(t)
	})

	t.Run("should return error when failed to check releases existence", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.ManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREInstance).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(false, errors.NewBadRequest("error"))

		reClient := &mocks.RemoteEnvironmentClient{}

		reReconciler := NewReconciler(managerClient, releaseManager, reClient)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.Error(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		releaseManager.AssertExpectations(t)
	})
}

func getREFromArgs(args mock.Arguments) *v1alpha1.RemoteEnvironment {
	reInstance := args.Get(2).(*v1alpha1.RemoteEnvironment)
	reInstance.Name = reName
	return reInstance
}

func setupREInstance(args mock.Arguments) {
	reInstance := getREFromArgs(args)
	reInstance.Spec.AccessLabel = reName
}

func setupREWithoutAccessLabel(args mock.Arguments) {
	reInstance := getREFromArgs(args)
	reInstance.Spec.AccessLabel = ""
}

func setupREWithWrongAccessLabel(args mock.Arguments) {
	reInstance := getREFromArgs(args)
	reInstance.Spec.AccessLabel = ""
}
