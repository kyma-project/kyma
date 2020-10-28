package controllers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// SubscriptionReconciler reconciles a Subscription object
type SubscriptionReconciler struct {
	client.Client
	cache.Cache
	Log       logr.Logger
	recorder  record.EventRecorder
	Scheme    *runtime.Scheme
	bebClient *handlers.Beb
}

var (
	FinalizerName = eventingv1alpha1.GroupVersion.Group
)

const (
	SinkURLPrefix = "webhook"
	// TODO: Think about the collisions when using length 6
	SuffixLength                 = 10
	ClusterLocalAPIGateway       = "kyma-gateway.kyma-system.svc.cluster.local"
	ControllerServiceLabelKey    = "service"
	ControllerIdentityLabelKey   = "beb"
	ControllerIdentityLabelValue = "webhook"
	ExternalHostPrefix           = "web"
	ClusterLocalURLSuffix        = "svc.cluster.local"
)

func NewSubscriptionReconciler(
	client client.Client,
	cache cache.Cache,
	log logr.Logger,
	recorder record.EventRecorder,
	scheme *runtime.Scheme,
) *SubscriptionReconciler {
	bebClient := &handlers.Beb{
		Log: log,
	}
	return &SubscriptionReconciler{
		Client:    client,
		Cache:     cache,
		Log:       log,
		recorder:  recorder,
		Scheme:    scheme,
		bebClient: bebClient,
	}
}

// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch

// Generate required RBAC to emit kubernetes events in the controller
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// Source: https://book-v1.book.kubebuilder.io/beyond_basics/creating_events.html

// +kubebuilder:printcolumn:name="Ready",type=bool,JSONPath=`.status.Ready`
// Source: https://book.kubebuilder.io/reference/generating-crd.html#additional-printer-columns

// TODO: Optimize number of reconciliation calls in eventing-controller #9766: https://github.com/kyma-project/kyma/issues/9766
func (r *SubscriptionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("subscription", req.NamespacedName)

	cachedSubscription := &eventingv1alpha1.Subscription{}

	result := ctrl.Result{}

	// Ensure the object was not deleted in the meantime
	if err := r.Client.Get(ctx, req.NamespacedName, cachedSubscription); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Handle only the new subscription
	subscription := cachedSubscription.DeepCopy()

	// Bind fields to logger
	log := r.Log.WithValues("kind", subscription.GetObjectKind().GroupVersionKind().Kind,
		"name", subscription.GetName(),
		"namespace", subscription.GetNamespace(),
		"version", subscription.GetGeneration(),
	)

	if !r.isInDeletion(subscription) {
		// Ensure the finalizer is set
		if err := r.syncFinalizer(subscription, &result, ctx, log); err != nil {
			log.Error(err, "error while syncing finalizer")
			return ctrl.Result{}, err
		}
		if result.Requeue {
			return result, nil
		}
		if err := r.syncInitialStatus(subscription, &result, ctx); err != nil {
			log.Error(err, "error while syncing status")
			return ctrl.Result{}, err
		}
		if result.Requeue {
			return result, nil
		}
	}

	// mark if the subscription status was changed
	statusChanged := false

	// Sync with APIRule, expose the webhook
	if statusChangedForAPIRule, err := r.syncAPIRule(subscription, &result, ctx, log); err != nil {
		log.Error(err, "error while syncing API rule")
		return ctrl.Result{}, err
	} else {
		statusChanged = statusChanged || statusChangedForAPIRule
	}

	// Sync the BEB Subscription with the Subscription CR
	//if statusChangedForBeb, err := r.syncBEBSubscription(subscription, &result, ctx, log); err != nil {
	//	log.Error(err, "error while syncing BEB subscription")
	//	return ctrl.Result{}, err
	//} else {
	//	statusChanged = statusChanged || statusChangedForBeb
	//}

	if r.isInDeletion(subscription) {
		// Remove finalizers
		if err := r.removeFinalizer(subscription, ctx, log); err != nil {
			return ctrl.Result{}, err
		}
		result.Requeue = false
		return result, nil
	}

	// Save the subscription status if it was changed
	if statusChanged {
		if err := r.Status().Update(ctx, subscription); err != nil {
			log.Error(err, "Update subscription status failed")
			return ctrl.Result{}, err
		}
		result.Requeue = true
	}

	return result, nil
}

