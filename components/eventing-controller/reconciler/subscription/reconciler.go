package subscription

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	. "github.com/kyma-project/kyma/components/eventing-controller/reconciler/errors"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

// Reconciler reconciles a Subscription object
type Reconciler struct {
	ctx context.Context
	client.Client
	cache.Cache
	Log              logr.Logger
	recorder         record.EventRecorder
	Backend          handlers.MessagingBackend
	Domain           string
	eventTypeCleaner eventtype.Cleaner
}

var (
	Finalizer = eventingv1alpha1.GroupVersion.Group
)

const (
	suffixLength          = 10
	externalHostPrefix    = "web"
	externalSinkScheme    = "https"
	apiRuleNamePrefix     = "webhook-"
	clusterLocalURLSuffix = "svc.cluster.local"
)

func NewReconciler(ctx context.Context, client client.Client, applicationLister *application.Lister, cache cache.Cache,
	log logr.Logger, recorder record.EventRecorder, cfg env.Config) *Reconciler {
	bebHandler := &handlers.Beb{Log: log}
	bebHandler.Initialize(cfg)

	return &Reconciler{
		ctx:              ctx,
		Client:           client,
		Cache:            cache,
		Log:              log,
		recorder:         recorder,
		Backend:          bebHandler,
		Domain:           cfg.Domain,
		eventTypeCleaner: eventtype.NewCleaner(cfg.EventTypePrefix, applicationLister, log),
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
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = r.Log.WithValues("subscription", req.NamespacedName)

	actualSubscription := &eventingv1alpha1.Subscription{}

	result := ctrl.Result{}

	// Ensure the object was not deleted in the meantime
	if err := r.Client.Get(ctx, req.NamespacedName, actualSubscription); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Handle only the new subscription
	desiredSubscription := actualSubscription.DeepCopy()

	// Bind fields to logger
	log := r.Log.WithValues("kind", desiredSubscription.GetObjectKind().GroupVersionKind().Kind,
		"name", desiredSubscription.GetName(),
		"namespace", desiredSubscription.GetNamespace(),
		"version", desiredSubscription.GetGeneration(),
	)

	// the APIRule for the desired subscription
	var apiRule *apigatewayv1alpha1.APIRule

	if !isInDeletion(desiredSubscription) {
		// ensure the finalizer is set
		if err := r.syncFinalizer(desiredSubscription, &result, ctx, log); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to sync finalizer")
		}
		if result.Requeue {
			return result, nil
		}

		// sync the initial Subscription status
		if err := r.syncInitialStatus(desiredSubscription, &result, ctx); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to sync status")
		}
		if result.Requeue {
			return result, nil
		}

		// sync APIRule
		var err error
		apiRule, err = r.syncAPIRule(desiredSubscription, ctx, log)
		if !IsSkippable(err) {
			return ctrl.Result{}, err
		}

		// sync the Subscription status for the APIRule
		statusChanged, err := r.syncSubscriptionAPIRuleStatus(actualSubscription, desiredSubscription, apiRule)
		if err != nil {
			return ctrl.Result{}, err
		}
		if statusChanged {
			if err := r.Client.Status().Update(ctx, desiredSubscription); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// mark if the subscription status was changed
	statusChanged := false

	// Sync the BEB Subscription with the Subscription CR
	if statusChangedForBeb, err := r.syncBEBSubscription(desiredSubscription, &result, ctx, log, apiRule); err != nil {
		log.Error(err, "error while syncing BEB subscription")
		return ctrl.Result{}, err
	} else {
		statusChanged = statusChanged || statusChangedForBeb
	}

	if isInDeletion(desiredSubscription) {
		// Remove finalizers
		if err := r.removeFinalizer(desiredSubscription, ctx, log); err != nil {
			return ctrl.Result{}, err
		}
		result.Requeue = false
		return result, nil
	}

	// Save the subscription status if it was changed
	if statusChanged {
		if err := r.Status().Update(ctx, desiredSubscription); err != nil {
			log.Error(err, "Update subscription status failed")
			return ctrl.Result{}, err
		}
		result.Requeue = true
	}

	return result, nil
}

func (r *Reconciler) syncSubscriptionAPIRuleStatus(actualSubscription, desiredSubscription *eventingv1alpha1.Subscription, apiRule *apigatewayv1alpha1.APIRule) (bool, error) {
	apiRuleReady := computeAPIRuleReadyStatus(apiRule)

	// set the default subscription status
	desiredSubscription.Status.APIRuleName = ""
	desiredSubscription.Status.ExternalSink = ""
	desiredSubscription.Status.SetConditionAPIRuleStatus(apiRuleReady)

	if apiRule != nil {
		desiredSubscription.Status.APIRuleName = apiRule.Name
	}

	// set subscription sink only if the APIRule is ready
	if apiRuleReady {
		if err := setSubscriptionStatusExternalSink(desiredSubscription, apiRule); err != nil {
			return false, errors.Wrapf(err, "Failed to set Subscription status externalSink [%s/%s]", desiredSubscription.Namespace, desiredSubscription.Name)
		}
	}

	return isApiRuleStatueChanged(actualSubscription, desiredSubscription), nil
}

// syncFinalizer sets the finalizer in the Subscription
func (r *Reconciler) syncFinalizer(subscription *eventingv1alpha1.Subscription, result *ctrl.Result, ctx context.Context, logger logr.Logger) error {
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

// syncBEBSubscription delegates the subscription synchronization to the backend client. It returns true if the subscription status was changed.
func (r *Reconciler) syncBEBSubscription(subscription *eventingv1alpha1.Subscription, result *ctrl.Result,
	ctx context.Context, logger logr.Logger, apiRule *apigatewayv1alpha1.APIRule) (bool, error) {
	logger.Info("Syncing subscription with BEB")

	// if object is marked for deletion, we need to delete the BEB subscription
	if isInDeletion(subscription) {
		return false, r.deleteBEBSubscription(subscription, logger, ctx)
	}

	if apiRule == nil {
		return false, errors.Errorf("APIRule is required")
	}

	var statusChanged bool
	var err error
	if statusChanged, err = r.Backend.SyncSubscription(subscription, r.eventTypeCleaner, apiRule); err != nil {
		logger.Error(err, "Update BEB subscription failed")
		condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreationFailed, corev1.ConditionFalse, "")
		if err := r.updateCondition(subscription, condition, ctx); err != nil {
			return statusChanged, err
		}
		return false, err
	}

	if !subscription.Status.IsConditionSubscribed() {
		condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue, "")
		if err := r.updateCondition(subscription, condition, ctx); err != nil {
			return statusChanged, err
		}
		statusChanged = true
	}

	statusChangedAtCheck, retry, errTimeout := r.checkStatusActive(subscription)
	statusChanged = statusChanged || statusChangedAtCheck
	if errTimeout != nil {
		logger.Error(errTimeout, "timeout at retry")
		result.Requeue = false
		return statusChanged, errTimeout
	}
	if retry {
		logger.Info("Wait for subscription to be active", "name:", subscription.Name, "status:", subscription.Status.EmsSubscriptionStatus.SubscriptionStatus)
		condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionNotActive, corev1.ConditionFalse, "")
		if err := r.updateCondition(subscription, condition, ctx); err != nil {
			return statusChanged, err
		}
		result.RequeueAfter = time.Second * 1
	} else if statusChanged {
		condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionActive, corev1.ConditionTrue, "")
		if err := r.updateCondition(subscription, condition, ctx); err != nil {
			return statusChanged, err
		}
	}
	// OK
	return statusChanged, nil
}

