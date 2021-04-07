package subscription_nats

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

// Reconciler reconciles a Subscription object
type Reconciler struct {
	client.Client
	cache.Cache
	natsClient       *handlers.Nats
	Log              logr.Logger
	recorder         record.EventRecorder
	eventTypeCleaner eventtype.Cleaner
}

var (
	Finalizer = eventingv1alpha1.GroupVersion.Group
)

const (
	NATSProtocol = "NATS"
)

func NewReconciler(client client.Client, applicationLister *application.Lister, cache cache.Cache, log logr.Logger, recorder record.EventRecorder, cfg env.NatsConfig) *Reconciler {
	natsClient := handlers.NewNats(cfg, log)
	err := natsClient.Initialize()
	if err != nil {
		log.Error(err, "reconciler can't start")
		panic(err)
	}
	return &Reconciler{
		Client:           client,
		Cache:            cache,
		natsClient:       natsClient,
		Log:              log,
		recorder:         recorder,
		eventTypeCleaner: eventtype.NewCleaner(cfg.EventTypePrefix, applicationLister, log),
	}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Subscription{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	r.Log.Info("received subscription reconciliation request", "namespace", req.Namespace, "name",
		req.Name)

	actualSubscription := &eventingv1alpha1.Subscription{}
	result := ctrl.Result{}

	// Ensure the object was not deleted in the meantime
	err := r.Client.Get(ctx, req.NamespacedName, actualSubscription)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle only the new subscription
	desiredSubscription := actualSubscription.DeepCopy()
	//Bind fields to logger
	log := r.Log.WithValues("kind", desiredSubscription.GetObjectKind().GroupVersionKind().Kind,
		"name", desiredSubscription.GetName(),
		"namespace", desiredSubscription.GetNamespace(),
		"version", desiredSubscription.GetGeneration(),
	)

	if !desiredSubscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if utils.ContainsString(desiredSubscription.ObjectMeta.Finalizers, Finalizer) {
			if err := r.natsClient.DeleteSubscription(desiredSubscription); err != nil {
				r.Log.Error(err, "failed to delete subscription")
				// if failed to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			log.Info("Removing finalizer from subscription object")
			desiredSubscription.ObjectMeta.Finalizers = utils.RemoveString(desiredSubscription.ObjectMeta.Finalizers,
				Finalizer)
			if err := r.Client.Update(context.Background(), desiredSubscription); err != nil {
				log.Error(err, "failed to remove finalizer from subscription object")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}
	// Check for valid sink
	if err := r.assertSinkValidity(actualSubscription.Spec.Sink); err != nil {
		r.Log.Error(err, "failed to parse sink URL")
		if err := r.syncSubscriptionStatus(ctx, actualSubscription, false, err.Error()); err != nil {
			return ctrl.Result{}, err
		}
		// No point in reconciling as the sink is invalid
		return ctrl.Result{}, nil
	}

	// Clean up the old subscriptions
	err = r.natsClient.DeleteSubscription(desiredSubscription)
	if err != nil {
		log.Error(err, "failed to delete subscriptions")
		if err := r.syncSubscriptionStatus(ctx, actualSubscription, false, err.Error()); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object. This is equivalent
	// registering our finalizer.
	if !utils.ContainsString(desiredSubscription.ObjectMeta.Finalizers, Finalizer) {
		log.Info("Adding finalizer to subscription object")
		desiredSubscription.ObjectMeta.Finalizers = append(desiredSubscription.ObjectMeta.Finalizers, Finalizer)
		if err := r.Update(context.Background(), desiredSubscription); err != nil {
			return ctrl.Result{}, err
		}
		result.Requeue = true
	}

	if result.Requeue {
		return result, nil
	}

	err = r.natsClient.SyncSubscription(desiredSubscription, r.eventTypeCleaner)
	if err != nil {
		r.Log.Error(err, "failed to sync subscription")
		if err := r.syncSubscriptionStatus(ctx, actualSubscription, false, err.Error()); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	log.Info("successfully created Nats subscriptions")

	// Update status
	if err := r.syncSubscriptionStatus(ctx, actualSubscription, true, ""); err != nil {
		return ctrl.Result{}, err
	}

	return result, nil
}

// syncSubscriptionStatus syncs Subscription status
func (r Reconciler) syncSubscriptionStatus(ctx context.Context, sub *eventingv1alpha1.Subscription,
	isNatsSubReady bool, message string) error {
	desiredSubscription := sub.DeepCopy()
	desiredConditions := make([]eventingv1alpha1.Condition, 0)
	conditionAdded := false
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive, corev1.ConditionFalse, message)
	if isNatsSubReady {
		condition = eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
			eventingv1alpha1.ConditionReasonNATSSubscriptionActive, corev1.ConditionTrue, "")
	}
	for _, c := range sub.Status.Conditions {
		var chosenCondition eventingv1alpha1.Condition
		if c.Type == condition.Type {
			// take given condition
			chosenCondition = condition
			conditionAdded = true
		} else {
			// take already present condition
			chosenCondition = c
		}
		desiredConditions = append(desiredConditions, chosenCondition)
	}
	if !conditionAdded {
		desiredConditions = append(desiredConditions, condition)
	}
	desiredSubscription.Status.Conditions = desiredConditions
	desiredSubscription.Status.Ready = isNatsSubReady

	if !reflect.DeepEqual(sub.Status, desiredSubscription.Status) {
		err := r.Client.Status().Update(ctx, desiredSubscription, &client.UpdateOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to update subscription status")
		}
		r.Log.Info("successfully updated subscription status")
	}
	return nil
}

func (r Reconciler) assertSinkValidity(sink string) error {
	_, err := url.ParseRequestURI(sink)
	return err
}

func (r Reconciler) assertProtocolValidity(protocol string) error {
	if protocol != NATSProtocol {
		return fmt.Errorf("invalid protocol: %s", protocol)
	}
	return nil
}
