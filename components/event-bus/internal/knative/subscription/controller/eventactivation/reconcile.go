package eventactivation

import (
	"context"
	eventingclientv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/typed/eventing/v1alpha1"
	"knative.dev/eventing/pkg/logging"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/controller"


	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	applicationconnectorclientv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/typed/applicationconnector/v1alpha1"
	applicationconnectorlistersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/lister/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

const (
	finalizerName = "eventactivation.finalizers.kyma-project.io"
)

type Reconciler struct {
	// wrapper for core controller components (clients, logger, ...)
	*reconciler.Base

	// listers index properties about resources
	eventActivationLister applicationconnectorlistersv1alpha1.EventActivationLister

	// clients allow interactions with API objects
	applicationconnectorClient applicationconnectorclientv1alpha1.ApplicationconnectorV1alpha1Interface

	eventingClient eventingclientv1alpha1.EventingV1alpha1Interface

}

// Reconcile reconciles a EventActivation object
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	ea, err := eventActivationByKey(key, r.eventActivationLister)
	if err != nil {
		return err
	}

	// Modify a copy, not the original.
	ea = ea.DeepCopy()

	// Reconcile this copy of the EventActivation and then write back any status
	// updates regardless of whether the reconcile error out.
	requeue, reconcileErr := r.reconcile(ctx, ea)
	if reconcileErr != nil {
		r.Recorder.Eventf(ea, corev1.EventTypeWarning, "EventactivationReconcileFailed",
			"Eventactivation reconciliation failed: %v", reconcileErr)
	}

	// FIXME
	if updateStatusErr := util.UpdateEventActivation(r.applicationconnectorClient, ea); updateStatusErr != nil {
		r.Recorder.Eventf(ea, corev1.EventTypeWarning, "EventactivationReconcileFailed", "Updating EventActivation status failed: %v", updateStatusErr)
		return updateStatusErr
	}

	if !requeue && reconcileErr == nil {
		r.Recorder.Eventf(ea, corev1.EventTypeNormal, "EventactivationReconciled", "EventActivation reconciled, name: %q; namespace: %q", ea.Name, ea.Namespace)
	}

	return reconcileErr
}

// eventActivationByKey retrieves a EventActivation object from a lister by ns/name key.
func eventActivationByKey(key string, l applicationconnectorlistersv1alpha1.EventActivationLister) (*applicationconnectorv1alpha1.EventActivation, error) {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, controller.NewPermanentError(pkgerrors.Wrap(err, "invalid object key"))
	}

	src, err := l.EventActivations(ns).Get(name)
	switch {
	case apierrors.IsNotFound(err):
		return nil, controller.NewPermanentError(pkgerrors.Wrap(err, "object no longer exists"))
	case err != nil:
		return nil, err
	}

	return src, nil
}

//// Verify the struct implements reconcile.Reconciler
//var _ reconcile.Reconciler = &reconciler{}
//
//func (r *reconciler) InjectClient(c client.Client) error {
//	r.client = c
//	return nil
//}
//
//// Reconcile reconciles a EventActivation object
//func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
//	log.Info("Reconcile: ", "request", request)
//
//	ctx := context.TODO()
//	ea := &eventingv1alpha1.EventActivation{}
//	err := r.client.Get(ctx, request.NamespacedName, ea)
//
//	// The EventActivation may have been deleted since it was added to the workqueue. If so, there is
//	// nothing to be done.
//	if errors.IsNotFound(err) {
//		log.Info("Could not find EventActivation: ", "err", err)
//		return reconcile.Result{}, nil
//	}
//
//	// Any other error should be retrieved in another reconciliation. ???
//	if err != nil {
//		log.Error(err, "Unable to Get EventActivation object")
//		return reconcile.Result{}, err
//	}
//
//	log.Info("Reconciling Event Activation", "UID", string(ea.ObjectMeta.UID))
//
//	// Modify a copy, not the original.
//	ea = ea.DeepCopy()
//
//	// Reconcile this copy of the EventActivation and then write back any status
//	// updates regardless of whether the reconcile error out.
//	requeue, reconcileErr := r.reconcile(ctx, ea)
//	if reconcileErr != nil {
//		log.Error(reconcileErr, "Reconciling EventActivation")
//		r.recorder.Eventf(ea, corev1.EventTypeWarning, "EventactivationReconcileFailed", "Eventactivation reconciliation failed: %v", reconcileErr)
//	}
//
//	if updateStatusErr := util.UpdateEventActivation(ctx, r.client, ea); updateStatusErr != nil {
//		log.Error(updateStatusErr, "Updating EventActivation status")
//		r.recorder.Eventf(ea, corev1.EventTypeWarning, "EventactivationReconcileFailed", "Updating EventActivation status failed: %v", updateStatusErr)
//		return reconcile.Result{}, updateStatusErr
//	}
//
//	if !requeue && reconcileErr == nil {
//		log.Info("EventActivation reconciled")
//		r.recorder.Eventf(ea, corev1.EventTypeNormal, "EventactivationReconciled", "EventActivation reconciled, name: %q; namespace: %q", ea.Name, ea.Namespace)
//	}
//	return reconcile.Result{
//		Requeue: requeue,
//	}, reconcileErr
//}
//

func (r *Reconciler) reconcile(ctx context.Context, ea *applicationconnectorv1alpha1.EventActivation) (bool, error) {
	// delete or add finalizers
	if !ea.DeletionTimestamp.IsZero() {
		// deactivate all Kyma subscriptions related to this ea
		subs, _ := util.GetSubscriptionsForEventActivation(r.eventingClient, ea)
		util.DeactivateSubscriptions(r.eventingClient, subs, logging.FromContext(ctx), r.time)

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