// deleteBEBSubscription deletes the BEB subscription and updates the condition and k8s events
func (r *Reconciler) deleteBEBSubscription(subscription *eventingv1alpha1.Subscription, logger logr.Logger, ctx context.Context) error {
	logger.Info("Deleting BEB subscription")
	if err := r.Backend.DeleteSubscription(subscription); err != nil {
		return err
	}
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionDeleted, corev1.ConditionFalse, "")
	return r.updateCondition(subscription, condition, ctx)
}

// syncAPIRule validate the given subscription sink URL and sync its APIRule.
func (r *Reconciler) syncAPIRule(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) (*apigatewayv1alpha1.APIRule, error) {
	if err := r.isSinkURLValid(ctx, subscription); err != nil {
		return nil, err
	}

	sURL, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		r.eventWarn(subscription, reasonValidationFailed, "Failed to parse sink URI %s", subscription.Spec.Sink)
		return nil, NewSkippable(errors.Wrapf(err, "failed to parse sink URI"))
	}

	apiRule, err := r.createOrUpdateAPIRule(subscription, ctx, *sURL, logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create or update APIRule")
	}

	return apiRule, nil
}

func (r *Reconciler) isSinkURLValid(ctx context.Context, subscription *eventingv1alpha1.Subscription) error {
	if !isValidScheme(subscription.Spec.Sink) {
		r.eventWarn(subscription, reasonValidationFailed, "Sink URL scheme should be 'http' or 'https' %s", subscription.Spec.Sink)
		return NewSkippable(fmt.Errorf("sink URL scheme should be 'http' or 'https'"))
	}

	sURL, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		r.eventWarn(subscription, reasonValidationFailed, "Sink URL is not valid %s", err.Error())
		return NewSkippable(err)
	}

	// Validate sink URL is a cluster local URL
	trimmedHost := strings.Split(sURL.Host, ":")[0]
	if !strings.HasSuffix(trimmedHost, clusterLocalURLSuffix) {
		r.eventWarn(subscription, reasonValidationFailed, "sink does not contain suffix: %s in the URL", clusterLocalURLSuffix)
		return NewSkippable(fmt.Errorf("sink does not contain suffix: %s in the URL", clusterLocalURLSuffix))
	}

	// we expected a sink in the format "service.namespace.svc.cluster.local"
	subDomains := strings.Split(trimmedHost, ".")
	if len(subDomains) != 5 {
		r.eventWarn(subscription, reasonValidationFailed, "sink should contain 5 sub-domains %s", trimmedHost)
		return NewSkippable(fmt.Errorf("sink should contain 5 sub-domains: %s", trimmedHost))
	}

	// Assumption: Subscription CR and Subscriber should be deployed in the same namespace
	svcNs := subDomains[1]
	if subscription.Namespace != svcNs {
		r.eventWarn(subscription, reasonValidationFailed, "the namespace of Subscription: %s and the namespace of subscriber: %s are different", subscription.Namespace, svcNs)
		return NewSkippable(fmt.Errorf("the namespace of Subscription: %s and the namespace of subscriber: %s are different", subscription.Namespace, svcNs))
	}

	// Validate svc is a cluster-local one
	svcName := subDomains[0]
	if _, err := r.getClusterLocalService(ctx, svcNs, svcName); err != nil {
		if k8serrors.IsNotFound(err) {
			r.eventWarn(subscription, reasonValidationFailed, "sink doesn't correspond to a valid cluster local svc")
			return NewSkippable(errors.Wrapf(err, "sink doesn't correspond to a valid cluster local svc"))
		}

		r.eventWarn(subscription, reasonValidationFailed, "failed to fetch cluster-local svc %s/%s", svcNs, svcName)
		return errors.Wrapf(err, "failed to fetch cluster-local svc %s/%s", svcNs, svcName)
	}

	return nil
}

