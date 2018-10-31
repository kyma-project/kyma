package controller

import (
	"context"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-controller/pkg/kymahelm"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	reChartDirectory = "remote-environments"
)

type ManagerClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
}

type RemoteEnvironmentClient interface {
	Update(*v1alpha1.RemoteEnvironment) (*v1alpha1.RemoteEnvironment, error)
}

type RemoteEnvironmentReconciler interface {
	Reconcile(request reconcile.Request) (reconcile.Result, error)
}

type remoteEnvironmentReconciler struct {
	mgrClient      ManagerClient
	releaseManager kymahelm.ReleaseManager
	reClient       RemoteEnvironmentClient
}

func NewReconciler(mgrClient ManagerClient, releaseManager kymahelm.ReleaseManager, reClient RemoteEnvironmentClient) RemoteEnvironmentReconciler {
	return &remoteEnvironmentReconciler{
		mgrClient: mgrClient,
		releaseManager: releaseManager,
		reClient:  reClient,
	}
}

func (r *remoteEnvironmentReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.RemoteEnvironment{}

	err := r.mgrClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("Remote Environment %s deleted", request.Name)
			err = r.releaseManager.DeleteREChart(request.Name)
			if err != nil {
				log.Errorf("Error while deleting release for %s RE: %s", request.Name, err.Error())
				return reconcile.Result{}, err
			}
			log.Infof("Release %s successfully deleted", request.Name)
			return reconcile.Result{}, nil
		}
		log.Errorf("Error getting %s Remote Environment: %s", request.Name, err.Error())
		return reconcile.Result{}, err
	}

	status, description, err := r.installOrGetStatus(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	r.ensureAccessLabel(instance)

	err = r.updateREStatus(instance, status, description)
	if err != nil {
		log.Errorf("Error while updating status of %s Remote Environment", instance.Name)
		return reconcile.Result{}, err
	}

	// TODO - consider updating with retries
	_, err = r.reClient.Update(instance)
	if err != nil {
		log.Errorf("Error while updating RE %s", instance.Name, err.Error())
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
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
		status, description, err = r.releaseManager.InstallNewREChart(remoteEnv.Name)
		if err != nil {
			log.Errorf("Error installing release for %s RE", remoteEnv.Name)
			return hapi_4.Status_FAILED, "", err
		}
		log.Info("Release for RE %s, installed successfully. Release status: %s", remoteEnv.Name, status)
	} else {
		status, description, err = r.releaseManager.CheckReleaseStatus(remoteEnv.Name)
		if err != nil {
			log.Errorf("Error checking release status for %s RE", remoteEnv.Name)
			return hapi_4.Status_FAILED, "", err
		}
		log.Info("Release status for %s RE: %s", remoteEnv.Name, status)
	}

	return status, description, nil
}

func (r *remoteEnvironmentReconciler) updateREStatus(remoteEnv *v1alpha1.RemoteEnvironment, status hapi_4.Status_Code, description string) error {
	log.Infof("Updating status. Status: %s, Message: %s", status, description)
	// TODO - set Status and Conditions after updating the RE

	return nil
}

func (r *remoteEnvironmentReconciler) ensureAccessLabel(remoteEnv *v1alpha1.RemoteEnvironment) {
	if remoteEnv.Spec.AccessLabel != remoteEnv.Name {
		log.Infof("Invalid access-label, setting access-label to %s", remoteEnv.Name)

		remoteEnv.Spec.AccessLabel = remoteEnv.Name
		if remoteEnv.Spec.Services == nil {
			remoteEnv.Spec.Services = []v1alpha1.Service{}
		}
	}
}
