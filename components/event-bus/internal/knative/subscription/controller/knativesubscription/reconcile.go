package knativesubscription

import (
	"context"

	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	finalizerName = "subscription.finalizers.kyma-project.io"
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

// Reconcile reconciles a Kn Subscription object
func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile: ", "request", request)

	ctx := context.TODO()
	sub := &evapisv1alpha1.Subscription{}
	err := r.client.Get(ctx, request.NamespacedName, sub)

	// The Knative Subscription may have been deleted since it was added to the work queue. If so, there is
	// nothing to be done.
	if errors.IsNotFound(err) {
		log.Info("Could not find Knative subscription: ", "err", err)
		return reconcile.Result{}, nil
	}

	if err != nil {
		log.Error(err, "unable to Get Knative subscription object")
		return reconcile.Result{}, err
	}

	// Modify a copy, not the original.
	sub = sub.DeepCopy()

	// Reconcile this copy of the EventActivation and then write back any status
	// updates regardless of whether the reconcile error out.
	requeue, reconcileErr := r.reconcile(ctx, sub)
	if reconcileErr != nil {
		log.Error(reconcileErr, "error in reconciling Subscription")
		r.recorder.Eventf(sub, corev1.EventTypeWarning, "SubscriptionReconcileFailed", "Subscription reconciliation failed: %v", reconcileErr)
	}

	if updateStatusErr := util.UpdateKnativeSubscription(ctx, r.client, sub); updateStatusErr != nil {
		log.Error(updateStatusErr, "failed in updating Knative Subscription status")
		r.recorder.Eventf(sub, corev1.EventTypeWarning, "KnativeSubscriptionReconcileFailed", "Updating Kn subscription status failed: %v", updateStatusErr)
		return reconcile.Result{}, updateStatusErr
	}

	if !requeue && reconcileErr == nil {
		log.Info("Knative subscriptions reconciled")
		r.recorder.Eventf(sub, corev1.EventTypeNormal, "KnativeSubscriptionReconciled", "KnativeSubscription reconciled, name: %q; namespace: %q", sub.Name, sub.Namespace)
	}
	return reconcile.Result{
		Requeue: requeue,
	}, reconcileErr
}

func (r *reconciler) reconcile(ctx context.Context, sub *evapisv1alpha1.Subscription) (bool, error) {
	var isSubReady, isKnSubReadyInSub bool

	kymaSub, err := util.GetKymaSubscriptionForSubscription(ctx, r.client, sub)
	if err != nil {
		log.Error(err, "GetKymaSubscriptionForSubscription() failed")
		return false, err
	}
	if kymaSub == nil {
		log.Info("No matching Kyma subscription found for Knative subscription: " + sub.Namespace + "/" + sub.Name)
		return false, nil
	}

	// Delete or add finalizers
	if !sub.DeletionTimestamp.IsZero() {
		err := util.DeactivateSubscriptionForKnSubscription(ctx, r.client, kymaSub, log, r.time)
		if err != nil {
			log.Error(err, "DeactivateSubscriptionForKnSubscription() failed")
			return false, err
		}
		sub.ObjectMeta.Finalizers = util.RemoveString(&sub.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer removed for Knative Subscription", "Finalizer name", finalizerName)
		return false, nil
	}

	// If we are adding the finalizer for the first time, then ensure that finalizer is persisted
	if !util.ContainsString(&sub.ObjectMeta.Finalizers, finalizerName) {
		sub.ObjectMeta.Finalizers = append(sub.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer added for Knative Subscription", "Finalizer name", finalizerName)
		return true, nil
	}

	for _, condition := range sub.Status.Conditions {
		if condition.Type == apis.ConditionReady && condition.Status == corev1.ConditionTrue {
			isSubReady = true
			break
		}
	}
	for _, cond := range kymaSub.Status.Conditions {
		if cond.Type == subApis.SubscriptionReady && cond.Status == subApis.ConditionTrue {
			isKnSubReadyInSub = true
			break
		}
	}
	if isSubReady && !isKnSubReadyInSub {
		if err := util.ActivateSubscriptionForKnSubscription(ctx, r.client, kymaSub, log, r.time); err != nil {
			log.Error(err, "ActivateSubscriptionForKnSubscription() failed")
			return false, err
		}
	}

	if !isSubReady && isKnSubReadyInSub {
		if err := util.DeactivateSubscriptionForKnSubscription(ctx, r.client, kymaSub, log, r.time); err != nil {
			log.Error(err, "DeactivateSubscriptionForKnSubscription() failed")
			return false, err
		}
	}
	return false, nil
}
