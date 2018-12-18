package controller

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/controller/mocks"
	helmmocks "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/remoteenvironemnts/mocks"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	reName = "re-name"
)

type reChecker struct {
	t                   *testing.T
	expectedStatus      string
	expectedDescription string
}

func (checker reChecker) checkAccessLabel(args mock.Arguments) {
	reInstance := args.Get(1).(*v1alpha1.RemoteEnvironment)

	assert.Equal(checker.t, reName, reInstance.Spec.AccessLabel)
}

func (checker reChecker) checkStatus(args mock.Arguments) {
	reInstance := args.Get(1).(*v1alpha1.RemoteEnvironment)

	assert.Equal(checker.t, checker.expectedStatus, reInstance.Status.InstallationStatus.Status)
	assert.Equal(checker.t, checker.expectedDescription, reInstance.Status.InstallationStatus.Description)
}

func TestRemoteEnvironmentReconciler_Reconcile(t *testing.T) {
	releaseStatus := hapi_4.Status_DEPLOYED
	statusDescription := "Deployed"

	statusChecker := reChecker{
		t:                   t,
		expectedStatus:      releaseStatus.String(),
		expectedDescription: statusDescription,
	}

	t.Run("should install chart when new RE is created", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREInstance).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(statusChecker.checkStatus).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(false, nil)
		releaseManager.On("InstallNewREChart", reName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, releaseManager)

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

	t.Run("should skip chart installation when skip-installation label set to true", func(t *testing.T) {
		// given
		skippedChecker := reChecker{
			t:                   t,
			expectedStatus:      installationSkippedStatus,
			expectedDescription: "Installation will not be performed",
		}

		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREWhichIsNotInstalled).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(skippedChecker.checkStatus).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(false, nil)

		reReconciler := NewReconciler(managerClient, releaseManager)

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

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREWithoutAccessLabel).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(statusChecker.checkAccessLabel).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(false, nil)
		releaseManager.On("InstallNewREChart", reName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, releaseManager)

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

	t.Run("should check status if chart exist despite skip-installation label set to true", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREWhichIsNotInstalled).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(statusChecker.checkStatus).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(true, nil)
		releaseManager.On("CheckReleaseStatus", reName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, releaseManager)

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

	t.Run("should delete chart when RE is deleted", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Return(errors.NewNotFound(schema.GroupResource{}, reName))

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(true, nil)
		releaseManager.On("DeleteREChart", reName).Return(nil)

		reReconciler := NewReconciler(managerClient, releaseManager)

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

	t.Run("should not delete chart when RE is deleted and release does not exist", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Return(errors.NewNotFound(schema.GroupResource{}, reName))

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(false, nil)

		reReconciler := NewReconciler(managerClient, releaseManager)

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

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREInstance).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(statusChecker.checkStatus).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(true, nil)
		releaseManager.On("CheckReleaseStatus", reName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, releaseManager)

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

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREWithWrongAccessLabel).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(statusChecker.checkStatus).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(true, nil)
		releaseManager.On("CheckReleaseStatus", reName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, releaseManager)

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

	t.Run("should return error if error while Getting instance different than NotFound", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Return(errors.NewResourceExpired("error"))

		releaseManager := &helmmocks.ReleaseManager{}

		reReconciler := NewReconciler(managerClient, releaseManager)

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

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREInstance).Return(nil)

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(false, errors.NewBadRequest("error"))

		reReconciler := NewReconciler(managerClient, releaseManager)

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

	t.Run("should return error when failed to update RE", func(t *testing.T) {
		namespacedName := types.NamespacedName{
			Name: reName,
		}

		managerClient := &mocks.RemoteEnvironmentManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).
			Run(setupREWithWrongAccessLabel).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).Return(errors.NewBadRequest("Error"))

		releaseManager := &helmmocks.ReleaseManager{}
		releaseManager.On("CheckReleaseExistence", reName).Return(true, nil)
		releaseManager.On("CheckReleaseStatus", reName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, releaseManager)

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

func setupREWhichIsNotInstalled(args mock.Arguments) {
	reInstance := getREFromArgs(args)
	reInstance.Spec.SkipInstallation = true
}

func setupREWithoutAccessLabel(args mock.Arguments) {
	reInstance := getREFromArgs(args)
	reInstance.Spec.AccessLabel = ""
}

func setupREWithWrongAccessLabel(args mock.Arguments) {
	reInstance := getREFromArgs(args)
	reInstance.Spec.AccessLabel = ""
}
