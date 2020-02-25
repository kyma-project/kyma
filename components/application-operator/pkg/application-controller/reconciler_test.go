package application_controller

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/application-controller/mocks"
	helmmocks "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/application/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	applicationName = "app-name"
)

type applicationChecker struct {
	t                   *testing.T
	expectedStatus      string
	expectedDescription string
}

func (checker applicationChecker) checkAccessLabel(args mock.Arguments) {
	appInstance := args.Get(1).(*v1alpha1.Application)

	assert.Equal(checker.t, applicationName, appInstance.Spec.AccessLabel)
}

func (checker applicationChecker) checkStatus(args mock.Arguments) {
	appInstance := args.Get(1).(*v1alpha1.Application)

	assert.Equal(checker.t, checker.expectedStatus, appInstance.Status.InstallationStatus.Status)
	assert.Equal(checker.t, checker.expectedDescription, appInstance.Status.InstallationStatus.Description)
}

func TestApplicationReconciler_Reconcile(t *testing.T) {
	releaseStatus := hapi_4.Status_DEPLOYED
	statusDescription := "Deployed"

	statusChecker := applicationChecker{
		t:                   t,
		expectedStatus:      releaseStatus.String(),
		expectedDescription: statusDescription,
	}

	logger := logrus.WithField("controller", "Application Tests")

	t.Run("should install chart when new application is created", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppInstance).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application")).
			Run(statusChecker.checkStatus).Return(nil)

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("CheckReleaseExistence", applicationName).Return(false, nil)
		ApplicationReleaseManager.On("InstallChart", mock.AnythingOfType("*v1alpha1.Application")).Return(releaseStatus, statusDescription, nil)

		applicationReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := applicationReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should skip chart installation when skip-installation label set to true", func(t *testing.T) {
		// given
		skippedChecker := applicationChecker{
			t:                   t,
			expectedStatus:      installationSkippedStatus,
			expectedDescription: "Installation will not be performed",
		}

		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppWhichIsNotInstalled).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application")).
			Run(skippedChecker.checkStatus).Return(nil)

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("CheckReleaseExistence", applicationName).Return(false, nil)

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should set access-label when new Application is created", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppWithoutAccessLabel).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application")).
			Run(statusChecker.checkAccessLabel).Return(nil)

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("CheckReleaseExistence", applicationName).Return(false, nil)
		ApplicationReleaseManager.On("InstallChart", mock.AnythingOfType("*v1alpha1.Application")).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should check status if chart exist despite skip-installation label set to true", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppWhichIsNotInstalled).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application")).
			Run(statusChecker.checkStatus).Return(nil)

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("CheckReleaseExistence", applicationName).Return(true, nil)
		ApplicationReleaseManager.On("CheckReleaseStatus", applicationName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should delete chart when Application is deleted", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Return(errors.NewNotFound(schema.GroupResource{}, applicationName))

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("DeleteReleaseIfExists", applicationName).Return(nil)

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should delete chart when deletion timestamp is set", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		checkFinilizerRemoved := func(args mock.Arguments) {
			appInstance := args.Get(1).(*v1alpha1.Application)
			assert.Empty(t, appInstance.Finalizers)
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppWithDeletionTimestamp).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application")).
			Run(checkFinilizerRemoved).Return(nil)

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("DeleteReleaseIfExists", applicationName).Return(nil)

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should update status if Application is updated", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppInstance).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application")).
			Run(statusChecker.checkStatus).Return(nil)

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("CheckReleaseExistence", applicationName).Return(true, nil)
		ApplicationReleaseManager.On("CheckReleaseStatus", applicationName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should correct access-label if updated with wrong value", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppWithWrongAccessLabel).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application")).
			Run(statusChecker.checkStatus).Return(nil)

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("CheckReleaseExistence", applicationName).Return(true, nil)
		ApplicationReleaseManager.On("CheckReleaseStatus", applicationName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should return error if error while Getting instance different than NotFound", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Return(errors.NewResourceExpired("error"))

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.Error(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should return error when failed to check releases existence", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppInstance).Return(nil)

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("CheckReleaseExistence", applicationName).Return(false, errors.NewBadRequest("error"))

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.Error(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should return error when release installation failed", func(t *testing.T) {
		// given
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppInstance).Return(nil)

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("CheckReleaseExistence", applicationName).Return(false, nil)
		ApplicationReleaseManager.On("InstallChart", mock.AnythingOfType("*v1alpha1.Application")).Return(hapi_4.Status_FAILED, "", errors.NewBadRequest("error"))

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.Error(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})

	t.Run("should return error when failed to update Application", func(t *testing.T) {
		namespacedName := types.NamespacedName{
			Name: applicationName,
		}

		managerClient := &mocks.ApplicationManagerClient{}
		managerClient.On(
			"Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.Application")).
			Run(setupAppWithWrongAccessLabel).Return(nil)
		managerClient.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application")).Return(errors.NewBadRequest("Error"))

		ApplicationReleaseManager := &helmmocks.ApplicationReleaseManager{}
		ApplicationReleaseManager.On("CheckReleaseExistence", applicationName).Return(true, nil)
		ApplicationReleaseManager.On("CheckReleaseStatus", applicationName).Return(releaseStatus, statusDescription, nil)

		reReconciler := NewReconciler(managerClient, ApplicationReleaseManager, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		// when
		result, err := reReconciler.Reconcile(request)

		// then
		assert.Error(t, err)
		assert.NotNil(t, result)
		managerClient.AssertExpectations(t)
		ApplicationReleaseManager.AssertExpectations(t)
	})
}

func getAppFromArgs(args mock.Arguments) *v1alpha1.Application {
	appInstance := args.Get(2).(*v1alpha1.Application)
	appInstance.Name = applicationName
	return appInstance
}

func setupAppInstance(args mock.Arguments) {
	reInstance := getAppFromArgs(args)
	reInstance.Spec.AccessLabel = applicationName
}

func setupAppWhichIsNotInstalled(args mock.Arguments) {
	appInstance := getAppFromArgs(args)
	appInstance.Spec.SkipInstallation = true
}

func setupAppWithoutAccessLabel(args mock.Arguments) {
	appInstance := getAppFromArgs(args)
	appInstance.Spec.AccessLabel = ""
}

func setupAppWithWrongAccessLabel(args mock.Arguments) {
	appInstance := getAppFromArgs(args)
	appInstance.Spec.AccessLabel = ""
}

func setupAppWithDeletionTimestamp(args mock.Arguments) {
	appInstance := getAppFromArgs(args)
	appInstance.DeletionTimestamp = &v1.Time{}
}
