package subscription

import (
	"context"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/metrics"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// Name of the corev1.Events emitted from the reconciliation process
	subReconciled      = "SubscriptionReconciled"
	subReconcileFailed = "SubscriptionReconcileFailed"

	// Finalizer for deleting Knative Subscriptions
	finalizerName = "subscription.finalizers.kyma-project.io"
)

type reconciler struct {
	client     client.Client
	recorder   record.EventRecorder
	knativeLib util.KnativeAccessLib
	opts       *opts.Options
	time       util.CurrentTime
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
	KymaSubscriptionsGauge := metrics.KymaSubscriptionsGaugeObj

	err := r.client.Get(ctx, request.NamespacedName, subscription)

	// The Subscription may have been deleted since it was added to the work queue. If
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

	requeue, reconcileErr := r.reconcile(ctx, subscription, KymaSubscriptionsGauge)
	if reconcileErr != nil {
		log.Error(reconcileErr, "Reconciling Subscription")
		if err := util.SetNotReadySubscription(ctx, r.client, subscription, r.time); err != nil {
			log.Error(err, "SetNotReadySubscription() failed for the subscription:", "subscription", subscription)
		}
		r.recorder.Eventf(subscription, corev1.EventTypeWarning, subReconcileFailed, "Subscription reconciliation failed: %v", err)
	} else if !requeue {
		// the work was done without errors
		if !subscription.ObjectMeta.DeletionTimestamp.IsZero() {
			// subscription is marked for deletion, all the work was done
			r.recorder.Eventf(subscription, corev1.EventTypeNormal, subReconciled,
				"Subscription reconciled and deleted, name: %q; namespace: %q", subscription.Name, subscription.Namespace)
		} else if util.IsSubscriptionActivated(subscription) {
			// reconcile finished with no errors
			if err := util.SetReadySubscription(ctx, r.client, subscription, "", r.time); err != nil {
				log.Error(err, "SetReadySubscription() failed for the subscription:", "subscription", subscription)
				reconcileErr = err
			} else {
				KymaSubscriptionsGauge.DeleteKymaSubscriptionsGaugeLabelValues(subscription.Namespace, subscription.Name)
				KymaSubscriptionsGauge.Metric.With(prometheus.Labels{
					metrics.Namespace: subscription.Namespace,
					metrics.Name:      subscription.Name,
					metrics.Ready:     "true"}).Set(1)
				log.Info("Subscription reconciled")
				r.recorder.Eventf(subscription, corev1.EventTypeNormal, subReconciled,
					"Subscription reconciled, name: %q; namespace: %q", subscription.Name, subscription.Namespace)
			}
		} else {
			// reconcile finished with no errors, but subscription is not activated
			if err := util.SetNotReadySubscription(ctx, r.client, subscription, r.time); err != nil {
				log.Error(err, "SetNotReadySubscription() failed for the subscription:", "subscription", subscription)
				reconcileErr = err
			} else {
				KymaSubscriptionsGauge.DeleteKymaSubscriptionsGaugeLabelValues(subscription.Namespace, subscription.Name)
				KymaSubscriptionsGauge.Metric.With(prometheus.Labels{metrics.Namespace: subscription.Namespace,
					metrics.Name:  subscription.Name,
					metrics.Ready: "false"}).Set(1)
				log.Info("Subscription reconciled")
				r.recorder.Eventf(subscription, corev1.EventTypeNormal, subReconciled,
					"Subscription reconciled, name: %q; namespace: %q", subscription.Name, subscription.Namespace)
			}
		}
	}

	return reconcile.Result{
		Requeue: requeue,
	}, reconcileErr
}

