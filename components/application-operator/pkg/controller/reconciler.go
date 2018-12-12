package controller

import (
	"context"

	reReleases "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/application"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
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
	releaseManager       reReleases.ReleaseManager
}

func NewReconciler(reMgrClient ApplicationManagerClient, releaseManager reReleases.ReleaseManager) ApplicationReconciler {
	return &applicationReconciler{
		applicationMgrClient: reMgrClient,
		releaseManager:       releaseManager,
	}
}

func (r *applicationReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.RemoteEnvironment{}

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
		return reconcile.Result{}, logAndError(err, "Error while updating RE %s: %s", instance.Name, err.Error())
	}

	return reconcile.Result{}, nil
}

func (r *applicationReconciler) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		log.Infof("Application %s deleted", request.Name)

		releaseExist, err := r.releaseManager.CheckReleaseExistence(request.Name)
		if err != nil {
			return reconcile.Result{}, logAndError(err, "Error while checking release existence for %s RE: %s", request.Name, err.Error())
		}

		if releaseExist {
			err = r.releaseManager.DeleteChart(request.Name)
			if err != nil {
				return reconcile.Result{}, logAndError(err, "Error while deleting release for %s RE: %s", request.Name, err.Error())
			}
			log.Infof("Release %s successfully deleted", request.Name)
		}

		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, logAndError(err, "Error getting %s Application: %s", request.Name, err.Error())
}

func (r *applicationReconciler) enforceDesiredState(application *v1alpha1.RemoteEnvironment) error {
	appStatus, statusDescription, err := r.manageInstallation(application)
	if err != nil {
		return logAndError(err, "Error managing Helm release: %s", err.Error())
	}
	log.Infof("Release status for %s Application: %s", application.Name, appStatus)

	r.ensureAccessLabel(application)
	r.setCurrentStatus(application, appStatus, statusDescription)

	return nil
}

func (r *applicationReconciler) manageInstallation(application *v1alpha1.RemoteEnvironment) (string, string, error) {
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

func shouldSkipInstallation(application *v1alpha1.RemoteEnvironment) bool {
	return application.Spec.SkipInstallation == true
}

func (r *applicationReconciler) installApplication(application *v1alpha1.RemoteEnvironment) (string, string, error) {
	log.Infof("Installing release for %s Application...", application.Name)
	status, description, err := r.releaseManager.InstallChart(application.Name)
	if err != nil {
		return "", "", errors.Wrapf(err, "Error installing release for %s RE, %s", application.Name, err.Error())
	}
	log.Infof("Release for %s Application, installed successfully", application.Name)

	return status.String(), description, nil
}

func (r *applicationReconciler) checkApplicationStatus(application *v1alpha1.RemoteEnvironment) (string, string, error) {
	status, description, err := r.releaseManager.CheckReleaseStatus(application.Name)
	if err != nil {
		return "", "", errors.Wrapf(err, "Error checking release status for %s RE, %s", application.Name, err.Error())
	}

	return status.String(), description, err
}

func (r *applicationReconciler) updateApplication(application *v1alpha1.RemoteEnvironment) error {
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

func (r *applicationReconciler) setCurrentStatus(application *v1alpha1.RemoteEnvironment, status string, description string) {
	installationStatus := v1alpha1.InstallationStatus{
		Status:      status,
		Description: description,
	}

	application.Status.InstallationStatus = installationStatus
}

func (r *applicationReconciler) ensureAccessLabel(application *v1alpha1.RemoteEnvironment) {
	if application.Spec.AccessLabel != application.Name {
		log.Infof("Invalid access-label, setting access-label to %s", application.Name)
		application.Spec.AccessLabel = application.Name
	}
}

func logAndError(err error, format string, arg ...interface{}) error {
	log.Errorf(format, arg...)
	return err
}