// syncFinalizer sets the finalizer in the Subscription
func (r *SubscriptionReconciler) syncFinalizer(subscription *eventingv1alpha1.Subscription, result *ctrl.Result, ctx context.Context, logger logr.Logger) error {
	// Check if finalizer is already set
	if r.isFinalizerSet(subscription) {
		return nil
	}
	if err := r.addFinalizer(subscription, ctx, logger); err != nil {
		return err
	}
	result.Requeue = true
	return nil
}

// syncBEBSubscription delegates the subscription synchronization to the backend client. It returns true if the subscription sattus was changed.
func (r *SubscriptionReconciler) syncBEBSubscription(subscription *eventingv1alpha1.Subscription,
	result *ctrl.Result, ctx context.Context, logger logr.Logger) (bool, error) {
	logger.Info("Syncing subscription with BEB")

	r.bebClient.Initialize()

	// if object is marked for deletion, we need to delete the BEB subscription
	if r.isInDeletion(subscription) {
		return false, r.deleteBEBSubscription(subscription, logger, ctx)
	}

	var statusChanged bool
	var err error
	if statusChanged, err = r.bebClient.SyncBebSubscription(subscription); err != nil {
		logger.Error(err, "Update BEB subscription failed")
		condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreationFailed, corev1.ConditionFalse)
		if err := r.updateCondition(subscription, condition, ctx); err != nil {
			return statusChanged, err
		}
		return false, err
	}

	if !subscription.Status.IsConditionSubscribed() {
		condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue)
		if err := r.updateCondition(subscription, condition, ctx); err != nil {
			return statusChanged, err
		}
		statusChanged = true
	}

	statusChangedAtCheck, retry, errTimeout := r.checkStatusActive(subscription)
	statusChanged = statusChanged || statusChangedAtCheck
	if errTimeout != nil {
		logger.Error(errTimeout, "Timeout at retry")
		result.Requeue = false
		return statusChanged, errTimeout
	}
	if retry {
		logger.Info("Wait for subscription to be active", "name:", subscription.Name, "status:", subscription.Status.EmsSubscriptionStatus.SubscriptionStatus)
		condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionNotActive, corev1.ConditionFalse)
		if err := r.updateCondition(subscription, condition, ctx); err != nil {
			return statusChanged, err
		}
		result.RequeueAfter = time.Second * 1
	} else if statusChanged {
		condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionActive, corev1.ConditionTrue)
		if err := r.updateCondition(subscription, condition, ctx); err != nil {
			return statusChanged, err
		}
	}
	// OK
	return statusChanged, nil
}

// deleteBEBSubscription deletes the BEB subscription and updates the condition and k8s events
func (r *SubscriptionReconciler) deleteBEBSubscription(subscription *eventingv1alpha1.Subscription, logger logr.Logger, ctx context.Context) error {
	logger.Info("Deleting BEB subscription")
	if err := r.bebClient.DeleteBebSubscription(subscription); err != nil {
		return err
	}
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionDeleted, corev1.ConditionFalse)
	return r.updateCondition(subscription, condition, ctx)
}