func (r *Reconciler) getClusterLocalService(ctx context.Context, svcNs, svcName string) (*corev1.Service, error) {
	svcLookupKey := k8stypes.NamespacedName{Name: svcName, Namespace: svcNs}
	svc := &corev1.Service{}
	if err := r.Client.Get(ctx, svcLookupKey, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

// createOrUpdateAPIRule create new or update existing APIRule for the given subscription.
func (r *Reconciler) createOrUpdateAPIRule(subscription *eventingv1alpha1.Subscription, ctx context.Context, sink url.URL, logger logr.Logger) (*apigatewayv1alpha1.APIRule, error) {
	svcNs, svcName, err := getSvcNsAndName(sink.Host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse svc name and ns in createOrUpdateAPIRule")
	}
	labels := map[string]string{
		constants.ControllerServiceLabelKey:  svcName,
		constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
	}

	svcPort, err := utils.GetPortNumberFromURL(sink)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert URL port to APIRule port")
	}
	var reusableAPIRule *apigatewayv1alpha1.APIRule
	existingAPIRules, err := r.getAPIRulesForASvc(ctx, labels, svcNs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch existing ApiRule for labels: %v", labels)
	}
	if existingAPIRules != nil {
		reusableAPIRule = r.filterAPIRulesOnPort(existingAPIRules, svcPort)
	}

	// Get all subscriptions valid for the cluster-local subscriber
	subscriptions, err := r.getSubscriptionsForASvc(svcNs, svcName, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch subscriptions for the subscriber %s/%s", svcNs, svcName)
	}
	filteredSubscriptions := r.filterSubscriptionsOnPort(subscriptions, svcPort)

	desiredAPIRule, err := r.makeAPIRule(svcNs, svcName, labels, filteredSubscriptions, svcPort)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make an APIRule")
	}

	// update or remove the previous APIRule if it is not used by other subscriptions
	if err := r.handlePreviousAPIRule(subscription, reusableAPIRule, ctx); err != nil {
		return nil, err
	}

	// no APIRule to reuse, create a new one
	if reusableAPIRule == nil {
		if err := r.Client.Create(ctx, desiredAPIRule, &client.CreateOptions{}); err != nil {
			r.eventWarn(subscription, reasonCreateFailed, "Create APIRule failed %s", desiredAPIRule.Name)
			return nil, errors.Wrap(err, "failed to create APIRule")
		}

		r.eventNormal(subscription, reasonCreate, "Created APIRule %s", desiredAPIRule.Name)
		return desiredAPIRule, nil
	}
	logger.Info("Existing APIRules", fmt.Sprintf("in ns: %s for svc: %s", svcNs, svcName), fmt.Sprintf("%s", reusableAPIRule.Name))

	object.ApplyExistingAPIRuleAttributes(reusableAPIRule, desiredAPIRule)
	if object.Semantic.DeepEqual(reusableAPIRule, desiredAPIRule) {
		return reusableAPIRule, nil
	}
	err = r.Client.Update(ctx, desiredAPIRule, &client.UpdateOptions{})
	if err != nil {
		r.eventWarn(subscription, reasonUpdateFailed, "Update APIRule failed %s", desiredAPIRule.Name)
		return nil, errors.Wrap(err, "failed to update an APIRule")
	}
	r.eventNormal(subscription, reasonUpdate, "Updated APIRule %s", desiredAPIRule.Name)

	return desiredAPIRule, nil
}