func (r *reconciler) reconcile(ctx context.Context, subscription *eventingv1alpha1.Subscription,
	KymaSubscriptionsGauge *metrics.SubscriptionsGauge) (bool, error) {

	knativeSubscriptionsGauge := metrics.KnativeSubscriptionsGaugeObj
	knativeChannelGauge := metrics.KnativeChanelGaugeObj
	knativeSubsName := util.GetKnSubscriptionName(&subscription.Name, &subscription.Namespace)
	knativeSubsNamespace := util.GetDefaultChannelNamespace()
	knativeSubsURI := subscription.Endpoint
	timeout := r.opts.ChannelTimeout

	//Adding the event-metadata as channel labels
	knativeChannelLabels := make(map[string]string)
	knativeChannelLabels[util.SubscriptionSourceID] = subscription.SourceID
	knativeChannelLabels[util.SubscriptionEventType] = subscription.EventType
	knativeChannelLabels[util.SubscriptionEventTypeVersion] = subscription.EventTypeVersion

	if subscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !util.ContainsString(&subscription.ObjectMeta.Finalizers, finalizerName) {
			subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, finalizerName)
			err := util.WriteSubscription(context.Background(), r.client, subscription)
			if err == nil {
				return true, nil
			}
			return false, err
		}
	} else {
		// The object is being deleted
		KymaSubscriptionsGauge.DeleteKymaSubscriptionsGaugeLabelValues(subscription.Namespace, subscription.Name)
		if util.ContainsString(&subscription.ObjectMeta.Finalizers, finalizerName) {
			// our finalizer is present, so lets handle our external dependency
			if err := r.deleteExternalDependency(ctx, knativeSubsName, knativeChannelLabels, knativeSubsNamespace,
				subscription.Name, knativeSubscriptionsGauge, knativeChannelGauge); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return false, err
			}

			// remove our finalizer from the list and update it.
			subscription.ObjectMeta.Finalizers = util.RemoveString(&subscription.ObjectMeta.Finalizers, finalizerName)
			if err := util.WriteSubscription(context.Background(), r.client, subscription); err != nil {
				return false, err
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return false, nil
	}

	// Check if Kyma Subscription has events-activated condition.
	if subscription.HasCondition(eventingv1alpha1.SubscriptionCondition{Type: eventingv1alpha1.EventsActivated, Status: eventingv1alpha1.ConditionTrue}) {
		// Check if Knative Channel already exists, create if not.
		knativeChannel, err := r.knativeLib.GetChannelByLabels(knativeSubsNamespace, knativeChannelLabels)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if errors.IsNotFound(err) {

			knativeChannel, err = r.knativeLib.CreateChannel(subscription.SubscriptionSpec.EventType,
				knativeSubsNamespace, knativeChannelLabels, util.WaitForChannelWithTimeout(timeout))
			if err != nil {
				return false, err
			}
			log.Info("Knative Channel is created", "Channel", knativeChannel)
		}
		knativeChannelGauge.Metric.With(prometheus.Labels{metrics.Name: subscription.Name}).Set(1)
		log.Info("Knative channel gauge is created:", "subscription", subscription.Name)

		// Check if Knative Subscription already exists, if not create one.
		sub, err := r.knativeLib.GetSubscription(knativeSubsName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if errors.IsNotFound(err) {
			knativeChannelLabels[util.SubNs] = subscription.Namespace
			err = r.knativeLib.CreateSubscription(knativeSubsName, knativeSubsNamespace, knativeChannel.Name, &knativeSubsURI, knativeChannelLabels)
			if err != nil {
				return false, err
			}
			knativeSubscriptionsGauge.DeleteKnativeSubscriptionsGaugeLabelValues(subscription.Namespace, subscription.Name)
			knativeSubscriptionsGauge.Metric.With(prometheus.Labels{
				metrics.Namespace: subscription.Namespace,
				metrics.Name:      subscription.Name,
				metrics.Ready:     "true"}).Set(1)
			log.Info("Knative Subscription is created", "Subscription", knativeSubsName)
		} else {
			// In case there is a change in Channel name or URI, delete and re-create Knative Subscription because update does not work.
			if knativeChannel.Name != sub.Spec.Channel.Name || knativeSubsURI != *sub.Spec.Subscriber.URI {
				err = r.knativeLib.DeleteSubscription(knativeSubsName, knativeSubsNamespace)
				if err != nil {
					return false, err
				}
				log.Info("Knative Subscription is deleted", "Subscription", knativeSubsName)
				err = r.knativeLib.CreateSubscription(knativeSubsName, knativeSubsNamespace, knativeChannel.Name, &knativeSubsURI, knativeChannelLabels)
				if err != nil {
					return false, err
				}
				knativeSubscriptionsGauge.DeleteKnativeSubscriptionsGaugeLabelValues(subscription.Namespace, subscription.Name)
				knativeSubscriptionsGauge.Metric.With(prometheus.Labels{
					metrics.Namespace: subscription.Namespace,
					metrics.Name:      subscription.Name,
					metrics.Ready:     "false"}).Set(1)
				log.Info("Knative Subscription is re-created", "Subscription", knativeSubsName)
			}
		}
	} else if util.CheckIfEventActivationExistForSubscription(ctx, r.client, subscription) {
		// In case Kyma Subscription does not have events-activated condition, but there is an EventActivation for it.
		// Activate subscription
		if err := util.ActivateSubscriptions(ctx, r.client, []*eventingv1alpha1.Subscription{subscription}, log, r.time); err != nil {
			log.Error(err, "ActivateSubscriptions() failed")
			return false, err
		}
		log.Info("Kyma Subscription is activated", "Subscription", subscription.Name)
		return true, nil
	} else {
		// In case Kyma Subscription does not have events-activated condition and there is no EventActivation, delete Knative Subscription & Channel if exist.
		knativeSub, err := r.knativeLib.GetSubscription(knativeSubsName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if err == nil && knativeSub != nil {
			err = r.knativeLib.DeleteSubscription(knativeSubsName, knativeSubsNamespace)
			if err != nil {
				return false, err
			}
			knativeSubscriptionsGauge.DeleteKnativeSubscriptionsGaugeLabelValues(subscription.Namespace, subscription.Name)
			log.Info("Knative Subscription is deleted", "Subscription", knativeSubsName)
			knativeChannelGauge.DeleteKnativeChannelGaugeLabelValues(subscription.Name)
			log.Info("Knative Channel Gauge is deleted", "subscription", subscription.Name)
		}

		// Check if Channel has any other Subscription, if not, delete it.
		knativeChannel, err := r.knativeLib.GetChannelByLabels(knativeSubsNamespace, knativeChannelLabels)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if err == nil && knativeChannel != nil {
			if knativeChannel.Spec.Subscribable == nil || len(knativeChannel.Spec.Subscribable.Subscribers) == 0 ||
				(len(knativeChannel.Spec.Subscribable.Subscribers) == 1 && knativeChannel.Spec.Subscribable.Subscribers[0].SubscriberURI == subscription.Endpoint) {
				err = r.knativeLib.DeleteChannel(knativeChannel.Name, knativeSubsNamespace)
				if err != nil {
					return false, err
				}
				log.Info("Knative Channel is deleted", "Channel", knativeChannel)
				knativeChannelGauge.DeleteKnativeChannelGaugeLabelValues(subscription.Name)
				log.Info("Knative Channel Gauge is deleted", "subscription", subscription.Name)
			}
		}
	}

	return false, nil
}

func (r *reconciler) deleteExternalDependency(ctx context.Context, knativeSubsName string, channelLabels map[string]string,
	namespace string, kymaSubscriptionName string, knativeSubscriptionsGauge *metrics.SubscriptionsGauge,
	knativeChannelGauge *metrics.SubscriptionsGauge) error {
	log.Info("Deleting the external dependencies")

	// In case Knative Subscription exists, delete it.
	knativeSubs, err := r.knativeLib.GetSubscription(knativeSubsName, namespace)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		err = r.knativeLib.DeleteSubscription(knativeSubs.Name, knativeSubs.Namespace)
		if err != nil {
			return err
		}
		knativeSubscriptionsGauge.DeleteKnativeSubscriptionsGaugeLabelValues(namespace, kymaSubscriptionName)
		log.Info("Knative Subscription is deleted", "Subscription", knativeSubs.Name)
	}

	// Check if Channel has any other Subscription, if not, delete it.
	knativeChannel, err := r.knativeLib.GetChannelByLabels(namespace, channelLabels)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		if knativeChannel.Spec.Subscribable == nil || (len(knativeChannel.Spec.Subscribable.Subscribers) == 1 && knativeSubs != nil &&
			knativeChannel.Spec.Subscribable.Subscribers[0].SubscriberURI == *knativeSubs.Spec.Subscriber.URI) {
			err = r.knativeLib.DeleteChannel(knativeChannel.Name, namespace)
			if err != nil {
				return err
			}
			log.Info("Knative Channel is deleted", "Channel", knativeChannel)
			knativeChannelGauge.DeleteKnativeChannelGaugeLabelValues(kymaSubscriptionName)
			log.Info("Knative Channel gauge is deleted", "subscription", kymaSubscriptionName)
		}
	}
	return nil
}