func (r *SubscriptionReconciler) syncAPIRule(subscription *eventingv1alpha1.Subscription, result *ctrl.Result,
	ctx context.Context, logger logr.Logger) (bool, error) {
	var statusChanged bool
	// Validate correctness of a URL
	sURL, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		logger.Error(err, "url is invalid")
		return statusChanged, nil
	}

	// Validate svcNs and svcName from sink URL
	svcNs, svcName, err := getSvcNsAndName(sURL.Host)
	if err != nil {
		logger.Error(err, "failed to parse svcName and svcNamespace")
		return statusChanged, nil
	}

	// Assumption: Subscription CR and Subscriber should be deployed in the same namespace
	if subscription.Namespace != svcNs {
		logger.Error(fmt.Errorf("stopping reconciliation as the namespace of Subscription: %s and the namespace of subscriber: %s are different", subscription.Namespace, svcNs), "")
		return statusChanged, nil
	}

	// Validate sink URL is a cluster local URL
	trimmedHost := strings.Split(sURL.Host, ":")[0]
	if !strings.HasSuffix(trimmedHost, ClusterLocalURLSuffix) {
		logger.Error(fmt.Errorf("sink does not contain %s URL", ClusterLocalURLSuffix), "")
		return statusChanged, nil
	}
	// Validate svc is a cluster local one
	_, err = r.validateClusterLocalService(ctx, svcNs, svcName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Error(err, "sink doesn't correspond to a valid cluster local svc")
		}
		return statusChanged, errors.Wrap(err, "failed to get the svc")
	}

	err = r.createOrUpdateAPIRule(*sURL, subscription, ctx, logger)
	if err != nil {
		return statusChanged, errors.Wrap(err, "failed to createOrUpdateAPIRule")
	}
	return statusChanged, nil
}

func (r *SubscriptionReconciler) validateClusterLocalService(ctx context.Context, svcNs, svcName string) (*corev1.Service, error) {
	svcLookupKey := k8stypes.NamespacedName{Name: svcName, Namespace: svcNs}
	svc := &corev1.Service{}
	if err := r.Cache.Get(ctx, svcLookupKey, svc); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, errors.Wrap(err, "sink does not correspond to a valid k8s svc")
		}
		return nil, err
	}
	return svc, nil
}

func (r *SubscriptionReconciler) createOrUpdateAPIRule(sink url.URL, subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	svcNs, svcName, err := getSvcNsAndName(sink.Host)
	if err != nil {
		return errors.Wrap(err, "failed to parse svc name and ns in createOrUpdateAPIRule")
	}
	labels := map[string]string{
		ControllerServiceLabelKey:  svcName,
		ControllerIdentityLabelKey: ControllerIdentityLabelValue,
	}

	svcPort, err := convertURLPortForApiRulePort(sink)
	if err != nil {
		logger.Error(err, "error while converting URL port to APIRule port")
		return nil
	}
	existingAPIRules, err := r.getAPIRulesForASvc(ctx, labels, svcNs)
	if err != nil {
		logger.Error(err, "error while fetching oldApiRule for labels", labels)
		return nil
	}
	var existingAPIRule *apigatewayv1alpha1.APIRule
	if existingAPIRules != nil {
		existingAPIRule = r.filterAPIRulesOnPort(existingAPIRules, svcPort)
	}

	// Get all subscriptions valid for the cluster-local subscriber
	subscriptions, err := r.getSubscriptionsForASvc(svcNs, svcName, ctx)
	if err != nil {
		logger.Error(err, "failed to fetch subscriptions for the subscriber is focus")
		return nil
	}
	filteredSubscriptions := r.filterSubscriptionsOnPort(subscriptions, svcPort)

	desiredAPIRule, err := r.makeAPIRule(svcNs, svcName, labels, filteredSubscriptions, svcPort)
	if err != nil {
		return errors.Wrap(err, "failed to make an APIRule")
	}
	if existingAPIRule == nil {
		err = r.Client.Create(ctx, desiredAPIRule, &client.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create APIRule")
		}
		return nil
	}
	logger.Info("Existing APIRules", fmt.Sprintf("in ns: %s for svc: %s", svcNs, svcName), fmt.Sprintf("%v", *existingAPIRule))
	// Assumption: there will be one APIRule for an svc with the labels injected by the controller hence trusting the 0th element in existingAPIRules list
	object.ApplyExistingAPIRuleAttributes(existingAPIRule, desiredAPIRule)
	if object.Semantic.DeepEqual(existingAPIRule, desiredAPIRule) {
		return nil
	}
	// Update the existing APIRule
	err = r.Client.Update(ctx, desiredAPIRule, &client.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update an APIRule")
	}
	// TODO: Cleanup APIRules if need be
	//err = r.cleanupAPIRulesByPort(subscriptions, existingAPIRules)
	//if err != nil {
	//	return errors.Wrap(err, "failed to cleanup APIRules")
	//}
	return nil
}

