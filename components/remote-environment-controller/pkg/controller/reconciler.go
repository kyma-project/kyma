package controller

import (
	"context"
	remoteenvironmentv1alpha1 "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-controller/pkg/kymahelm"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	reChartDirectory = "remote-environments"
)

type ManagerClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
}

type RemoteEnvironmentReconciler interface {
	Reconcile(request reconcile.Request) (reconcile.Result, error)
}

type remoteEnvironmentReconciler struct {
	mgrClient  ManagerClient
	helmClient kymahelm.HelmClient
	overrides  string
	namespace  string
}

func NewReconciler(mgrClient ManagerClient, helmClient kymahelm.HelmClient, overrides string, namespace string) RemoteEnvironmentReconciler {
	return &remoteEnvironmentReconciler{
		mgrClient:  mgrClient,
		helmClient: helmClient,
		overrides:  overrides,
		namespace:  namespace,
	}
}

func (r *remoteEnvironmentReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &remoteenvironmentv1alpha1.RemoteEnvironment{}

	err := r.mgrClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("Remote Environment %s deleted", request.Name)
			r.deleteREChart(request.Name)
			return reconcile.Result{}, nil
		}
		log.Errorf("Error getting %s Remote Environment: %s", request.Name, err.Error())
		return reconcile.Result{}, err
	}

	releaseExist, err := r.checkReleaseExistence(request.Name)
	if err != nil {
		return reconcile.Result{}, err
	}

	if releaseExist {
		// Handles cases: RE with chart exist on controller startup, RE is updated, full RE chart is installed
		log.Infof("Helm chart for %s Remote Environment already exists", request.Name)
	} else {
		r.installNewREChart(request.Name)
	}

	return reconcile.Result{}, nil
}

// TODO: consider returning error to requeue the request
// Note that reoccurring error may keep requeuing the same request
func (r *remoteEnvironmentReconciler) installNewREChart(name string) {
	_, err := r.helmClient.InstallReleaseFromChart(reChartDirectory, r.namespace, name, r.overrides)
	if err != nil {
		log.Error("Error while installing release for %s RE: %s", name, err.Error())
	} else {
		log.Infof("New Helm chart for %s Remote Environment installed", name)
	}
}

// TODO: consider returning error to requeue the request
func (r *remoteEnvironmentReconciler) deleteREChart(name string) {
	_, err := r.helmClient.DeleteRelease(name)
	if err != nil {
		log.Error("Error while deleting release for %s RE: %s", name, err.Error())
	} else {
		log.Infof("Release %s successfully deleted", name)
	}
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
