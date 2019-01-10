package controller

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appReleases "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/application"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	installationSkippedStatus = "INSTALLATION_SKIPPED"
	applicationFinalizer      = "finalizer.applicationconnector.kyma-project.io"
)

type ApplicationManagerClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	Update(ctx context.Context, obj runtime.Object) error
}

type ApplicationReconciler interface {
	Reconcile(request reconcile.Request) (reconcile.Result, error)
}

type applicationReconciler struct {
	applicationMgrClient ApplicationManagerClient
	releaseManager       appReleases.ReleaseManager
}

func NewReconciler(appMgrClient ApplicationManagerClient, releaseManager appReleases.ReleaseManager) ApplicationReconciler {
	return &applicationReconciler{
		applicationMgrClient: appMgrClient,
		releaseManager:       releaseManager,
	}
}

func (r *applicationReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.Application{}

	err := r.applicationMgrClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return r.handleErrorWhileGettingInstance(err, request)
	}

	err = r.enforceDesiredState(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.updateApplication(instance)
	if err != nil {
		return reconcile.Result{}, logAndError(err, "Error while updating Application %s", instance.Name)
	}

	return reconcile.Result{}, nil
}

func (r *applicationReconciler) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		log.Infof("Application %s deleted", request.Name)

		err := r.releaseManager.DeleteReleaseIfExists(request.Name)
		if err != nil {
			return reconcile.Result{}, logAndError(err, "")
		}
		log.Infof("Release %s successfully deleted", request.Name)

		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, logAndError(err, "Error getting %s Application", request.Name)
}

func (r *applicationReconciler) enforceDesiredState(application *v1alpha1.Application) error {
	if shouldBeRemoved(application) {
		return r.removeApplicationWithResources(application)
	}

	appStatus, statusDescription, err := r.manageInstallation(application)
	if err != nil {
		return logAndError(err, "Error managing Helm release for %s Application", application.Name)
	}
	log.Infof("Release status for %s Application: %s", application.Name, appStatus)

	application.SetFinalizer(applicationFinalizer)
	application.SetAccessLabel()
	r.setCurrentStatus(application, appStatus, statusDescription)

	return nil
}

func (r *applicationReconciler) removeApplicationWithResources(application *v1alpha1.Application) error {
	log.Infof("Removing %s Application with all resources...", application.Name)

	err := r.releaseManager.DeleteReleaseIfExists(application.Name)
	if err != nil {
		return errors.Wrapf(err, "Error removing %s Application with all resources", application.Name)
	}
	log.Infof("Release %s successfully deleted", application.Name)

	application.RemoveFinalizer(applicationFinalizer)
	return nil
}

func (r *applicationReconciler) manageInstallation(application *v1alpha1.Application) (string, string, error) {
	releaseExist, err := r.releaseManager.CheckReleaseExistence(application.Name)
	if err != nil {
		return "", "", err
	}

	if !releaseExist {
		if shouldSkipInstallation(application) {
			return installationSkippedStatus, "Installation will not be performed", nil
		}

		return r.installApplication(application)
	} else {
		return r.checkApplicationStatus(application)
	}
}

func shouldBeRemoved(application *v1alpha1.Application) bool {
	return application.DeletionTimestamp != nil
}

func shouldSkipInstallation(application *v1alpha1.Application) bool {
	return application.Spec.SkipInstallation == true
}

func (r *applicationReconciler) installApplication(application *v1alpha1.Application) (string, string, error) {
	log.Infof("Installing release for %s Application...", application.Name)
	status, description, err := r.releaseManager.InstallChart(application.Name)
	if err != nil {
		return "", "", errors.Wrapf(err, "Error installing release for %s Application", application.Name)
	}
	log.Infof("Release for %s Application, installed successfully", application.Name)

	return status.String(), description, nil
}

func (r *applicationReconciler) checkApplicationStatus(application *v1alpha1.Application) (string, string, error) {
	status, description, err := r.releaseManager.CheckReleaseStatus(application.Name)
	if err != nil {
		return "", "", errors.Wrapf(err, "Error checking release status for %s Application", application.Name)
	}

	return status.String(), description, err
}

func (r *applicationReconciler) updateApplication(application *v1alpha1.Application) error {
	if application.Spec.Services == nil {
		application.Spec.Services = []v1alpha1.Service{}
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.applicationMgrClient.Update(context.Background(), application)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *applicationReconciler) setCurrentStatus(application *v1alpha1.Application, status string, description string) {
	installationStatus := v1alpha1.InstallationStatus{
		Status:      status,
		Description: description,
	}

	application.SetInstallationStatus(installationStatus)
}

func logAndError(err error, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	log.Errorf("%s: %s", msg, err.Error())
	return err
}