func (r *SubscriptionReconciler) cleanupAPIRulesByPort(subs []eventingv1alpha1.Subscription, apiRules []apigatewayv1alpha1.APIRule) error {
	//apiRulesForPortExcept := apiRulesForPortExcept(apiRules, svcPort)
	cleanupMap := make(map[*apigatewayv1alpha1.APIRule][]eventingv1alpha1.Subscription)
	for _, apiRule := range apiRules {
		tempSubArr := make([]eventingv1alpha1.Subscription, 0)
		ownerRefs := apiRule.OwnerReferences
		for _, or := range ownerRefs {
			for _, sub := range subs {
				if or.Name == sub.Name && or.UID == sub.UID {

					subSinkURL, err := url.ParseRequestURI(sub.Spec.Sink)
					if err != nil {
						// It's ok as this subscription doesn't have a port anyway
						continue
					}
					port, err := convertURLPortForApiRulePort(*subSinkURL)
					if err != nil {
						// It's ok as the port is not valid anyway
						continue
					}
					if port == *apiRule.Spec.Service.Port {
						tempSubArr = append(tempSubArr, sub)
					}
				}

			}
		}
		cleanupMap[&apiRule] = tempSubArr
		// Update the apiRule based on the result
	}
	return nil
}

func apiRulesForPortExcept(apiRules []apigatewayv1alpha1.APIRule, svcPort uint32) []apigatewayv1alpha1.APIRule {
	filteredAPIRules := make([]apigatewayv1alpha1.APIRule, 0)
	for _, apiRule := range apiRules {
		if *apiRule.Spec.Service.Port != svcPort {
			filteredAPIRules = append(filteredAPIRules, apiRule)
		}
	}
	return filteredAPIRules
}

// getRelevantSubscriptions returns a list of Subscriptions which are valid for the subscriber in focus
func (r *SubscriptionReconciler) getSubscriptionsForASvc(svcNs, svcName string, ctx context.Context) ([]eventingv1alpha1.Subscription, error) {
	subscriptions := &eventingv1alpha1.SubscriptionList{}
	relevantSubs := make([]eventingv1alpha1.Subscription, 0)
	err := r.Cache.List(ctx, subscriptions, &client.ListOptions{
		Namespace: svcNs,
	})
	if err != nil {
		return []eventingv1alpha1.Subscription{}, err
	}
	for _, sub := range subscriptions.Items {
		// Filtering subscriptions which are being deleted at the moment
		if sub.DeletionTimestamp != nil {
			continue
		}
		hostURL, err := url.ParseRequestURI(sub.Spec.Sink)
		if err != nil {
			// It's ok as the relevant subscription will have a valid cluster local URL in the same namespace
			continue
		}
		// Filtering subscriptions valid for a valid subscriber
		svcNsForSub, svcNameForSub, err := getSvcNsAndName(hostURL.Host)
		if err != nil {
			// It's ok as the relevant subscription will have a valid cluster local URL in the same namespace
			continue
		}
		//svcPortForSub, err := convertURLPortForApiRulePort(*hostURL)
		if svcNs == svcNsForSub && svcName == svcNameForSub {
			relevantSubs = append(relevantSubs, sub)
		}
	}
	return relevantSubs, nil
}

// filterSubscriptionsOnPort returns a list of Subscriptions which matches a particular port
func (r *SubscriptionReconciler) filterSubscriptionsOnPort(subList []eventingv1alpha1.Subscription, svcPort uint32) []eventingv1alpha1.Subscription {
	filteredSubs := make([]eventingv1alpha1.Subscription, 0)
	for _, sub := range subList {
		// Filtering subscriptions which are being deleted at the moment
		if sub.DeletionTimestamp != nil {
			continue
		}
		hostURL, err := url.ParseRequestURI(sub.Spec.Sink)
		if err != nil {
			// It's ok as the relevant subscription will have a valid cluster local URL in the same namespace
			continue
		}

		svcPortForSub, err := convertURLPortForApiRulePort(*hostURL)
		if err != nil {
			// It's ok as the relevant subscription will have a valid port to filter on
			continue
		}
		if svcPort == svcPortForSub {
			filteredSubs = append(filteredSubs, sub)
		}
	}
	return filteredSubs
}

