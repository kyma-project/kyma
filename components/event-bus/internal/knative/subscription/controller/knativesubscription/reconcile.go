package knativesubscription

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	messagingapisv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	subscriptionlistersv1alpha1 "knative.dev/eventing/pkg/client/listers/messaging/v1alpha1"
	"knative.dev/eventing/pkg/logging"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/controller"

	kymaeventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	kymaeventingclientsetv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/typed/eventing/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

const (
	finalizerName                      = "subscription.finalizers.kyma-project.io"
	knativesubscriptionreconciled      = "KnativeSubscriptionReconciled"
	knativesubscriptionreconcilefailed = "KnativeSubscriptionReconcileFailed"
)

//Reconciler Knative subscriptions reconciler
type Reconciler struct {
	// wrapper for core controller components (clients, logger, ...)
	*reconciler.Base

	// wrapper for core controller components (clients, logger, ...)
	subscriptionLister subscriptionlistersv1alpha1.SubscriptionLister

	// clients allow interactions with API objects
	kymaEventingClient kymaeventingclientsetv1alpha1.EventingV1alpha1Interface

	time util.CurrentTime
}

// Reconcile reconciles a Kn Subscription object
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	log := logging.FromContext(ctx)

	subscription, err := subscriptionByKey(key, r.subscriptionLister)
	if err != nil {
		return err
	}

	// Modify a copy, not the original.
	subscription = subscription.DeepCopy()

	// Reconcile this copy of the Knative Subscription and then write back any status
	// updates regardless of whether the reconcile error out.
	requeue, reconcileErr := r.reconcile(ctx, subscription)
	if reconcileErr != nil {
		log.Error("error in reconciling Knative Subscription", zap.Error(reconcileErr))
		r.Recorder.Eventf(subscription, corev1.EventTypeWarning, knativesubscriptionreconcilefailed, "Subscription reconciliation failed: %v", reconcileErr)
	}

	if err := util.UpdateKnativeSubscription(r.EventingClientSet.MessagingV1alpha1(), subscription); err != nil {
		log.Error("failed in updating Knative Subscription status", zap.Error(err))
		r.Recorder.Eventf(subscription, corev1.EventTypeWarning, knativesubscriptionreconcilefailed, "Updating Kn subscription status failed: %v", err)
		return err
	}

	if !requeue && reconcileErr == nil {
		log.Info("Knative subscriptions reconciled")
		r.Recorder.Eventf(subscription, corev1.EventTypeNormal, knativesubscriptionreconciled, "KnativeSubscription reconciled, name: %q; namespace: %q", subscription.Name, subscription.Namespace)
	}
	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, sub *messagingapisv1alpha1.Subscription) (bool, error) {
	var isSubReady, isKnSubReadyInSub bool
	log := logging.FromContext(ctx)
	knSubscriptionMarkedForDeletion := !sub.DeletionTimestamp.IsZero()

	kymaSub, err := util.GetKymaSubscriptionForSubscription(r.kymaEventingClient, sub)
	if err != nil {
		log.Error("GetKymaSubscriptionForSubscription() failed", zap.Error(err))
		return false, err
	}
	// there is no kyma subscription => delete finalizer from knative subscription
	if kymaSub == nil {
		sub.ObjectMeta.Finalizers = util.RemoveString(&sub.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer removed for Knative Subscription", zap.String("Finalizer name", finalizerName))
		return false, nil
	} else { // there is a kyma subscription
		// Delete finalizer when knative subscription is marked for deletion
		if knSubscriptionMarkedForDeletion {
			err := util.DeactivateSubscriptionForKnSubscription(r.kymaEventingClient, kymaSub, log, r.time)
			if err != nil {
				log.Error("DeactivateSubscriptionForKnSubscription() failed", zap.Error(err))
				return false, err
			}
			sub.ObjectMeta.Finalizers = util.RemoveString(&sub.ObjectMeta.Finalizers, finalizerName)
			log.Info("Finalizer removed for Knative Subscription", zap.String("Finalizer name", finalizerName))
			return false, nil
		} else { // Add finalizer when knative subscription is not marked for deletion
			// If we are adding the finalizer for the first time, then ensure that finalizer is persisted
			if !util.ContainsString(&sub.ObjectMeta.Finalizers, finalizerName) {
				sub.ObjectMeta.Finalizers = append(sub.ObjectMeta.Finalizers, finalizerName)
				log.Info("Finalizer added for Knative Subscription", zap.String("Finalizer name", finalizerName))
				return true, nil
			}
		}
	}

	for _, condition := range sub.Status.Conditions {
		if condition.Type == apis.ConditionReady && condition.Status == corev1.ConditionTrue {
			isSubReady = true
			break
		}
	}
	for _, cond := range kymaSub.Status.Conditions {
		if cond.Type == kymaeventingv1alpha1.SubscriptionReady && cond.Status == kymaeventingv1alpha1.ConditionTrue {
			isKnSubReadyInSub = true
			break
		}
	}
	if isSubReady && !isKnSubReadyInSub {
		if err := util.ActivateSubscriptionForKnSubscription(r.kymaEventingClient, kymaSub, log, r.time); err != nil {
			log.Error("ActivateSubscriptionForKnSubscription() failed", zap.Error(err))
			return false, err
		}
	}

	if !isSubReady && isKnSubReadyInSub {
		if err := util.DeactivateSubscriptionForKnSubscription(r.kymaEventingClient, kymaSub, log, r.time); err != nil {
			log.Error("DeactivateSubscriptionForKnSubscription() failed", zap.Error(err))
			return false, err
		}
	}
	return false, nil
}

// subscriptionByKey retrieves a Subscription object from a lister by ns/name key.
func subscriptionByKey(key string, l subscriptionlistersv1alpha1.SubscriptionLister) (*messagingv1alpha1.Subscription, error) {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, controller.NewPermanentError(pkgerrors.Wrap(err, "invalid object key"))
	}

	src, err := l.Subscriptions(ns).Get(name)
	switch {
	case apierrors.IsNotFound(err):
		return nil, controller.NewPermanentError(pkgerrors.Wrap(err, "object no longer exists"))
	case err != nil:
		return nil, err
	}

	return src, nil
}
