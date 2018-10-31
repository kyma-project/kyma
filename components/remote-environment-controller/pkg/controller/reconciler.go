package controller

import (
	"context"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-controller/pkg/kymahelm"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
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
	mgrClient  ManagerClient
	helmClient kymahelm.HelmClient
	overrides  string
	namespace  string
	reClient   RemoteEnvironmentClient
}

func NewReconciler(mgrClient ManagerClient, helmClient kymahelm.HelmClient, reClient RemoteEnvironmentClient, overrides string, namespace string) RemoteEnvironmentReconciler {
	return &remoteEnvironmentReconciler{
		mgrClient:  mgrClient,
		helmClient: helmClient,
		reClient:   reClient,
		overrides:  overrides,
		namespace:  namespace,
	}
}

func (r *remoteEnvironmentReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.RemoteEnvironment{}

	err := r.mgrClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("Remote Environment %s deleted", request.Name)
			err = r.deleteREChart(request.Name)
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

	r.ensureAccessLabel(instance)

	releaseExist, err := r.checkReleaseExistence(instance.Name)
	if err != nil {
		return reconcile.Result{}, err
	}

	var status hapi_4.Status_Code
	var description string

	if !releaseExist {
		status, description, err = r.installNewREChart(instance.Name)
		if err != nil {
			log.Errorf("Error installing release for %s RE", instance.Name)
			return reconcile.Result{}, err
		}
		log.Info("Release for RE %s, installed successfully. Release status: %s", instance.Name, status)
	} else {
		status, description, err = r.checkReleaseStatus(instance.Name)
		if err != nil {
			log.Errorf("Error checking release status for %s RE", instance.Name)
			return reconcile.Result{}, err
		}
		log.Info("Release status for %s RE: %s", instance.Name, status)
	}

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

func (r *remoteEnvironmentReconciler) checkReleaseStatus(name string) (hapi_4.Status_Code, string, error) {
	status, err := r.helmClient.ReleaseStatus(name)
	if err != nil {
		return hapi_4.Status_UNKNOWN, "", err
	}

	return status.Info.Status.Code, status.Info.Description, nil
}

func (r *remoteEnvironmentReconciler) updateREStatus(remoteEnv *v1alpha1.RemoteEnvironment, status hapi_4.Status_Code, description string) (error){
	status, message, err := r.checkReleaseStatus(remoteEnv.Name)
	if err != nil {
		log.Errorf("Error while checking helm release status for RE %s", remoteEnv.Name)
		return err
	}

	log.Infof("Status: %s, Message: %s", status, message)
	// TODO - set Status and Conditions after updating the RE

	return nil
}

func (r *remoteEnvironmentReconciler) installNewREChart(name string) (hapi_4.Status_Code, string, error) {
	installResponse, err := r.helmClient.InstallReleaseFromChart(reChartDirectory, r.namespace, name, r.overrides)
	if err != nil {
		return hapi_4.Status_FAILED, "", err
	}

	return installResponse.Release.Info.Status.Code, installResponse.Release.Info.Description, nil
}

func (r *remoteEnvironmentReconciler) deleteREChart(name string) error {
	_, err := r.helmClient.DeleteRelease(name)
	if err != nil {
		return err
	}

	return nil
}

func (r *remoteEnvironmentReconciler) checkReleaseExistence(name string) (bool, error) {
	listResponse, err := r.helmClient.ListReleases()
	if err != nil {
		return false, err
	}

	releases := listResponse.Releases

	for _, rel := range releases {
		if rel.Name == name {
			return true, nil
		}
	}

	return false, nil
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
