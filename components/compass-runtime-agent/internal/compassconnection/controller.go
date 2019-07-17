package compassconnection

import (
	"context"

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
	client Client
	scheme *runtime.Scheme
	log    *logrus.Entry
}

func InitCompassConnectionController(mgr manager.Manager) error {
	reconciler := newReconciler(mgr.GetClient())

	return startController(mgr, reconciler)
}

func startController(mgr manager.Manager, reconciler reconcile.Reconciler) error {
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.CompassConnection{}}, &handler.EnqueueRequestForObject{})
}

func newReconciler(client Client) reconcile.Reconciler {
	return &Reconciler{
		client: client,
		log:    logrus.WithField("Controller", "CompassConnection"),
	}
}

// Reconcile reads that state of the cluster for a CompassConnection object and makes changes based on the state read
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the CompassConnection instance
	instance := &v1alpha1.CompassConnection{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			r.log.Infof("Compass Connection %s deleted.", request.Name)
			// TODO - read config map
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.log.Infof("Failed to read %s Compass Connection.", request.Name)
		return reconcile.Result{}, err
	}

	r.log.Infof("Processing %s Compass Connection, current status: %s", instance.Name, "TODO")

	// TODO - fetch certificate
	// TODO - fetch config

	return reconcile.Result{}, nil
}
