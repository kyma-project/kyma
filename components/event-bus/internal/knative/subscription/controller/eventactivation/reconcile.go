package eventactivation

import (
	"context"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	finalizerName = "eventactivation.finalizers.kyma-project.io"
)

type reconciler struct {
	client   client.Client
	recorder record.EventRecorder
	time     util.CurrentTime
}

// Verify the struct implements reconcile.Reconciler
var _ reconcile.Reconciler = &reconciler{}

func (r *reconciler) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

// Reconcile reconciles a EventActivation object
func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile: ", "request", request)

	ctx := context.TODO()
	ea := &eventingv1alpha1.EventActivation{}
	err := r.client.Get(ctx, request.NamespacedName, ea)

	// The EventActivation may have been deleted since it was added to the workqueue. If so, there is
	// nothing to be done.
	if errors.IsNotFound(err) {
		log.Info("Could not find EventActivation: ", "err", err)
		return reconcile.Result{}, nil
	}

	// Any other error should be retrieved in another reconciliation. ???
	if err != nil {
		log.Error(err, "Unable to Get EventActivation object")
		return reconcile.Result{}, err
	}

	log.Info("Reconciling Event Activation", "UID", string(ea.ObjectMeta.UID))

	// Modify a copy, not the original.
	ea = ea.DeepCopy()

	// Reconcile this copy of the EventActivation and then write back any status
	// updates regardless of whether the reconcile error out.
	requeue, reconcileErr := r.reconcile(ctx, ea)
	if reconcileErr != nil {
		log.Error(reconcileErr, "Reconciling EventActivation")
		r.recorder.Eventf(ea, corev1.EventTypeWarning, "EventactivationReconcileFailed", "Eventactivation reconciliation failed: %v", reconcileErr)
	}

	if updateStatusErr := util.UpdateEventActivation(ctx, r.client, ea); updateStatusErr != nil {
		log.Error(updateStatusErr, "Updating EventActivation status")
		r.recorder.Eventf(ea, corev1.EventTypeWarning, "EventactivationReconcileFailed", "Updating EventActivation status failed: %v", updateStatusErr)
		return reconcile.Result{}, updateStatusErr
	}

	if !requeue && reconcileErr == nil {
		log.Info("EventActivation reconciled")
		r.recorder.Eventf(ea, corev1.EventTypeNormal, "EventactivationReconciled", "EventActivation reconciled, name: %q; namespace: %q", ea.Name, ea.Namespace)
	}
	return reconcile.Result{
		Requeue: requeue,
	}, reconcileErr
}

func (r *reconciler) reconcile(ctx context.Context, ea *eventingv1alpha1.EventActivation) (bool, error) {

	// delete or add finalizers
	if !ea.DeletionTimestamp.IsZero() {
		// deactivate all Kyma subscriptions related to this ea
		subs, _ := util.GetSubscriptionsForEventActivation(ctx, r.client, ea)
		util.DeactivateSubscriptions(ctx, r.client, subs, log, r.time)

		// remove the finalizer from the list
		ea.ObjectMeta.Finalizers = util.RemoveString(&ea.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer removed", "Finalizer name", finalizerName)
		return false, nil
	}

	// If we are adding the finalizer for the first time, then ensure that finalizer is persisted
	if !util.ContainsString(&ea.ObjectMeta.Finalizers, finalizerName) {
		//Finalizer is not added, let's add it
		ea.ObjectMeta.Finalizers = append(ea.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer added", "Finalizer name", finalizerName)
		return true, nil
	}

	// check an activate, if necessary, all the subscriptions
	if subs, err := util.GetSubscriptionsForEventActivation(ctx, r.client, ea); err != nil {
		log.Error(err, "GetSubscriptionsForEventActivation() failed")
	} else {
		log.Info("Kyma subscriptions found: ", "subs", subs)
		// activate all subscriptions
		if err := util.ActivateSubscriptions(ctx, r.client, subs, log, r.time); err != nil {
			log.Error(err, "ActivateSubscriptions() failed")
			return false, err
		}
	}
	return false, nil
}