// handlePreviousAPIRule computes the OwnerReferences list for the previous subscription APIRule (if any)
// if the OwnerReferences list is empty, then the APIRule will be deleted
// else if the OwnerReferences list length was decreased, then the APIRule will be updated
// TODO write more tests https://github.com/kyma-project/kyma/issues/9950
func (r *Reconciler) handlePreviousAPIRule(subscription *eventingv1alpha1.Subscription, reusableApiRule *apigatewayv1alpha1.APIRule, ctx context.Context) error {
	// subscription does not have a previous APIRule
	if len(subscription.Status.APIRuleName) == 0 {
		return nil
	}

	// the previous APIRule for the subscription is the current one no need to update it
	if reusableApiRule != nil && subscription.Status.APIRuleName == reusableApiRule.Name {
		return nil
	}

	// get the previous APIRule
	previousAPIRule := &apigatewayv1alpha1.APIRule{}
	key := k8stypes.NamespacedName{Namespace: subscription.Namespace, Name: subscription.Status.APIRuleName}
	if err := r.Client.Get(ctx, key, previousAPIRule); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	// build a new OwnerReference list and exclude the current subscription from the list (if exists)
	ownerReferences := make([]v1.OwnerReference, 0, len(previousAPIRule.OwnerReferences))
	for _, ownerReference := range previousAPIRule.OwnerReferences {
		if ownerReference.UID != subscription.UID {
			ownerReferences = append(ownerReferences, ownerReference)
		}
	}

	// delete the ApiRule if the new OwnerReference list is empty
	if len(ownerReferences) == 0 {
		if err := r.Client.Delete(ctx, previousAPIRule); err != nil {
			return err
		}
		return nil
	}

	// update the ApiRule if the new OwnerReference list length is decreased
	if len(ownerReferences) < len(previousAPIRule.OwnerReferences) {
		// list all subscriptions in the APIRule namespace
		namespaceSubscriptions := &eventingv1alpha1.SubscriptionList{}
		if err := r.Client.List(ctx, namespaceSubscriptions, &client.ListOptions{Namespace: previousAPIRule.Namespace}); err != nil {
			return err
		}

		// build a new subscription list and exclude the current subscription from the list
		subscriptions := make([]eventingv1alpha1.Subscription, 0, len(namespaceSubscriptions.Items))
		for _, namespaceSubscription := range namespaceSubscriptions.Items {
			// skip the current subscription
			if namespaceSubscription.UID == subscription.UID {
				continue
			}

			// skip not relevant subscriptions to the previous APIRule
			if namespaceSubscription.Status.APIRuleName != previousAPIRule.Name {
				continue
			}

			subscriptions = append(subscriptions, namespaceSubscription)
		}

		// update the APIRule OwnerReferences list and Spec Rules
		object.WithOwnerReference(subscriptions)(previousAPIRule)
		object.WithRules(subscriptions, http.MethodPost, http.MethodOptions)(previousAPIRule)

		if err := r.Client.Update(ctx, previousAPIRule); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) cleanup(subscription *eventingv1alpha1.Subscription, ctx context.Context, subs []eventingv1alpha1.Subscription, apiRules []apigatewayv1alpha1.APIRule) error {
	for _, apiRule := range apiRules {
		filteredOwnerRefs := make([]metav1.OwnerReference, 0)
		for _, or := range apiRule.OwnerReferences {
			for _, sub := range subs {
				if isOwnerRefBelongingToSubscription(sub, or) {
					subSinkURL, err := url.ParseRequestURI(sub.Spec.Sink)
					if err != nil {
						// It's ok as this subscription doesn't have a port anyway
						continue
					}
					port, err := utils.GetPortNumberFromURL(*subSinkURL)
					if err != nil {
						// It's ok as the port is not valid anyway
						continue
					}
					if port == *apiRule.Spec.Service.Port {
						filteredOwnerRefs = append(filteredOwnerRefs, or)
					}
				}
			}
		}

		// Delete the APIRule as the port for the concerned svc is not used by any subscriptions
		if len(filteredOwnerRefs) == 0 {
			err := r.Client.Delete(ctx, &apiRule, &client.DeleteOptions{})
			if err != nil {
				r.eventWarn(subscription, reasonDeleteFailed, "Deleted APIRule failed %s", apiRule.Name)
				return errors.Wrap(err, "failed to delete APIRule while cleanupAPIRules")
			}
			r.eventNormal(subscription, reasonDelete, "Deleted APIRule %s", apiRule.Name)
			return nil
		}

		// Take the subscription out of the OwnerReferences and update the APIRule
		desiredAPIRule := apiRule.DeepCopy()
		object.ApplyExistingAPIRuleAttributes(&apiRule, desiredAPIRule)
		desiredAPIRule.OwnerReferences = filteredOwnerRefs
		err := r.Client.Update(ctx, desiredAPIRule, &client.UpdateOptions{})
		if err != nil {
			r.eventWarn(subscription, reasonUpdateFailed, "Update APIRule failed %s", apiRule.Name)
			return errors.Wrap(err, "failed to update APIRule while cleanupAPIRules")
		}
		r.eventNormal(subscription, reasonUpdate, "Updated APIRule %s", apiRule.Name)
		return nil
	}
	return nil
}

func isOwnerRefBelongingToSubscription(sub eventingv1alpha1.Subscription, ownerRef metav1.OwnerReference) bool {
	if sub.Name == ownerRef.Name && sub.UID == ownerRef.UID {
		return true
	}
	return false
}

// getSubscriptionsForASvc returns a list of Subscriptions which are valid for the subscriber in focus
func (r *Reconciler) getSubscriptionsForASvc(svcNs, svcName string, ctx context.Context) ([]eventingv1alpha1.Subscription, error) {
	subscriptions := &eventingv1alpha1.SubscriptionList{}
	relevantSubs := make([]eventingv1alpha1.Subscription, 0)
	err := r.Client.List(ctx, subscriptions, &client.ListOptions{
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
func (r *Reconciler) filterSubscriptionsOnPort(subList []eventingv1alpha1.Subscription, svcPort uint32) []eventingv1alpha1.Subscription {
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

		svcPortForSub, err := utils.GetPortNumberFromURL(*hostURL)
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

func (r *Reconciler) makeAPIRule(svcNs, svcName string, labels map[string]string, subs []eventingv1alpha1.Subscription, port uint32) (*apigatewayv1alpha1.APIRule, error) {

	randomSuffix := handlers.GetRandString(suffixLength)
	hostName := fmt.Sprintf("%s-%s.%s", externalHostPrefix, randomSuffix, r.Domain)

	apiRule := object.NewAPIRule(svcNs, apiRuleNamePrefix,
		object.WithLabels(labels),
		object.WithOwnerReference(subs),
		object.WithService(hostName, svcName, port),
		object.WithGateway(constants.ClusterLocalAPIGateway),
		object.WithRules(subs, http.MethodPost, http.MethodOptions))
	return apiRule, nil
}

func (r *Reconciler) getAPIRulesForASvc(ctx context.Context, labels map[string]string, svcNs string) ([]apigatewayv1alpha1.APIRule, error) {
	existingAPIRules := &apigatewayv1alpha1.APIRuleList{}
	err := r.Client.List(ctx, existingAPIRules, &client.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(labels),
		Namespace:     svcNs,
	})
	if err != nil {
		return nil, err
	}
	return existingAPIRules.Items, nil
}

func (r *Reconciler) filterAPIRulesOnPort(existingAPIRules []apigatewayv1alpha1.APIRule, port uint32) *apigatewayv1alpha1.APIRule {
	// Assumption: there will be one APIRule for an svc with the labels injected by the controller hence trusting the first match
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

// updateCondition replaces the given condition on the subscription and updates the status as well as emitting a kubernetes event
func (r *Reconciler) updateCondition(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition, ctx context.Context) error {
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

// replaceStatusCondition replaces the given condition on the subscription. Also it sets the readiness in the status.
// So make sure you always use this method then changing a condition
func (r *Reconciler) replaceStatusCondition(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition) (bool, error) {
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
func (r *Reconciler) emitConditionEvent(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition) {
	eventType := corev1.EventTypeNormal
	if condition.Status == corev1.ConditionFalse {
		eventType = corev1.EventTypeWarning
	}
	r.recorder.Event(subscription, eventType, string(condition.Reason), condition.Message)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Subscription{}).
		Watches(&source.Kind{Type: &apigatewayv1alpha1.APIRule{}}, r.getAPIRuleEventHandler()).
		Complete(r)
}

//  SetupUnmanaged creates a controller under the client control
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged("beb-subscription-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		r.Log.Error(err, "failed to create a unmanaged BEB controller")
		return err
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha1.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		r.Log.Error(err, "unable to watch pods")
		return err
	}

	go func(r *Reconciler, c controller.Controller) {
		if err := c.Start(r.ctx); err != nil {
			r.Log.Error(err, "failed to start the beb-subscription-controller")
			os.Exit(1)
		}
	}(r, ctru)

	return nil
}

// getAPIRuleEventHandler returns an APIRule event handler.
func (r *Reconciler) getAPIRuleEventHandler() handler.EventHandler {
	eventHandler := func(eventType, name, namespace string, q workqueue.RateLimitingInterface) {
		log := r.Log.WithValues("event", eventType, "kind", "APIRule", "name", name, "namespace", namespace)
		if err := r.handleAPIRuleEvent(name, namespace, q, log); err != nil {
			log.Error(err, "Failed to handle APIRule Event, requeue event again")
			q.Add(reconcile.Request{NamespacedName: k8stypes.NamespacedName{Name: name, Namespace: namespace}})
		}
	}

	return handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			eventType, name, namespace := "Create", e.Object.GetName(), e.Object.GetNamespace()
			eventHandler(eventType, name, namespace, q)
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			eventType, name, namespace := "Update", e.ObjectNew.GetName(), e.ObjectNew.GetNamespace()
			eventHandler(eventType, name, namespace, q)
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			eventType, name, namespace := "Delete", e.Object.GetName(), e.Object.GetNamespace()
			eventHandler(eventType, name, namespace, q)
		},
		GenericFunc: func(e event.GenericEvent, q workqueue.RateLimitingInterface) {
			eventType, name, namespace := "Generic", e.Object.GetName(), e.Object.GetNamespace()
			eventHandler(eventType, name, namespace, q)
		},
	}
}

// handleAPIRuleEvent handles APIRule event.
func (r *Reconciler) handleAPIRuleEvent(name, namespace string, q workqueue.RateLimitingInterface, log logr.Logger) error {
	// skip not relevant APIRules
	if !isRelevantAPIRuleName(name) {
		return nil
	}

	log.Info("Handle APIRule Event")

	// try to get the APIRule from the API server
	ctx := context.Background()
	apiRule := &apigatewayv1alpha1.APIRule{}
	key := k8stypes.NamespacedName{Name: name, Namespace: namespace}
	if err := r.Client.Get(ctx, key, apiRule); err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	// list all namespace subscriptions
	namespaceSubscriptions := &eventingv1alpha1.SubscriptionList{}
	if err := r.Client.List(ctx, namespaceSubscriptions, client.InNamespace(namespace)); err != nil {
		log.Error(err, "Failed to list namespace Subscriptions")
		return err
	}

	// filter namespace subscriptions that are relevant to the current APIRule
	apiRuleSubscriptions := make([]eventingv1alpha1.Subscription, 0, len(apiRule.ObjectMeta.OwnerReferences))
	for _, subscription := range namespaceSubscriptions.Items {
		// skip if the subscription is marked for deletion
		if subscription.DeletionTimestamp != nil {
			continue
		}

		// check if APIRule name match
		if subscription.Status.APIRuleName == name {
			apiRuleSubscriptions = append(apiRuleSubscriptions, subscription)
			continue
		}

		// check if APIRule OwnerReferences contains subscription info
		if containsOwnerReference(apiRule.ObjectMeta.OwnerReferences, subscription.UID) {
			apiRuleSubscriptions = append(apiRuleSubscriptions, subscription)
			continue
		}
	}

	// queue reconcile requests for APIRule subscriptions
	r.queueReconcileRequestForSubscriptions(apiRuleSubscriptions, q, log)

	return nil
}

// containsOwnerReference returns true if the OwnerReferences list contains the given uid, otherwise returns false.
func containsOwnerReference(ownerReferences []v1.OwnerReference, uid k8stypes.UID) bool {
	for _, ownerReference := range ownerReferences {
		if ownerReference.UID == uid {
			return true
		}
	}
	return false
}

// queueReconcileRequestForSubscriptions queues reconciliation requests for the given subscriptions.
func (r *Reconciler) queueReconcileRequestForSubscriptions(subscriptions []eventingv1alpha1.Subscription, q workqueue.RateLimitingInterface, log logr.Logger) {
	subscriptionNames := make([]string, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		request := reconcile.Request{
			NamespacedName: k8stypes.NamespacedName{
				Name:      subscription.Name,
				Namespace: subscription.Namespace,
			},
		}
		q.Add(request)
		subscriptionNames = append(subscriptionNames, subscription.Name)
	}
	log.Info("Queue Subscription Reconcile Requests", "Subscriptions", subscriptionNames)
}

// isRelevantAPIRuleName returns true if the given name matches the APIRule name pattern
// used by the eventing-controller, otherwise returns false.
func isRelevantAPIRuleName(name string) bool {
	return strings.HasPrefix(name, apiRuleNamePrefix)
}

// computeAPIRuleReadyStatus returns true if all APIRule statuses is ok, otherwise returns false.
func computeAPIRuleReadyStatus(apiRule *apigatewayv1alpha1.APIRule) bool {
	if apiRule == nil || apiRule.Status.APIRuleStatus == nil || apiRule.Status.AccessRuleStatus == nil || apiRule.Status.VirtualServiceStatus == nil {
		return false
	}
	apiRuleStatus := apiRule.Status.APIRuleStatus.Code == apigatewayv1alpha1.StatusOK
	accessRuleStatus := apiRule.Status.AccessRuleStatus.Code == apigatewayv1alpha1.StatusOK
	virtualServiceStatus := apiRule.Status.VirtualServiceStatus.Code == apigatewayv1alpha1.StatusOK
	return apiRuleStatus && accessRuleStatus && virtualServiceStatus
}

// setSubscriptionStatusExternalSink sets the subscription external sink based on the given APIRule service host.
func setSubscriptionStatusExternalSink(subscription *eventingv1alpha1.Subscription, apiRule *apigatewayv1alpha1.APIRule) error {
	if apiRule.Spec.Service == nil {
		return errors.Errorf("APIRule has nil service")
	}

	if apiRule.Spec.Service.Host == nil {
		return errors.Errorf("APIRule has nil host")
	}

	u, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(fmt.Sprintf("subscription: [%s/%s] has invalid sink", subscription.Name, subscription.Namespace)))
	}

	path := u.Path
	if u.Path == "" {
		path = "/"
	}

	subscription.Status.ExternalSink = fmt.Sprintf("%s://%s%s", externalSinkScheme, *apiRule.Spec.Service.Host, path)

	return nil
}

func (r *Reconciler) addFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, Finalizer)
	logger.V(1).Info("Adding finalizer")
	if err := r.Update(ctx, subscription); err != nil {
		return errors.Wrapf(err, "error while adding Finalizer with name: %s", Finalizer)
	}
	logger.V(1).Info("Added finalizer")
	return nil
}

