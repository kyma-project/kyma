package subscription

import (
	"context"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// Name is the name of the Kyma Subscription.
	Name = "subscription"

	// Name of the corev1.Events emitted from the reconciliation process
	subReconciled      = "SubscriptionReconciled"
	subReconcileFailed = "SubscriptionReconcileFailed"

	// Finalizer for deleting Knative Subscriptions
	finalizerName = "subscription.finalizers.kyma-project.io"
)

type reconciler struct {
	client     client.Client
	recorder   record.EventRecorder
	knativeLib *util.KnativeLib
	opts       *opts.Options
}

// Verify the struct implements reconcile.Reconciler
var _ reconcile.Reconciler = &reconciler{}

func (r *reconciler) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile: ", "request", request)

	ctx := context.TODO()
	// Fetch the Subscription instance
	subscription := &eventingv1alpha1.Subscription{}
	err := r.client.Get(ctx, request.NamespacedName, subscription)

	// The Subscription may have been deleted since it was added to the workqueue. If
	// so, there is nothing to be done.
	if errors.IsNotFound(err) {
		log.Info("Could not find Subscription: ", "err", err)
		return reconcile.Result{}, nil
	}

	// Any other error should be retried in another reconciliation.
	if err != nil {
		log.Error(err, "Unable to Get Subscription object")
		return reconcile.Result{}, err
	}

	log.Info("Reconciling Subscription", "UID", string(subscription.ObjectMeta.UID))

	subscription = subscription.DeepCopy()

	requeue, reconcileErr := r.reconcile(ctx, subscription)
	if reconcileErr != nil {
		log.Error(reconcileErr, "Reconciling Subscription")
		r.recorder.Eventf(subscription, corev1.EventTypeWarning, subReconcileFailed, "Subscription reconciliation failed: %v", err)
	} else {
		log.Info("Subscription reconciled")
		r.recorder.Eventf(subscription, corev1.EventTypeNormal, subReconciled, "Subscription reconciled: %q", subscription.Name)
	}

	/*if err = util.UpdateSubscriptionStatus(ctx, r.client, subscription); err != nil {
		logging.FromContext(ctx).Info("Error updating Subscription Status", zap.Error(err))
		r.recorder.Eventf(ccp, corev1.EventTypeWarning, subUpdateStatusFailed, "Failed to update Subscription's status: %v", err)
		return reconcile.Result{}, err
	}*/

	return reconcile.Result{
		Requeue: requeue,
	}, reconcileErr
}

