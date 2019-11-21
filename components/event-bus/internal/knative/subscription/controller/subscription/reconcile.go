package subscription

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"knative.dev/eventing/pkg/logging"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/controller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	"github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/typed/eventing/v1alpha1"
	applicationconnectorlistersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/lister/applicationconnector/v1alpha1"
	subscriptionlistersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/lister/eventing/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/metrics"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

const (
	// Name of the corev1.Events emitted from the reconciliation process
	subReconciled      = "SubscriptionReconciled"
	subReconcileFailed = "SubscriptionReconcileFailed"

	// Finalizer for deleting Knative Subscriptions
	finalizerName = "subscription.finalizers.kyma-project.io"
)

//Reconciler Kyma subscriptions reconciler
type Reconciler struct {
	*reconciler.Base
	subscriptionLister    subscriptionlistersv1alpha1.SubscriptionLister
	eventActivationLister applicationconnectorlistersv1alpha1.EventActivationLister
	kymaEventingClient    eventingv1alpha1.EventingV1alpha1Interface

	knativeLib util.KnativeAccessLib
	opts       *opts.Options
	time       util.CurrentTime
}

//Reconcile reconciles a Kyma Subscription
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	subscription, err := subscriptionByKey(key, r.subscriptionLister)
	if err != nil {
		return err
	}
	log := logging.FromContext(ctx)
	// Fetch the Subscription instance
	//subscription := &v1alpha1.Subscription{}
	KymaSubscriptionsGauge := metrics.KymaSubscriptionsGaugeObj

	subscription = subscription.DeepCopy()

	requeue, reconcileErr := r.reconcile(ctx, subscription, KymaSubscriptionsGauge)
	if reconcileErr != nil {
		//log.Error(reconcileErr, "Reconciling Subscription")
		if err := util.SetNotReadySubscription(r.kymaEventingClient, subscription, r.time); err != nil {
			log.Error("SetNotReadySubscription() failed for the subscription:", zap.String("subscription", subscription.Name), zap.Error(err))
		}
		r.Recorder.Eventf(subscription, corev1.EventTypeWarning, subReconcileFailed, "Subscription reconciliation failed: %v", err)
	} else if !requeue {
		// the work was done without errors
		if !subscription.ObjectMeta.DeletionTimestamp.IsZero() {
			// subscription is marked for deletion, all the work was done
			r.Recorder.Eventf(subscription, corev1.EventTypeNormal, subReconciled,
				"Subscription reconciled and deleted, name: %q; namespace: %q", subscription.Name, subscription.Namespace)
		} else if util.IsSubscriptionActivated(subscription) {
			// reconcile finished with no errors
			if err := util.SetReadySubscription(r.kymaEventingClient, subscription, "", r.time); err != nil {
				log.Error("SetReadySubscription() failed for the subscription:", zap.String("subscription", subscription.Name), zap.Error(err))
				reconcileErr = err
			} else {
				KymaSubscriptionsGauge.DeleteKymaSubscriptionsGaugeLabelValues(subscription.Namespace, subscription.Name)
				KymaSubscriptionsGauge.Metric.With(prometheus.Labels{
					metrics.Namespace: subscription.Namespace,
					metrics.Name:      subscription.Name,
					metrics.Ready:     "true"}).Set(1)
				log.Info("Subscription reconciled")
				r.Recorder.Eventf(subscription, corev1.EventTypeNormal, subReconciled,
					"Subscription reconciled, name: %q; namespace: %q", subscription.Name, subscription.Namespace)
			}
		} else {
			// reconcile finished with no errors, but subscription is not activated
			if err := util.SetNotReadySubscription(r.kymaEventingClient, subscription, r.time); err != nil {
				log.Error("SetNotReadySubscription() failed for the subscription:", zap.String("subscription", subscription.Name), zap.Error(err))
				reconcileErr = err
			} else {
				KymaSubscriptionsGauge.DeleteKymaSubscriptionsGaugeLabelValues(subscription.Namespace, subscription.Name)
				KymaSubscriptionsGauge.Metric.With(prometheus.Labels{metrics.Namespace: subscription.Namespace,
					metrics.Name:  subscription.Name,
					metrics.Ready: "false"}).Set(1)
				log.Info("Subscription reconciled")
				r.Recorder.Eventf(subscription, corev1.EventTypeNormal, subReconciled,
					"Subscription reconciled, name: %q; namespace: %q", subscription.Name, subscription.Namespace)
			}
		}
	}

	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, subscription *v1alpha1.Subscription,
	KymaSubscriptionsGauge *metrics.SubscriptionsGauge) (bool, error) {
	log := logging.FromContext(ctx)

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
			err := util.WriteSubscription(r.kymaEventingClient, subscription)
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
			if err := util.WriteSubscription(r.kymaEventingClient, subscription); err != nil {
				return false, err
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return false, nil
	}

	// Check if Kyma Subscription has events-activated condition.
	if subscription.HasCondition(v1alpha1.SubscriptionCondition{Type: v1alpha1.EventsActivated, Status: v1alpha1.ConditionTrue}) {
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
			log.Info("Knative Channel is created", zap.String("Channel", knativeChannel.Name))
		}
		knativeChannelGauge.Metric.With(prometheus.Labels{metrics.Name: subscription.Name}).Set(1)
		log.Info("Knative channel gauge is created:", zap.String("subscription", subscription.Name))

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
			log.Info("Knative Subscription is created", zap.String("Subscription", knativeSubsName))
		} else {
			// In case there is a change in Channel name or URI, delete and re-create Knative Subscription because update does not work.
			if knativeChannel.Name != sub.Spec.Channel.Name || knativeSubsURI != sub.Spec.Subscriber.URI.String() {
				err = r.knativeLib.DeleteSubscription(knativeSubsName, knativeSubsNamespace)
				if err != nil {
					return false, err
				}
				log.Info("Knative Subscription is deleted", zap.String("Subscription", knativeSubsName))
				err = r.knativeLib.CreateSubscription(knativeSubsName, knativeSubsNamespace, knativeChannel.Name, &knativeSubsURI, knativeChannelLabels)
				if err != nil {
					return false, err
				}
				knativeSubscriptionsGauge.DeleteKnativeSubscriptionsGaugeLabelValues(subscription.Namespace, subscription.Name)
				knativeSubscriptionsGauge.Metric.With(prometheus.Labels{
					metrics.Namespace: subscription.Namespace,
					metrics.Name:      subscription.Name,
					metrics.Ready:     "false"}).Set(1)
				log.Info("Knative Subscription is re-created", zap.String("Subscription", knativeSubsName))
			}
		}
	} else if util.CheckIfEventActivationExistForSubscription(r.eventActivationLister, subscription) {
		// In case Kyma Subscription does not have events-activated condition, but there is an EventActivation for it.
		// Activate subscription
		if err := util.ActivateSubscriptions(r.kymaEventingClient, []*v1alpha1.Subscription{subscription}, log, r.time); err != nil {
			log.Error("ActivateSubscriptions() failed", zap.Error(err))
			return false, err
		}
		log.Info("Kyma Subscription is activated", zap.String("Subscription", subscription.Name))
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
			log.Info("Knative Subscription is deleted", zap.String("Subscription", knativeSubsName))
			knativeChannelGauge.DeleteKnativeChannelGaugeLabelValues(subscription.Name)
			log.Info("Knative Channel Gauge is deleted", zap.String("subscription", subscription.Name))
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
				log.Info("Knative Channel is deleted", zap.String("Channel", knativeChannel.Name))
				knativeChannelGauge.DeleteKnativeChannelGaugeLabelValues(subscription.Name)
				log.Info("Knative Channel Gauge is deleted", zap.String("subscription", subscription.Name))
			}
		}
	}

	return false, nil
}

