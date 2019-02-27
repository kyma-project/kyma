package eventactivation

import (
	"context"

	"k8s.io/client-go/tools/record"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	//appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

const (
	controllerAgentName              = "event-activation-controller"
	TestEventActivationFinalizerName = "event_activation.finalizers.kyma-project.io"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new EventActivation Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, opts *opts.Options) error {
	return add(mgr, newReconciler(mgr, opts))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, opts *opts.Options) reconcile.Reconciler {

	return &ReconcileEventActivation{
		mgr.GetClient(),
		mgr.GetScheme(),
		nil,
		opts,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("eventactivation-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to EventActivation
	err = c.Watch(&source.Kind{Type: &eventingv1alpha1.EventActivation{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileEventActivation{}

// ReconcileEventActivation reconciles a EventActivation object
type ReconcileEventActivation struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	opts     *opts.Options
}

func (r *ReconcileEventActivation) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

// Helper functions to check and remove string from a slice of strings.
//
func containsString(slice *[]string, s string) bool {
	for _, item := range *slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice *[]string, s string) (result []string) {
	for _, item := range *slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func (r *ReconcileEventActivation) deleteExternalDependency(instance *eventingv1alpha1.EventActivation) error {
	log.Info("deleting the external dependencies of the Event Activation instance")
	//
	// delete the external dependency here
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple types for same object.
	return nil
}

// TODO Controller logic
func (r *ReconcileEventActivation) controllerLogic(instance *eventingv1alpha1.EventActivation) error {
	log.Info("controller implementation")
	log.Info("event activation", "source ID", instance.SourceID)
	if instance.EventActivationSpec.SourceID == "applicationX" {
		log.Info("Taking decision A based on the Source ID: ", "source ID", instance.SourceID)
	} else {
		log.Info("Taking decision B based on the Source ID: ", "source ID", instance.SourceID)
	}
	return nil
}

// Reconcile reads that state of the cluster for a EventActivation object and makes changes based on the state read
// and what is in the EventActivation.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.
func (r *ReconcileEventActivation) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the EventActivation instance
	log.Info("Reconcile started...Here we go!")
	instance := &eventingv1alpha1.EventActivation{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{Requeue: true}, nil
	}
	if instance != nil {
		log.Info("Instance found", "UID", string(instance.ObjectMeta.UID))
		// print instance labels
		for key, value := range instance.Labels {
			log.Info("Label: ", "key", key, "value", value)
		}

	}
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// object is not being deleted, check if the finalizer is added
		if !containsString(&instance.ObjectMeta.Finalizers, TestEventActivationFinalizerName) {
			//Finalizer is not added, let's add it
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, TestEventActivationFinalizerName)
			log.Info("Finalizer added", "Finalizer name", TestEventActivationFinalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				return reconcile.Result{Requeue: true}, nil
			}
		}
		// handle object updates
		if err := r.controllerLogic(instance); err != nil {
			return reconcile.Result{Requeue: true}, nil
		}
	} else {
		//object is being deleted
		if containsString(&instance.ObjectMeta.Finalizers, TestEventActivationFinalizerName) {
			// finalizer is added, let's execute it
			if err := r.deleteExternalDependency(instance); err != nil {
				return reconcile.Result{}, err
			}
			// remove the finalizer from the list
			instance.ObjectMeta.Finalizers = removeString(&instance.ObjectMeta.Finalizers, TestEventActivationFinalizerName)
			log.Info("Finalizer removed", "Finalizer name", TestEventActivationFinalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	}
	return reconcile.Result{}, nil
}