func (r *Reconciler) removeFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	var finalizers []string

	// Build finalizer list without the one the controller owns
	for _, finalizer := range subscription.ObjectMeta.Finalizers {
		if finalizer == Finalizer {
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	logger.V(1).Info("Removing finalizer")
	subscription.ObjectMeta.Finalizers = finalizers
	if err := r.Update(ctx, subscription); err != nil {
		return errors.Wrapf(err, "error while removing Finalizer with name: %s", Finalizer)
	}
	logger.V(1).Info("Removed finalizer")
	return nil
}

// isFinalizerSet checks if a finalizer is set on the Subscription which belongs to this controller
func (r *Reconciler) isFinalizerSet(subscription *eventingv1alpha1.Subscription) bool {
	// Check if finalizer is already set
	for _, finalizer := range subscription.ObjectMeta.Finalizers {
		if finalizer == Finalizer {
			return true
		}
	}
	return false
}

// isInDeletion checks if the Subscription shall be deleted
func isInDeletion(subscription *eventingv1alpha1.Subscription) bool {
	return !subscription.DeletionTimestamp.IsZero()
}

const timeoutRetryActiveEmsStatus = time.Second * 30

// checkStatusActive checks if the subscription is active and if not, sets a timer for retry
func (r *Reconciler) checkStatusActive(subscription *eventingv1alpha1.Subscription) (statusChanged, retry bool, err error) {
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

// isValidScheme returns true if the sink scheme is http or https, otherwise returns false.
func isValidScheme(sink string) bool {
	return strings.HasPrefix(sink, "http://") || strings.HasPrefix(sink, "https://")
}