func (r *Reconciler) deleteExternalDependency(ctx context.Context, knativeSubsName string, channelLabels map[string]string,
	namespace string, kymaSubscriptionName string, knativeSubscriptionsGauge *metrics.SubscriptionsGauge,
	knativeChannelGauge *metrics.SubscriptionsGauge) error {
	log := logging.FromContext(ctx)
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
		log.Info("Knative Subscription is deleted", zap.String("Subscription", knativeSubs.Name))
	}

	// Check if Channel has any other Subscription, if not, delete it.
	knativeChannel, err := r.knativeLib.GetChannelByLabels(namespace, channelLabels)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		if knativeChannel.Spec.Subscribable == nil || (len(knativeChannel.Spec.Subscribable.Subscribers) == 1 && knativeSubs != nil &&
			knativeChannel.Spec.Subscribable.Subscribers[0].SubscriberURI == knativeSubs.Spec.Subscriber.URI.String()) {
			err = r.knativeLib.DeleteChannel(knativeChannel.Name, namespace)
			if err != nil {
				return err
			}
			log.Info("Knative Channel is deleted", zap.String("Channel", knativeChannel.Name))
			knativeChannelGauge.DeleteKnativeChannelGaugeLabelValues(kymaSubscriptionName)
			log.Info("Knative Channel gauge is deleted", zap.String("subscription", kymaSubscriptionName))
		}
	}
	return nil
}

// subscriptionByKey retrieves a Subscription object from a lister by ns/name key.
func subscriptionByKey(key string, l subscriptionlistersv1alpha1.SubscriptionLister) (*v1alpha1.Subscription, error) {
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
