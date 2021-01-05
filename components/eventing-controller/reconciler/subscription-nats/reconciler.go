package subscription_nats

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"

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

const (
	NATSProtocol = "NATS"
)

func NewReconciler(client client.Client, cache cache.Cache, log logr.Logger, recorder record.EventRecorder,
	cfg env.NatsConfig) *Reconciler {
	natsClient := &handlers.Nats{
		Log: log,
	}
	err := natsClient.Initialize(cfg)
	if err != nil {
		log.Error(err, "reconciler can't start")
		panic(err)
	}
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

	if !desiredSubscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if utils.ContainsString(desiredSubscription.ObjectMeta.Finalizers, Finalizer) {
			if err := r.deleteSubscriptions(req); err != nil {
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

	// Clean up the old subscriptions
	err = r.deleteSubscriptions(req)
	if err != nil {
		log.Error(err, "failed to delete subscriptions")
		err = r.syncSubscriptionStatus(ctx, actualSubscription, false, err.Error())
		return ctrl.Result{}, err
	}

	// Check for valid protocol
	actualProtocol := actualSubscription.Spec.Protocol
	if err := r.assertProtocolValidity(actualProtocol); err != nil {
		r.Log.Error(err, "invalid subscription protocol")
		err := r.syncSubscriptionStatus(ctx, actualSubscription, false, err.Error())
		if err != nil {
			r.Log.Error(err, "failed to sync subscription status")
			return ctrl.Result{}, err
		}
		// No point in reconciling as the sink is invalid
		return ctrl.Result{}, nil
	}

	// Check for valid sink
	if err := r.assertSinkValidity(actualSubscription.Spec.Sink); err != nil {
		r.Log.Error(err, "failed to parse sink URL")
		err := r.syncSubscriptionStatus(ctx, actualSubscription, false, err.Error())
		if err != nil {
			r.Log.Error(err, "failed to sync subscription status")
			return ctrl.Result{}, err
		}
		// No point in reconciling as the sink is invalid
		return ctrl.Result{}, nil
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

	// Create subscriptions in Nats
	for _, filter := range filters {
		eventType := filter.EventType.Value
		callback := func(msg *nats.Msg) {
			ce, err := convertMsgToCE(msg)
			if err != nil {
				r.Log.Error(err, "failed to convert Nats message to CE")
				return
			}
			ctx := cloudevents.ContextWithTarget(context.Background(), desiredSubscription.Spec.Sink)
			cev2Client, err := ceclient.NewDefault()
			if err != nil {
				r.Log.Error(err, "failed to create CE client")
				return
			}
			err = cev2Client.Send(ctx, *ce)
			if err != nil {
				log.Error(err, "failed to dispatch event")
				return
			}
		}

		if r.natsClient.Connection.Status() != nats.CONNECTED {
			r.Log.Info("connection to Nats", "status", fmt.Sprintf("%v", r.natsClient.Connection.Status()))
			initializeErr := r.natsClient.Initialize(r.config)
			if initializeErr != nil {
				log.Error(initializeErr, "failed to reset NATS connection")
				return result, initializeErr
			}
		}

		sub, subscribeErr := r.natsClient.Connection.Subscribe(eventType, callback)
		if subscribeErr != nil {
			log.Error(subscribeErr, "failed to create a Nats subscription")
			syncErr := r.syncSubscriptionStatus(ctx, actualSubscription, false, subscribeErr.Error())
			if syncErr != nil {
				log.Error(syncErr, "failed to update subscription status")
				return result, syncErr
			}
			return result, subscribeErr
		}
		r.subscriptions[fmt.Sprintf("%s.%s", req.NamespacedName.String(), filter.EventType.Value)] = sub
	}
	log.Info("successfully created Nats subscriptions")

	// Update status
	err = r.syncSubscriptionStatus(ctx, actualSubscription, true, "")
	if err != nil {
		log.Error(err, "failed to update subscription status")
	}
	log.Info("successfully updated subscription status")

	return result, nil
}

// syncSubscriptionStatus syncs Subscription status
func (r Reconciler) syncSubscriptionStatus(ctx context.Context, sub *eventingv1alpha1.Subscription,
	isNatsSubReady bool, message string) error {
	desiredSubscription := sub.DeepCopy()
	desiredConditions := make([]eventingv1alpha1.Condition, 0)
	conditionAdded := false
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonNATSSubscriptionActive, corev1.ConditionFalse, message)
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
			return errors.Wrapf(err, "failed to update Subscription status")
		}
		r.Log.Info("successfully updated subscription")
	}
	return nil
}

func (r Reconciler) deleteSubscriptions(req ctrl.Request) error {
	for k, v := range r.subscriptions {
		if strings.HasPrefix(k, req.NamespacedName.String()) {
			r.Log.Info("connection status", "status", r.natsClient.Connection.Status())
			// Unsubscribe call to Nats is async hence checking the status of the connection is important
			if r.natsClient.Connection.Status() != nats.CONNECTED {
				initializeErr := r.natsClient.Initialize(r.config)
				if initializeErr != nil {
					return errors.Wrapf(initializeErr, "can't connect to Nats server")
				}
			}
			err := v.Unsubscribe()
			if err != nil {
				return errors.Wrapf(err, "failed to unsubscribe")
			}
			delete(r.subscriptions, k)
			r.Log.Info("successfully unsubscribed/deleted subscription")
		}
	}
	return nil
}

func convertMsgToCE(msg *nats.Msg) (*cev2event.Event, error) {
	event := cev2event.New(cev2event.CloudEventsVersionV1)
	err := json.Unmarshal(msg.Data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
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
