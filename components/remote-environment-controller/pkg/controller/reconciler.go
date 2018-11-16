package controller

import (
	"context"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	reReleases "github.com/kyma-project/kyma/components/remote-environment-controller/pkg/kymahelm/remoteenvironemnts"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

	status, description, err := r.installOrGetStatus(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.updateRemoteEnv(instance, status, description)
	if err != nil {
		return reconcile.Result{}, logAndError(err, "Error while updating RE %s: %s", instance.Name, err.Error())
	}

	return reconcile.Result{}, nil
}

func (r *remoteEnvironmentReconciler) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		log.Infof("Remote Environment %s deleted", request.Name)
		err = r.releaseManager.DeleteREChart(request.Name)
		if err != nil {
			return reconcile.Result{}, logAndError(err, "Error while deleting release for %s RE: %s", request.Name, err.Error())
		}
		log.Infof("Release %s successfully deleted", request.Name)
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, logAndError(err, "Error getting %s Remote Environment: %s", request.Name, err.Error())
}

// Installs release if it does not exist or returns its status if it does
func (r *remoteEnvironmentReconciler) installOrGetStatus(remoteEnv *v1alpha1.RemoteEnvironment) (hapi_4.Status_Code, string, error) {
	releaseExist, err := r.releaseManager.CheckReleaseExistence(remoteEnv.Name)
	if err != nil {
		return hapi_4.Status_FAILED, "", err
	}

	var status hapi_4.Status_Code
	var description string

	if !releaseExist {
		log.Infof("Installing release for %s Remote Environment...", remoteEnv.Name)
		status, description, err = r.releaseManager.InstallNewREChart(remoteEnv.Name)
		if err != nil {
			return hapi_4.Status_FAILED, "", logAndError(err, "Error installing release for %s RE", remoteEnv.Name)
		}
		log.Infof("Release for RE %s, installed successfully. Release status: %s", remoteEnv.Name, status)
	} else {
		status, description, err = r.releaseManager.CheckReleaseStatus(remoteEnv.Name)
		if err != nil {
			return hapi_4.Status_FAILED, "", logAndError(err, "Error checking release status for %s RE", remoteEnv.Name)
		}
		log.Infof("Release status for %s RE: %s", remoteEnv.Name, status)
	}

	return status, description, nil
}

func (r *remoteEnvironmentReconciler) updateRemoteEnv(remoteEnv *v1alpha1.RemoteEnvironment, status hapi_4.Status_Code, description string) error {
	r.ensureAccessLabel(remoteEnv)
	r.updateREStatus(remoteEnv, status, description)

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

func (r *remoteEnvironmentReconciler) updateREStatus(remoteEnv *v1alpha1.RemoteEnvironment, status hapi_4.Status_Code, description string) {
	installationStatus := v1alpha1.InstallationStatus{
		Status:      status.String(),
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
