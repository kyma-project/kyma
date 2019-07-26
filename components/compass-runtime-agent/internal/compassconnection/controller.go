package compassconnection

import (
	"context"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	controllerName = "compass-connection-controller"
)

type Client interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	Update(ctx context.Context, obj runtime.Object) error
	Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error
}

// Reconciler reconciles a CompassConnection object
type Reconciler struct {
	client       Client
	supervisor   Supervisor
	synchronizer synchronization.Service
	configClient compass.ConfigClient

	log *logrus.Entry
}

func InitCompassConnectionController(
	mgr manager.Manager,
	supervisior Supervisor) error {

	reconciler := newReconciler(mgr.GetClient(), supervisior)

	return startController(mgr, reconciler)
}

func startController(mgr manager.Manager, reconciler reconcile.Reconciler) error {
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.CompassConnection{}}, &handler.EnqueueRequestForObject{})
}

func newReconciler(client Client, supervisior Supervisor) reconcile.Reconciler {
	return &Reconciler{
		client:     client,
		supervisor: supervisior,
		log:        logrus.WithField("Controller", "CompassConnection"),
	}
}

// Reconcile reads that state of the cluster for a CompassConnection object and makes changes based on the state read
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithField("CompassConnection", request.Name)

	// Fetch the CompassConnection instance
	instance := &v1alpha1.CompassConnection{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Compass Connection deleted.")

			// Try to establish new connection
			instance, err := r.supervisor.InitializeCompassConnection()
			if err != nil {
				log.Errorf("Failed to initialize Compass Connection: %s", err.Error())
				return reconcile.Result{}, err
			}

			// TODO: log some human readable status
			log.Infof("Attempt to initialize Compass Connection ended with status: %s", instance.Status)

			return reconcile.Result{}, nil
		}

		// SynchronizationFailed reading the object - requeue the request.
		log.Info("Failed to read Compass Connection.")
		return reconcile.Result{}, err
	}

	// TODO: human readable status
	log.Infof("Processing %s Compass Connection, current status: %s", instance.Name, "TODO")

	// If connection is not established read Config Map and try to fetch Certificate
	if instance.ShouldReconnect() {
		instance, err := r.supervisor.InitializeCompassConnection()
		if err != nil {
			log.Errorf("Failed to initialize Compass Connection: %s", err.Error())
			return reconcile.Result{}, err
		}

		log.Infof("Attempt to initialize Compass Connection ended with status: %s", instance.Status)
		return reconcile.Result{}, nil
	}

	// TODO: check last synchronization time and if it passed

	log.Info("Trying to connect to Compass and apply Runtime configuration...")

	synchronized, err := r.supervisor.SynchronizeWithCompass(instance)
	if err != nil {
		log.Errorf("Failed to synchronize with Compass: %s", err.Error())
		return reconcile.Result{}, err
	}

	log.Infof("Synchronization finished. Compass Connection status: %s", synchronized.Status)

	return reconcile.Result{}, nil
}