func (r *SubscriptionReconciler) makeAPIRule(svcNs, svcName string, labels map[string]string, subs []eventingv1alpha1.Subscription, port uint32) (*apigatewayv1alpha1.APIRule, error) {

	randomSuffix := handlers.GetRandSuffix(SuffixLength)
	hostName := fmt.Sprintf("%s-%s", ExternalHostPrefix, randomSuffix)

	apiRule := object.NewAPIRule(svcNs, SinkURLPrefix,
		object.WithLabels(labels),
		object.WithOwnerReference(subs),
		object.WithService(hostName, svcName, port),
		object.WithGateway(ClusterLocalAPIGateway),
		object.WithRules(subs, http.MethodPost, http.MethodOptions))
	return apiRule, nil
}

func convertURLPortForApiRulePort(sink url.URL) (uint32, error) {
	port := uint32(0)
	sinkPort := sink.Port()
	if sinkPort != "" {
		u64, err := strconv.ParseUint(sink.Port(), 10, 32)
		if err != nil {
			return port, errors.Wrapf(err, "failed to convert port: %s", sink.Port())
		}
		port = uint32(u64)
	}
	if port == uint32(0) {
		switch strings.ToLower(sink.Scheme) {
		case "http":
			port = uint32(80)
		case "https":
			port = uint32(443)
		}
	}
	return port, nil
}

func (r *SubscriptionReconciler) getAPIRulesForASvc(ctx context.Context, labels map[string]string, svcNs string) ([]apigatewayv1alpha1.APIRule, error) {
	existingAPIRules := &apigatewayv1alpha1.APIRuleList{}
	err := r.Cache.List(ctx, existingAPIRules, &client.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(labels),
		Namespace:     svcNs,
	})
	if err != nil {
		return nil, err
	}
	return existingAPIRules.Items, nil
}

func (r *SubscriptionReconciler) filterAPIRulesOnPort(existingAPIRules []apigatewayv1alpha1.APIRule, port uint32) *apigatewayv1alpha1.APIRule {
	for _, apiRule := range existingAPIRules {
		if *apiRule.Spec.Service.Port == port {
			return &apiRule
		}
	}
	return nil
}

// getSvcNsAndName returns namespace and name of the svc from the URL
func getSvcNsAndName(url string) (string, string, error) {
	parts := strings.Split(url, ".")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid sinkURL for cluster local svc: %s", url)
	}
	return parts[1], parts[0], nil
}

// syncInitialStatus determines the desires initial status and updates it accordingly (if conditions changed)
func (r *SubscriptionReconciler) syncInitialStatus(subscription *eventingv1alpha1.Subscription, result *ctrl.Result, ctx context.Context) error {
	currentStatus := subscription.Status

	expectedStatus := eventingv1alpha1.SubscriptionStatus{}
	expectedStatus.InitializeConditions()

	// case: conditions are already initialized
	if len(currentStatus.Conditions) >= len(expectedStatus.Conditions) {
		return nil
	}

	subscription.Status = expectedStatus
	if err := r.Status().Update(ctx, subscription); err != nil {
		return err
	}
	result.Requeue = true

	return nil
}

// updateCondition replaces the given condition on the subscription and updates the status as well as emitting a kubernetes event
func (r *SubscriptionReconciler) updateCondition(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition, ctx context.Context) error {
	needsUpdate, err := r.replaceStatusCondition(subscription, condition)
	if err != nil {
		return err
	}
	if !needsUpdate {
		return nil
	}

	if err := r.Status().Update(ctx, subscription); err != nil {
		return err
	}

	r.emitConditionEvent(subscription, condition)
	return nil
}

