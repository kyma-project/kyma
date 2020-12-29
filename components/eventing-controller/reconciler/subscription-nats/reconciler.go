package subscription_nats

import (
	"context"
	"encoding/json"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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
	if desiredSubscription.ObjectMeta.DeletionTimestamp.IsZero() {
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

		err := r.deleteSubscriptions(req)
		if err != nil {
			log.Error(err, "failed to delete subscriptions")
			return ctrl.Result{}, err
		}

		for _, filter := range filters {
			eventType := filter.EventType.Value
			sub, err := r.natsClient.Client.Subscribe(eventType, func(msg *nats.Msg) {
				log.Info(string(msg.Data), "payload", "value")
				ce, _ := r.convertMsgToCE(msg)
				ctx := cloudevents.ContextWithTarget(context.Background(), desiredSubscription.Spec.Sink)
				cev2Client, _ := ceclient.NewDefault()

				err := cev2Client.Send(ctx, *ce)
				if err != nil {
					log.Error(err, "failed to send event")
				}
			})
			r.subscriptions[req.NamespacedName.String()+"."+filter.EventType.Value] = sub
			if err != nil {
				log.Error(err, "failed to create a Nats subscription")
				return result, err
			}
		}

	} else {
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
			if err := r.Update(context.Background(), desiredSubscription); err != nil {
				log.Info("failed to remove finalizer from subscription object")
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	return result, nil
}

func (r Reconciler) deleteSubscriptions(req ctrl.Request) error {
	for k, v := range r.subscriptions {
		if strings.HasPrefix(k, req.NamespacedName.String()) {
			err := v.Unsubscribe()
			if err != nil {
				r.Log.Error(err, "Couldn't unsubscribe from NATS", "Subscription name: ", k)
				return err
			}
			delete(r.subscriptions, k)
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
	r.Log.Info("cloud event received", "type", event.Type(), "id", event.ID(), "time", event.Time(), "data", string(event.Data()))
	//evTime, err := time.Parse(time.RFC3339, msg.Header["ce-time"][0])
	//if err != nil {
	//	return nil, errors.Wrap(err, "failed to parse time")
	//}
	//event.SetTime(evTime)
	//
	//if err := event.SetData("application/json", msg.Data); err != nil {
	//	return nil, errors.Wrap(err, "failed to set data to CE data field")
	//}
	//
	//event.SetID(msg.Header["ce-id"][0])
	//
	//eventType := msg.Header["ce-type"][0]
	//event.SetType(eventType)
	//event.SetDataContentType("application/json")
	return &event, nil
}

// isInDeletion checks if the Subscription shall be deleted
func isInDeletion(subscription *eventingv1alpha1.Subscription) bool {
	return !subscription.DeletionTimestamp.IsZero()
}

// syncInitialStatus determines the desires initial status and updates it accordingly (if conditions changed)
func (r *Reconciler) syncInitialStatus(subscription *eventingv1alpha1.Subscription, result *ctrl.Result, ctx context.Context) error {
	currentStatus := subscription.Status
	expectedStatus := eventingv1alpha1.SubscriptionStatus{}
	expectedStatus.InitializeConditions()
	currentReadyStatusFromConditions := currentStatus.IsReady()

	var updateReadyStatus bool
	if currentReadyStatusFromConditions && !currentStatus.Ready {
		currentStatus.Ready = true
		updateReadyStatus = true
	} else if !currentReadyStatusFromConditions && currentStatus.Ready {
		currentStatus.Ready = false
		updateReadyStatus = true
	}
	// case: conditions are already initialized
	if len(currentStatus.Conditions) >= len(expectedStatus.Conditions) && !updateReadyStatus {
		return nil
	}
	if len(currentStatus.Conditions) == 0 {
		subscription.Status = expectedStatus
	} else {
		subscription.Status.Ready = currentStatus.Ready
	}
	if err := r.Status().Update(ctx, subscription); err != nil {
		return err
	}
	result.Requeue = true
	return nil
}
