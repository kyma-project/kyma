package subscription_nats

import (
	"context"
	"encoding/json"

	"fmt"

	"reflect"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	corev1 "k8s.io/api/core/v1"

	ceclient "github.com/cloudevents/sdk-go/v2/client"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler reconciles a Subscription object
type Reconciler struct {
	client.Client
	cache.Cache
	natsClient    *handlers.Nats
	Log           logr.Logger
	recorder      record.EventRecorder
	config        env.NatsConfig
	subscriptions map[string]*nats.Subscription
}

var (
	Finalizer = eventingv1alpha1.GroupVersion.Group
)

func NewReconciler(client client.Client, cache cache.Cache, log logr.Logger, recorder record.EventRecorder,
	cfg env.NatsConfig) *Reconciler {
	natsClient := &handlers.Nats{
		Log: log,
	}
	natsClient.Initialize(cfg)
	return &Reconciler{
		Client:        client,
		Cache:         cache,
		natsClient:    natsClient,
		Log:           log,
		recorder:      recorder,
		config:        cfg,
		subscriptions: make(map[string]*nats.Subscription),
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

	// examine DeletionTimestamp to determine if object is under deletion
	if !desiredSubscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if utils.ContainsString(desiredSubscription.ObjectMeta.Finalizers, Finalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteSubscriptions(req); err != nil {
				log.Info("failed to delete the external dependency of the subscription object")
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			log.Info("Removing finalizer from subscription object")
			desiredSubscription.ObjectMeta.Finalizers = utils.RemoveString(desiredSubscription.ObjectMeta.Finalizers,
				Finalizer)
			if err := r.Client.Update(context.Background(), desiredSubscription); err != nil {
				log.Info("failed to remove finalizer from subscription object")
				return ctrl.Result{}, err
			}
		}
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

	var filters []*eventingv1alpha1.BebFilter
	if desiredSubscription.Spec.Filter != nil {
		filters = desiredSubscription.Spec.Filter.Filters
	}

	// Clean up the old subscriptions
	err = r.deleteSubscriptions(req)
	if err != nil {
		log.Error(err, "failed to delete subscriptions")
		err = r.syncSubscriptionStatus(ctx, actualSubscription, false)
		return ctrl.Result{}, err
	}

	// Create subscriptions in Nats
	for _, filter := range filters {
		eventType := filter.EventType.Value
		callback := func(msg *nats.Msg) {
			log.Info(string(msg.Data), "payload", "value")
			ce, _ := r.convertMsgToCE(msg)
			ctx := cloudevents.ContextWithTarget(context.Background(), desiredSubscription.Spec.Sink)
			cev2Client, _ := ceclient.NewDefault()
			err := cev2Client.Send(ctx, *ce)
			if err != nil {
				log.Error(err, "failed to send event")
			}
		}
		sub, err := r.natsClient.Client.Subscribe(eventType, callback)
		if err != nil {
			log.Error(err, "failed to create a Nats subscription")
			// TODO retry once to create a new connection
			err = r.syncSubscriptionStatus(ctx, actualSubscription, false)
			if err != nil {
				log.Error(err, "failed to update subscription status")
			}
			return result, err
		}
		r.subscriptions[fmt.Sprintf("%s.%s", req.NamespacedName.String(), filter.EventType.Value)] = sub
	}

	// Update status
	err = r.syncSubscriptionStatus(ctx, actualSubscription, true)
	if err != nil {
		log.Error(err, "failed to update subscription status")
	}

	return result, nil
}

func (r Reconciler) syncSubscriptionStatus(ctx context.Context, sub *eventingv1alpha1.Subscription,
	isNatsSubReady bool) error {
	desiredSubscription := sub.DeepCopy()
	desiredConditions := make([]eventingv1alpha1.Condition, 0)
	conditionAdded := false
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonNATSSubscriptionActive, corev1.ConditionFalse)
	if isNatsSubReady {
		condition = eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
			eventingv1alpha1.ConditionReasonNATSSubscriptionActive, corev1.ConditionTrue)
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
			return err
		}
		r.Log.Info("successfully updated subscription", "namespace/name", fmt.Sprintf("%s/%s", sub.Namespace, sub.Name))
	}
	return nil
}

func (r Reconciler) deleteSubscriptions(req ctrl.Request) error {
	for k, v := range r.subscriptions {
		if strings.HasPrefix(k, req.NamespacedName.String()) {
			err := v.Unsubscribe()
			if err != nil {
				r.Log.Error(err, "failed to unsubscribe from NATS", "Subscription name: ", k)
				return err
			}
			delete(r.subscriptions, k)
			r.Log.Info("successfully deleted subscription", "namespace/name",
				req.NamespacedName.String())
		}
	}
	return nil
}

func (r Reconciler) convertMsgToCE(msg *nats.Msg) (*cev2event.Event, error) {
	event := cev2event.New(cev2event.CloudEventsVersionV1)
	err := json.Unmarshal(msg.Data, &event)
	if err != nil {
		return nil, err
	}
	r.Log.Info("cloud event received", "type", event.Type(), "id", event.ID(), "time", event.Time(),
		"data", string(event.Data()))
	return &event, nil
}
