package controller

import (
	"context"

	reReleases "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/remoteenvironemnts"
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

type RemoteEnvironmentManagerClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	Update(ctx context.Context, obj runtime.Object) error
}

type RemoteEnvironmentReconciler interface {
	Reconcile(request reconcile.Request) (reconcile.Result, error)
}

type remoteEnvironmentReconciler struct {
	reMgrClient    RemoteEnvironmentManagerClient
	releaseManager reReleases.ReleaseManager
}

func NewReconciler(reMgrClient RemoteEnvironmentManagerClient, releaseManager reReleases.ReleaseManager) RemoteEnvironmentReconciler {
	return &remoteEnvironmentReconciler{
		reMgrClient:    reMgrClient,
		releaseManager: releaseManager,
	}
}

func (r *remoteEnvironmentReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.RemoteEnvironment{}

	err := r.reMgrClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return r.handleErrorWhileGettingInstance(err, request)
	}

	err = r.enforceDesiredState(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.updateRemoteEnv(instance)
	if err != nil {
		return reconcile.Result{}, logAndError(err, "Error while updating RE %s: %s", instance.Name, err.Error())
	}

	return reconcile.Result{}, nil
}

func (r *remoteEnvironmentReconciler) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		log.Infof("Remote Environment %s deleted", request.Name)

		releaseExist, err := r.releaseManager.CheckReleaseExistence(request.Name)
		if err != nil {
			return reconcile.Result{}, logAndError(err, "Error while checking release existence for %s RE: %s", request.Name, err.Error())
		}

		if releaseExist {
			err = r.releaseManager.DeleteREChart(request.Name)
			if err != nil {
				return reconcile.Result{}, logAndError(err, "Error while deleting release for %s RE: %s", request.Name, err.Error())
			}
			log.Infof("Release %s successfully deleted", request.Name)
		}

		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, logAndError(err, "Error getting %s Remote Environment: %s", request.Name, err.Error())
}

func (r *remoteEnvironmentReconciler) enforceDesiredState(remoteEnv *v1alpha1.RemoteEnvironment) error {
	reStatus, statusDescription, err := r.manageInstallation(remoteEnv)
	if err != nil {
		return logAndError(err, "Error managing Helm release: %s", err.Error())
	}
	log.Infof("Release status for %s Remote Environment: %s", remoteEnv.Name, reStatus)

	r.ensureAccessLabel(remoteEnv)
	r.setCurrentStatus(remoteEnv, reStatus, statusDescription)

	return nil
}

func (r *remoteEnvironmentReconciler) manageInstallation(remoteEnv *v1alpha1.RemoteEnvironment) (string, string, error) {
	releaseExist, err := r.releaseManager.CheckReleaseExistence(remoteEnv.Name)
	if err != nil {
		return "", "", err
	}

	if !releaseExist {
		if shouldSkipInstallation(remoteEnv) {
			return installationSkippedStatus, "Installation will not be performed", nil
		}

		return r.installRemoteEnvironment(remoteEnv)
	} else {
		return r.checkRemoteEnvironmentStatus(remoteEnv)
	}
}

func shouldSkipInstallation(remoteEnv *v1alpha1.RemoteEnvironment) bool {
	return remoteEnv.Spec.SkipInstallation == true
}

func (r *remoteEnvironmentReconciler) installRemoteEnvironment(remoteEnv *v1alpha1.RemoteEnvironment) (string, string, error) {
	log.Infof("Installing release for %s Remote Environment...", remoteEnv.Name)
	status, description, err := r.releaseManager.InstallNewREChart(remoteEnv.Name)
	if err != nil {
		return "", "", errors.Wrapf(err, "Error installing release for %s RE, %s", remoteEnv.Name, err.Error())
	}
	log.Infof("Release for %s Remote Environment, installed successfully", remoteEnv.Name)

	return status.String(), description, nil
}

func (r *remoteEnvironmentReconciler) checkRemoteEnvironmentStatus(remoteEnv *v1alpha1.RemoteEnvironment) (string, string, error) {
	status, description, err := r.releaseManager.CheckReleaseStatus(remoteEnv.Name)
	if err != nil {
		return "", "", errors.Wrapf(err, "Error checking release status for %s RE, %s", remoteEnv.Name, err.Error())
	}

	return status.String(), description, err
}

func (r *remoteEnvironmentReconciler) updateRemoteEnv(remoteEnv *v1alpha1.RemoteEnvironment) error {
	if remoteEnv.Spec.Services == nil {
		remoteEnv.Spec.Services = []v1alpha1.Service{}
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.reMgrClient.Update(context.Background(), remoteEnv)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *remoteEnvironmentReconciler) setCurrentStatus(remoteEnv *v1alpha1.RemoteEnvironment, status string, description string) {
	installationStatus := v1alpha1.InstallationStatus{
		Status:      status,
		Description: description,
	}

	remoteEnv.Status.InstallationStatus = installationStatus
}

func (r *remoteEnvironmentReconciler) ensureAccessLabel(remoteEnv *v1alpha1.RemoteEnvironment) {
	if remoteEnv.Spec.AccessLabel != remoteEnv.Name {
		log.Infof("Invalid access-label, setting access-label to %s", remoteEnv.Name)
		remoteEnv.Spec.AccessLabel = remoteEnv.Name
	}
}

func logAndError(err error, format string, arg ...interface{}) error {
	log.Errorf(format, arg...)
	return err
}
