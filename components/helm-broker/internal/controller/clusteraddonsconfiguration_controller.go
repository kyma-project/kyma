package controller

import (
	"context"
	"github.com/go-logr/logr"
	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//
type ClusterAddonsConfigurationController struct {
	reconciler reconcile.Reconciler
}

//
func NewClusterAddonsConfigurationController(reconciler reconcile.Reconciler) *ClusterAddonsConfigurationController {
	return &ClusterAddonsConfigurationController{reconciler: reconciler}
}

//
func (cacc *ClusterAddonsConfigurationController) Start(mgr manager.Manager) error {
	// Create a new controller
	c, err := controller.New("clusteraddonsconfiguration-controller", mgr, controller.Options{Reconciler: cacc.reconciler})
	if err != nil {
		return err
	}

	// Watch for changes to ClusterAddonsConfiguration
	err = c.Watch(&source.Kind{Type: &addonsv1alpha1.ClusterAddonsConfiguration{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileClusterAddonsConfiguration{}

// ReconcileClusterAddonsConfiguration reconciles a ClusterAddonsConfiguration object
type ReconcileClusterAddonsConfiguration struct {
	log logr.Logger
	client.Client
	scheme *runtime.Scheme
}

// newReconciler returns a new reconcile.Reconciler
func NewReconcileClusterAddonsConfiguration(mgr manager.Manager, log logr.Logger) reconcile.Reconciler {
	return &ReconcileClusterAddonsConfiguration{
		log:    log,
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

// Reconcile reads that state of the cluster for a ClusterAddonsConfiguration object and makes changes based on the state read
// and what is in the ClusterAddonsConfiguration.Spec
func (r *ReconcileClusterAddonsConfiguration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the ClusterAddonsConfiguration instance
	instance := &addonsv1alpha1.ClusterAddonsConfiguration{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