// replaceStatusCondition replaces the given condition on the subscription. Also it sets the readyness in the status.
// So make sure you always use this method then changing a condition
func (r *SubscriptionReconciler) replaceStatusCondition(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition) (bool, error) {
	// the subscription is ready if all conditions are fulfilled
	isReady := true

	// compile list of desired conditions
	desiredConditions := make([]eventingv1alpha1.Condition, 0)
	for _, c := range subscription.Status.Conditions {
		var chosenCondition eventingv1alpha1.Condition
		if c.Type == condition.Type {
			// take given condition
			chosenCondition = condition
		} else {
			// take already present condition
			chosenCondition = c
		}
		desiredConditions = append(desiredConditions, chosenCondition)
		if string(chosenCondition.Status) != string(v1.ConditionTrue) {
			isReady = false
		}
	}

	// prevent unnecessary updates
	if conditionsEquals(subscription.Status.Conditions, desiredConditions) && subscription.Status.Ready == isReady {
		return false, nil
	}

	// update the status
	subscription.Status.Conditions = desiredConditions
	subscription.Status.Ready = isReady
	return true, nil
}

// emitConditionEvent emits a kubernetes event and sets the event type based on the Condition status
func (r *SubscriptionReconciler) emitConditionEvent(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition) {
	eventType := corev1.EventTypeNormal
	if condition.Status == corev1.ConditionFalse {
		eventType = corev1.EventTypeWarning
	}
	r.recorder.Event(subscription, eventType, string(condition.Reason), condition.Message)
}

func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Subscription{}).
		Complete(r)
}

func (r *SubscriptionReconciler) addFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, FinalizerName)
	logger.V(1).Info("Adding finalizer")
	if err := r.Update(ctx, subscription); err != nil {
		return errors.Wrapf(err, "error while adding Finalizer with name: %s", FinalizerName)
	}
	logger.V(1).Info("Added finalizer")
	return nil
}

func (r *SubscriptionReconciler) removeFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	var finalizers []string

	// Build finalizer list without the one the controller owns
	for _, finalizer := range subscription.ObjectMeta.Finalizers {
		if finalizer == FinalizerName {
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	logger.V(1).Info("Removing finalizer")
	subscription.ObjectMeta.Finalizers = finalizers
	if err := r.Update(ctx, subscription); err != nil {
		return errors.Wrapf(err, "error while removing Finalizer with name: %s", FinalizerName)
	}
	logger.V(1).Info("Removed finalizer")
	return nil
}

// isFinalizerSet checks if a finalizer is set on the Subscription which belongs to this controller
func (r *SubscriptionReconciler) isFinalizerSet(subscription *eventingv1alpha1.Subscription) bool {
	// Check if finalizer is already set
	for _, finalizer := range subscription.ObjectMeta.Finalizers {
		if finalizer == FinalizerName {
			return true
		}
	}
	return false
}

// isInDeletion checks if the Subscription shall be deleted
func (r *SubscriptionReconciler) isInDeletion(subscription *eventingv1alpha1.Subscription) bool {
	return !subscription.DeletionTimestamp.IsZero()
}

const timeoutRetryActiveEmsStatus = time.Second * 30

// checkStatusActive checks if the subscription is active and if not, sets a timer for retry
func (r *SubscriptionReconciler) checkStatusActive(subscription *eventingv1alpha1.Subscription) (statusChanged, retry bool, err error) {
	if subscription.Status.EmsSubscriptionStatus.SubscriptionStatus == string(types.SubscriptionStatusActive) {
		if len(subscription.Status.FailedActivation) > 0 {
			subscription.Status.FailedActivation = ""
			return true, false, nil
		}
		return false, false, nil
	}
	t1 := time.Now()
	if len(subscription.Status.FailedActivation) == 0 {
		// it's the first time
		subscription.Status.FailedActivation = t1.Format(time.RFC3339)
		return true, true, nil
	}
	// check the timeout
	if t0, er := time.Parse(time.RFC3339, subscription.Status.FailedActivation); er != nil {
		err = er
	} else if t1.Sub(t0) > timeoutRetryActiveEmsStatus {
		err = fmt.Errorf("timeout waiting for the subscription to be active: %v", subscription.Name)
	} else {
		retry = true
	}
	return false, retry, err
}