func (r *reconciler) reconcile(ctx context.Context, subscription *eventingv1alpha1.Subscription) (bool, error) {

	knativeSubsName := subscription.Name
	knativeSubsNamespace := subscription.Namespace
	knativeSubsURI := subscription.Endpoint
	knativeChannelName := util.GetChannelName(&subscription.SourceID, &subscription.EventType, &subscription.EventTypeVersion)
	knativeChannelProvisioner := "natss"
	timeout := r.opts.ChannelTimeout

	if subscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !util.ContainsString(&subscription.ObjectMeta.Finalizers, finalizerName) {
			subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), subscription); err != nil {
				return true, nil
			}
		}
	} else {
		// The object is being deleted
		if util.ContainsString(&subscription.ObjectMeta.Finalizers, finalizerName) {
			// our finalizer is present, so lets handle our external dependency
			if err := r.deleteExternalDependency(ctx, subscription, knativeChannelName); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return false, err
			}

			// remove our finalizer from the list and update it.
			subscription.ObjectMeta.Finalizers = util.RemoveString(&subscription.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), subscription); err != nil {
				return true, nil
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return false, nil
	}

	// Check if Kyma Subscription has events-activated condition.
	if subscription.HasCondition(eventingv1alpha1.SubscriptionCondition{Type: eventingv1alpha1.EventsActivated, Status: eventingv1alpha1.ConditionTrue}) {
		// Check if Knative Channel already exists, create if not.
		_, err := r.knativeLib.GetChannel(knativeChannelName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if errors.IsNotFound(err) {
			knativeChannel, err := r.knativeLib.CreateChannel(knativeChannelProvisioner, knativeChannelName, knativeSubsNamespace, timeout)
			if err != nil {
				return false, err
			}
			log.Info("Knative Channel is created", "Channel", knativeChannel)
		}

		// Check if Knative Subscription already exists, if not create one.
		sub, err := r.knativeLib.GetSubscription(knativeSubsName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if errors.IsNotFound(err) {
			err = r.knativeLib.CreateSubscription(knativeSubsName, knativeSubsNamespace, knativeChannelName, &knativeSubsURI)
			if err != nil {
				return false, err
			}
			log.Info("Knative Subscription is created", "Subscription", knativeSubsName)
		} else {
			// In case there is a change in Channel name or URI, delete and re-create Knative Subscription because update does not work.
			if knativeChannelName != sub.Spec.Channel.Name || knativeSubsURI != *sub.Spec.Subscriber.DNSName {
				err = r.knativeLib.DeleteSubscription(knativeSubsName, knativeSubsNamespace)
				if err != nil {
					return false, err
				}
				log.Info("Knative Subscription is deleted", "Subscription", knativeSubsName)
				err = r.knativeLib.CreateSubscription(knativeSubsName, knativeSubsNamespace, knativeChannelName, &knativeSubsURI)
				if err != nil {
					return false, err
				}
				log.Info("Knative Subscription is re-created", "Subscription", knativeSubsName)
			}
		}
	} else {
		// In case Kyma Subscription does not have events-activated condition, delete Knative Subscription if exists.
		knativeSubs, err := r.knativeLib.GetSubscription(knativeSubsName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if err == nil && knativeSubs != nil {
			err = r.knativeLib.DeleteSubscription(knativeSubsName, knativeSubsNamespace)
			if err != nil {
				return false, err
			}
			log.Info("Knative Subscription is deleted", "Subscription", knativeSubsName)
		}

		// Check if Channel has any other Subscription, if not, delete it.
		knativeChannel, err := r.knativeLib.GetChannel(knativeChannelName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if err == nil && knativeChannel != nil {
			if knativeChannel.Spec.Subscribable == nil || len(knativeChannel.Spec.Subscribable.Subscribers) == 0 ||
				(len(knativeChannel.Spec.Subscribable.Subscribers) == 1 && knativeChannel.Spec.Subscribable.Subscribers[0].SubscriberURI == subscription.Endpoint) {
				err = r.knativeLib.DeleteChannel(knativeChannelName, knativeSubsNamespace)
				if err != nil {
					return false, err
				}
				log.Info("Knative Channel is deleted", "Channel", knativeChannel)
			}
		}
	}

	return false, nil
}

func (r *reconciler) deleteExternalDependency(ctx context.Context, subscription *eventingv1alpha1.Subscription, channelName string) error {
	log.Info("Deleting the external dependencies")

	// In case Knative Subscription exists, delete it.
	knativeSubs, err := r.knativeLib.GetSubscription(subscription.Name, subscription.Namespace)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil && knativeSubs != nil {
		err = r.knativeLib.DeleteSubscription(knativeSubs.Name, knativeSubs.Namespace)
		if err != nil {
			return err
		}
		log.Info("Knative Subscription is deleted", "Subscription", knativeSubs.Name)
	}

	// Check if Channel has any other Subscription, if not, delete it.
	knativeChannel, err := r.knativeLib.GetChannel(channelName, subscription.Namespace)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil && knativeChannel != nil {
		if knativeChannel.Spec.Subscribable == nil || (len(knativeChannel.Spec.Subscribable.Subscribers) == 1 &&
			knativeChannel.Spec.Subscribable.Subscribers[0].SubscriberURI == subscription.Endpoint) {
			err = r.knativeLib.DeleteChannel(channelName, subscription.Namespace)
			if err != nil {
				return err
			}
			log.Info("Knative Channel is deleted", "Channel", knativeChannel)
		}
	}
	return nil
}
