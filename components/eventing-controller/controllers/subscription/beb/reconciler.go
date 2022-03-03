package beb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	recerrors "github.com/kyma-project/kyma/components/eventing-controller/controllers/errors"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

// Reconciler reconciles a Subscription object
type Reconciler struct {
	ctx context.Context
	client.Client
	logger            *logger.Logger
	recorder          record.EventRecorder
	Backend           handlers.BEBBackend
	Domain            string
	eventTypeCleaner  eventtype.Cleaner
	oauth2credentials *handlers.OAuth2ClientCredentials
	// nameMapper is used to map the Kyma subscription name to a subscription name on BEB
	nameMapper handlers.NameMapper
}

var (
	Finalizer = eventingv1alpha1.GroupVersion.Group
)

const (
	suffixLength                = 10
	externalHostPrefix          = "web"
	externalSinkScheme          = "https"
	apiRuleNamePrefix           = "webhook-"
	clusterLocalURLSuffix       = "svc.cluster.local"
	reconcilerName              = "beb-subscription-reconciler"
	timeoutRetryActiveEmsStatus = time.Second * 30
)

func NewReconciler(ctx context.Context, client client.Client, logger *logger.Logger, recorder record.EventRecorder, cfg env.Config,
	cleaner eventtype.Cleaner, bebBackend handlers.BEBBackend, credential *handlers.OAuth2ClientCredentials, mapper handlers.NameMapper) *Reconciler {
	if err := bebBackend.Initialize(cfg); err != nil {
		logger.WithContext().Errorw("start reconciler failed", "name", reconcilerName, "error", err)
		panic(err)
	}
	return &Reconciler{
		ctx:               ctx,
		Client:            client,
		logger:            logger,
		recorder:          recorder,
		Backend:           bebBackend,
		Domain:            cfg.Domain,
		eventTypeCleaner:  cleaner,
		oauth2credentials: credential,
		nameMapper:        mapper,
	}
}

// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch
// Generate required RBAC to emit kubernetes events in the controller
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:printcolumn:name="Ready",type=bool,JSONPath=`.status.Ready`

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// fetch current subscription object and ensure the object was not deleted in the meantime
	currentSubscription := &eventingv1alpha1.Subscription{}
	if err := r.Client.Get(ctx, req.NamespacedName, currentSubscription); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// copy the subscription object, so we don't modify the source object
	subscription := currentSubscription.DeepCopy()

	// bind fields to logger
	log := utils.LoggerWithSubscription(r.namedLogger(), subscription)
	log.Debugw("received new reconcile request")

	// instantiate a return object
	result := ctrl.Result{}

	// sync the initial Subscription status
	r.syncInitialStatus(subscription)

	// handle deletion of the subscription
	if isInDeletion(subscription) {
		return r.handleDeleteSubscription(ctx, subscription, log)
	}

	// sync Finalizers, ensure the finalizer is set
	if err := r.syncFinalizer(subscription, log); err != nil {
		if updateErr := r.updateSubscription(ctx, subscription, log); updateErr != nil {
			return ctrl.Result{}, errors.Wrap(err, updateErr.Error())
		}
		return ctrl.Result{}, errors.Wrap(err, "sync finalizer failed")
	}

	// sync APIRule for the desired subscription
	apiRule, err := r.syncAPIRule(ctx, subscription, log)
	if !recerrors.IsSkippable(err) {
		if updateErr := r.updateSubscription(ctx, subscription, log); updateErr != nil {
			return ctrl.Result{}, errors.Wrap(err, updateErr.Error())
		}
		return ctrl.Result{}, err
	}

	// sync the BEB Subscription with the Subscription CR
	ready, err := r.syncBEBSubscription(subscription, apiRule, log)
	if err != nil {
		log.Errorw("sync BEB subscription failed", "error", err)
		if updateErr := r.updateSubscription(ctx, subscription, log); updateErr != nil {
			return ctrl.Result{}, errors.Wrap(err, updateErr.Error())
		}
		return ctrl.Result{}, err
	}
	// if beb subscription is not ready, then requeue
	if !ready {
		log.Debugw("requeue reconciliation because BEB subscription is not ready")
		result.RequeueAfter = time.Second * 2
	}

	// update the subscription if modified
	if err := r.updateSubscription(ctx, subscription, log); err != nil {
		return ctrl.Result{}, err
	}

	return result, nil
}

// updateSubscription updates the subscription changes to k8s
func (r *Reconciler) updateSubscription(ctx context.Context, subscription *eventingv1alpha1.Subscription, logger *zap.SugaredLogger) error {
	namespacedName := &k8stypes.NamespacedName{
		Name:      subscription.Name,
		Namespace: subscription.Namespace,
	}

	// fetch the latest subscription object, to avoid k8s conflict errors
	latestSubscription := &eventingv1alpha1.Subscription{}
	if err := r.Client.Get(ctx, *namespacedName, latestSubscription); err != nil {
		return err
	}

	// copy new changes to the latest object
	newSubscription := latestSubscription.DeepCopy()
	newSubscription.Status = subscription.Status
	newSubscription.ObjectMeta.Finalizers = subscription.ObjectMeta.Finalizers

	// emit the condition events if needed
	r.emitConditionEvents(latestSubscription, newSubscription, logger)

	// sync sub status with k8s
	if err := r.updateStatus(ctx, latestSubscription, newSubscription, logger); err != nil {
		return err
	}

	// update the subscription object in k8s
	if !reflect.DeepEqual(latestSubscription.ObjectMeta.Finalizers, newSubscription.ObjectMeta.Finalizers) {
		if err := r.Update(ctx, newSubscription); err != nil {
			return errors.Wrapf(err, "remove finalizer failed name: %s", Finalizer)
		}
		logger.Debugw("update subscription meta for finalizers", "oldFinalizers", latestSubscription.ObjectMeta.Finalizers, "newFinalizers", newSubscription.ObjectMeta.Finalizers)
	}

	return nil
}

// emitConditionEvents check each condition, if the condition is modified then emit an event
func (r *Reconciler) emitConditionEvents(oldSubscription, newSubscription *eventingv1alpha1.Subscription, logger *zap.SugaredLogger) {
	for _, condition := range newSubscription.Status.Conditions {
		oldCondition := oldSubscription.Status.FindCondition(condition.Type)
		if oldCondition != nil && conditionEquals(*oldCondition, condition) {
			continue
		}
		// condition is modified, so emit an event
		r.emitConditionEvent(newSubscription, condition)
		logger.Debug("emitted condition event", condition)
	}
}

// updateStatus updates the status to k8s if modified
func (r *Reconciler) updateStatus(ctx context.Context, oldSubscription, newSubscription *eventingv1alpha1.Subscription, logger *zap.SugaredLogger) error {
	// compare the status taking into consideration lastTransitionTime in conditions
	if isSubscriptionStatusEqual(oldSubscription.Status, newSubscription.Status) {
		return nil
	}

	// update the status for subscription in k8s
	if err := r.Status().Update(ctx, newSubscription); err != nil {
		logger.Errorw("update subscription status failed", "error", err)
		return err
	}
	logger.Debugw("updated subscription status", "oldStatus", oldSubscription.Status, "newStatus", newSubscription.Status)

	return nil
}

// syncFinalizer sets the finalizer in the Subscription
func (r *Reconciler) syncFinalizer(subscription *eventingv1alpha1.Subscription, logger *zap.SugaredLogger) error {
	// Check if finalizer is already set
	if r.isFinalizerSet(subscription) {
		return nil
	}

	return r.addFinalizer(subscription, logger)
}

func (r *Reconciler) handleDeleteSubscription(ctx context.Context, subscription *eventingv1alpha1.Subscription, logger *zap.SugaredLogger) (ctrl.Result, error) {
	// delete beb subscriptions
	if err := r.deleteBEBSubscription(subscription, logger); err != nil {
		return ctrl.Result{}, err
	}

	// update condition in subscription status
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionDeleted, corev1.ConditionFalse, "")
	r.replaceStatusCondition(subscription, condition)

	// remove finalizers from subscription
	r.removeFinalizer(subscription)

	// update subscription CR with changes
	if err := r.updateSubscription(ctx, subscription, logger); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

// syncBEBSubscription delegates the subscription synchronization to the backend client. It returns true if the subscription is ready.
func (r *Reconciler) syncBEBSubscription(subscription *eventingv1alpha1.Subscription, apiRule *apigatewayv1alpha1.APIRule, logger *zap.SugaredLogger) (bool, error) {
	logger.Debug("sync subscription with BEB")

	if apiRule == nil {
		return false, errors.Errorf("APIRule is required")
	}

	if _, err := r.Backend.SyncSubscription(subscription, r.eventTypeCleaner, apiRule); err != nil {
		logger.Errorw("update BEB subscription failed", "error", err)

		r.syncConditionSubscribed(subscription, false)
		return false, err
	}

	// check if the beb subscription is active
	isActive, err := r.checkStatusActive(subscription)
	if err != nil {
		logger.Errorw("timeout at retry", "error", err)
		return false, err
	}

	// sync the condition: ConditionSubscribed
	r.syncConditionSubscribed(subscription, true)

	// sync the condition: ConditionSubscriptionActive
	r.syncConditionSubscriptionActive(subscription, isActive, logger)

	// sync the condition: WebhookCallStatus
	r.syncConditionWebhookCallStatus(subscription)

	return isActive, nil
}

// syncConditionSubscribed syncs the condition ConditionSubscribed
func (r *Reconciler) syncConditionSubscribed(subscription *eventingv1alpha1.Subscription, isSubscribed bool) {
	message := ""
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreationFailed, corev1.ConditionFalse, "")
	if isSubscribed {
		// Include the BEB subscription ID in the Condition message
		message = eventingv1alpha1.CreateMessageForConditionReasonSubscriptionCreated(r.nameMapper.MapSubscriptionName(subscription))
		condition = eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue, message)
	}

	r.replaceStatusCondition(subscription, condition)
}

// syncConditionSubscriptionActive syncs the condition ConditionSubscribed
func (r *Reconciler) syncConditionSubscriptionActive(subscription *eventingv1alpha1.Subscription, isActive bool, logger *zap.SugaredLogger) {
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionActive, corev1.ConditionTrue, "")
	if !isActive {
		logger.Debugw("wait for subscription to be active", "name", subscription.Name, "status", subscription.Status.EmsSubscriptionStatus.SubscriptionStatus)
		condition = eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionNotActive, corev1.ConditionFalse, "")
	}
	r.replaceStatusCondition(subscription, condition)
}

// syncConditionWebhookCallStatus syncs the condition WebhookCallStatus
// checks if the last webhook call returned an error
func (r *Reconciler) syncConditionWebhookCallStatus(subscription *eventingv1alpha1.Subscription) {
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionWebhookCallStatus, eventingv1alpha1.ConditionReasonWebhookCallStatus, corev1.ConditionFalse, "")
	if isWebhookCallError, err := r.checkLastFailedDelivery(subscription); err != nil {
		condition.Message = err.Error()
	} else if isWebhookCallError {
		condition.Message = subscription.Status.EmsSubscriptionStatus.LastFailedDeliveryReason
	} else {
		condition.Status = corev1.ConditionTrue
	}
	r.replaceStatusCondition(subscription, condition)
}

// deleteBEBSubscription deletes the BEB subscription and updates the condition and k8s events
func (r *Reconciler) deleteBEBSubscription(subscription *eventingv1alpha1.Subscription, logger *zap.SugaredLogger) error {
	logger.Debug("delete BEB subscription")
	if err := r.Backend.DeleteSubscription(subscription); err != nil {
		return err
	}

	return nil
}

// syncAPIRule validate the given subscription sink URL and sync its APIRule.
func (r *Reconciler) syncAPIRule(ctx context.Context, subscription *eventingv1alpha1.Subscription, logger *zap.SugaredLogger) (*apigatewayv1alpha1.APIRule, error) {
	if err := r.isSinkURLValid(ctx, subscription); err != nil {
		return nil, err
	}

	sURL, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Parse sink URI failed %s", subscription.Spec.Sink)
		return nil, recerrors.NewSkippable(errors.Wrapf(err, "parse sink URI failed"))
	}

	apiRule, err := r.createOrUpdateAPIRule(ctx, subscription, *sURL, logger)
	if err != nil {
		return nil, errors.Wrap(err, "create or update APIRule failed")
	}

	if apiRule != nil {
		subscription.Status.APIRuleName = apiRule.Name
	}

	// check if the apiRule is ready
	apiRuleReady := computeAPIRuleReadyStatus(apiRule)

	// sync the condition: ConditionAPIRuleStatus
	subscription.Status.SetConditionAPIRuleStatus(apiRuleReady)
	// set subscription sink only if the APIRule is ready
	if apiRuleReady {
		if err := setSubscriptionStatusExternalSink(subscription, apiRule); err != nil {
			return apiRule, errors.Wrapf(err, "set subscription status externalSink failed namespace:%s name:%s", subscription.Namespace, subscription.Name)
		}
	}

	return apiRule, nil
}

func (r *Reconciler) isSinkURLValid(ctx context.Context, subscription *eventingv1alpha1.Subscription) error {
	if !isValidScheme(subscription.Spec.Sink) {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Sink URL scheme should be HTTP or HTTPS %s", subscription.Spec.Sink)
		return recerrors.NewSkippable(fmt.Errorf("sink URL scheme should be 'http' or 'https'"))
	}

	sURL, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Sink URL is not valid %s", err.Error())
		return recerrors.NewSkippable(err)
	}

	// Validate sink URL is a cluster local URL
	trimmedHost := strings.Split(sURL.Host, ":")[0]
	if !strings.HasSuffix(trimmedHost, clusterLocalURLSuffix) {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Sink does not contain suffix %s", clusterLocalURLSuffix)
		return recerrors.NewSkippable(fmt.Errorf("sink does not contain suffix: %s in the URL", clusterLocalURLSuffix))
	}

	// we expected a sink in the format "service.namespace.svc.cluster.local"
	subDomains := strings.Split(trimmedHost, ".")
	if len(subDomains) != 5 {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Sink should contain 5 sub-domains %s", trimmedHost)
		return recerrors.NewSkippable(fmt.Errorf("sink should contain 5 sub-domains: %s", trimmedHost))
	}

	// Assumption: Subscription CR and Subscriber should be deployed in the same namespace
	svcNs := subDomains[1]
	if subscription.Namespace != svcNs {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Namespace of subscription %s and the subscriber %s are different", subscription.Namespace, svcNs)
		return recerrors.NewSkippable(fmt.Errorf("namespace of subscription: %s and the namespace of subscriber: %s are different", subscription.Namespace, svcNs))
	}

	// Validate svc is a cluster-local one
	svcName := subDomains[0]
	if _, err := r.getClusterLocalService(ctx, svcNs, svcName); err != nil {
		if k8serrors.IsNotFound(err) {
			events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Sink does not correspond to a valid cluster local svc")
			return recerrors.NewSkippable(errors.Wrapf(err, "sink is not valid cluster local svc"))
		}

		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Fetch cluster-local svc failed namespace %s name %s", svcNs, svcName)
		return errors.Wrapf(err, "fetch cluster-local svc failed namespace:%s name:%s", svcNs, svcName)
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
func (r *Reconciler) createOrUpdateAPIRule(ctx context.Context, subscription *eventingv1alpha1.Subscription, sink url.URL, logger *zap.SugaredLogger) (*apigatewayv1alpha1.APIRule, error) {
	svcNs, svcName, err := getSvcNsAndName(sink.Host)
	if err != nil {
		return nil, errors.Wrap(err, "parse svc name and ns in create or update APIRule failed")
	}
	labels := map[string]string{
		constants.ControllerServiceLabelKey:  svcName,
		constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
	}

	svcPort, err := utils.GetPortNumberFromURL(sink)
	if err != nil {
		return nil, errors.Wrap(err, "convert URL port to APIRule port failed")
	}
	var reusableAPIRule *apigatewayv1alpha1.APIRule
	existingAPIRules, err := r.getAPIRulesForASvc(ctx, labels, svcNs)
	if err != nil {
		return nil, errors.Wrapf(err, "fetch APIRule failed for labels: %v", labels)
	}
	if existingAPIRules != nil {
		reusableAPIRule = r.filterAPIRulesOnPort(existingAPIRules, svcPort)
	}

	// Get all subscriptions valid for the cluster-local subscriber
	subscriptions, err := r.getSubscriptionsForASvc(ctx, svcNs, svcName)
	if err != nil {
		return nil, errors.Wrapf(err, "fetch subscriptions failed for subscriber namespace:%s name:%s", svcNs, svcName)
	}
	filteredSubscriptions := r.filterSubscriptionsOnPort(subscriptions, svcPort)

	desiredAPIRule := r.makeAPIRule(svcNs, svcName, labels, filteredSubscriptions, svcPort)
	if err != nil {
		return nil, errors.Wrap(err, "make APIRule failed")
	}

	// update or remove the previous APIRule if it is not used by other subscriptions
	if err := r.handlePreviousAPIRule(ctx, subscription, reusableAPIRule); err != nil {
		return nil, err
	}

	// no APIRule to reuse, create a new one
	if reusableAPIRule == nil {
		if err := r.Client.Create(ctx, desiredAPIRule, &client.CreateOptions{}); err != nil {
			events.Warn(r.recorder, subscription, events.ReasonCreateFailed, "Create APIRule failed %s", desiredAPIRule.Name)
			return nil, errors.Wrap(err, "create APIRule failed")
		}

		events.Normal(r.recorder, subscription, events.ReasonCreate, "Create APIRule succeeded %s", desiredAPIRule.Name)
		return desiredAPIRule, nil
	}
	logger.Debugw("reuse APIRule", "namespace", svcNs, "name", reusableAPIRule.Name, "service", svcName)

	object.ApplyExistingAPIRuleAttributes(reusableAPIRule, desiredAPIRule)
	if object.Semantic.DeepEqual(reusableAPIRule, desiredAPIRule) {
		return reusableAPIRule, nil
	}
	err = r.Client.Update(ctx, desiredAPIRule, &client.UpdateOptions{})
	if err != nil {
		events.Warn(r.recorder, subscription, events.ReasonUpdateFailed, "Update APIRule failed %s", desiredAPIRule.Name)
		return nil, errors.Wrap(err, "update APIRule failed")
	}
	events.Normal(r.recorder, subscription, events.ReasonUpdate, "Update APIRule succeeded %s", desiredAPIRule.Name)

	return desiredAPIRule, nil
}

// handlePreviousAPIRule computes the OwnerReferences list for the previous subscription APIRule (if any)
// if the OwnerReferences list is empty, then the APIRule will be deleted
// else if the OwnerReferences list length was decreased, then the APIRule will be updated
// TODO write more tests https://github.com/kyma-project/kyma/issues/9950
func (r *Reconciler) handlePreviousAPIRule(ctx context.Context, subscription *eventingv1alpha1.Subscription, reusableAPIRule *apigatewayv1alpha1.APIRule) error {
	// subscription does not have a previous APIRule
	if len(subscription.Status.APIRuleName) == 0 {
		return nil
	}

	// the previous APIRule for the subscription is the current one no need to update it
	if reusableAPIRule != nil && subscription.Status.APIRuleName == reusableAPIRule.Name {
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

	// delete the APIRule if the new OwnerReference list is empty
	if len(ownerReferences) == 0 {
		if err := r.Client.Delete(ctx, previousAPIRule); err != nil {
			return err
		}
		return nil
	}

	// update the APIRule if the new OwnerReference list length is decreased
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

// getSubscriptionsForASvc returns a list of Subscriptions which are valid for the subscriber in focus
func (r *Reconciler) getSubscriptionsForASvc(ctx context.Context, svcNs, svcName string) ([]eventingv1alpha1.Subscription, error) {
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

func (r *Reconciler) makeAPIRule(svcNs, svcName string, labels map[string]string, subs []eventingv1alpha1.Subscription, port uint32) *apigatewayv1alpha1.APIRule {

	randomSuffix := handlers.GetRandString(suffixLength)
	hostName := fmt.Sprintf("%s-%s.%s", externalHostPrefix, randomSuffix, r.Domain)

	apiRule := object.NewAPIRule(svcNs, apiRuleNamePrefix,
		object.WithLabels(labels),
		object.WithOwnerReference(subs),
		object.WithService(hostName, svcName, port),
		object.WithGateway(constants.ClusterLocalAPIGateway),
		object.WithRules(subs, http.MethodPost, http.MethodOptions))
	return apiRule
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

// syncInitialStatus determines the desired initial status and updates it accordingly (if conditions changed)
func (r *Reconciler) syncInitialStatus(subscription *eventingv1alpha1.Subscription) {
	expectedStatus := eventingv1alpha1.SubscriptionStatus{}
	expectedStatus.InitializeConditions()

	// case: conditions are already initialized and there is no change in the Ready status
	if eventingv1alpha1.ContainSameConditionTypes(subscription.Status.Conditions, expectedStatus.Conditions) &&
		!subscription.Status.ShouldUpdateReadyStatus() {
		return
	}

	if len(subscription.Status.Conditions) == 0 {
		subscription.Status = expectedStatus
	} else {
		requiredConditions := getRequiredConditions(subscription.Status.Conditions, expectedStatus.Conditions)
		subscription.Status.Conditions = requiredConditions
		subscription.Status.Ready = !subscription.Status.Ready
	}

	// reset the status for apiRule
	subscription.Status.APIRuleName = ""
	subscription.Status.ExternalSink = ""
	subscription.Status.SetConditionAPIRuleStatus(false)
}

// getRequiredConditions removes the non-required conditions from the subscription  and adds any missing required-conditions
func getRequiredConditions(subscriptionConditions, expectedConditions []eventingv1alpha1.Condition) []eventingv1alpha1.Condition {
	var requiredConditions []eventingv1alpha1.Condition
	expectedConditionsMap := make(map[eventingv1alpha1.ConditionType]eventingv1alpha1.Condition)
	for _, condition := range expectedConditions {
		expectedConditionsMap[condition.Type] = condition
	}

	// add the current subscription's conditions if it exists in the expectedConditions
	for _, condition := range subscriptionConditions {
		if _, ok := expectedConditionsMap[condition.Type]; ok {
			requiredConditions = append(requiredConditions, condition)
			delete(expectedConditionsMap, condition.Type)
		}
	}
	// add the remaining conditions that weren't present in the subscription
	for _, condition := range expectedConditionsMap {
		requiredConditions = append(requiredConditions, condition)
	}

	return requiredConditions
}

// replaceStatusCondition replaces the given condition on the subscription. Also it sets the readiness in the status.
// So make sure you always use this method then changing a condition
func (r *Reconciler) replaceStatusCondition(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition) bool {
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
		return false
	}

	// update the status
	subscription.Status.Conditions = desiredConditions
	subscription.Status.Ready = isReady
	return true
}

// emitConditionEvent emits a kubernetes event and sets the event type based on the Condition status
func (r *Reconciler) emitConditionEvent(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition) {
	eventType := corev1.EventTypeNormal
	if condition.Status == corev1.ConditionFalse {
		eventType = corev1.EventTypeWarning
	}
	r.recorder.Event(subscription, eventType, string(condition.Reason), condition.Message)
}

// SetupUnmanaged creates a controller under the client control
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged(reconcilerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.namedLogger().Errorw("create unmanaged controller failed", "name", reconcilerName, "error", err)
		return err
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha1.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		r.namedLogger().Errorw("watch subscriptions failed", "error", err)
		return err
	}

	apiRuleEventHandler := &handler.EnqueueRequestForOwner{OwnerType: &eventingv1alpha1.Subscription{}, IsController: false}
	if err := ctru.Watch(&source.Kind{Type: &apigatewayv1alpha1.APIRule{}}, apiRuleEventHandler); err != nil {
		r.namedLogger().Errorw("watch APIRule failed", "error", err)
		return err
	}

	go func(r *Reconciler, c controller.Controller) {
		if err := c.Start(r.ctx); err != nil {
			r.namedLogger().Errorw("start controller failed", "name", reconcilerName, "error", err)
			os.Exit(1)
		}
	}(r, ctru)

	return nil
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
		return errors.Wrapf(err, "invalid sink for subscription namespace:%s name:%s", subscription.Namespace, subscription.Name)
	}

	path := u.Path
	if u.Path == "" {
		path = "/"
	}

	subscription.Status.ExternalSink = fmt.Sprintf("%s://%s%s", externalSinkScheme, *apiRule.Spec.Service.Host, path)

	return nil
}

func (r *Reconciler) addFinalizer(subscription *eventingv1alpha1.Subscription, logger *zap.SugaredLogger) error {
	subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, Finalizer)
	logger.Debug("add finalizer")
	return nil
}

func (r *Reconciler) removeFinalizer(subscription *eventingv1alpha1.Subscription) {
	var finalizers []string

	// Build finalizer list without the one the controller owns
	for _, finalizer := range subscription.ObjectMeta.Finalizers {
		if finalizer == Finalizer {
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	subscription.ObjectMeta.Finalizers = finalizers
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

// checkStatusActive checks if the subscription is active and if not, sets a timer for retry
func (r *Reconciler) checkStatusActive(subscription *eventingv1alpha1.Subscription) (active bool, err error) {
	// check if the EMS subscription status is active
	if subscription.Status.EmsSubscriptionStatus.SubscriptionStatus == string(types.SubscriptionStatusActive) {
		if len(subscription.Status.FailedActivation) > 0 {
			subscription.Status.FailedActivation = ""
		}
		return true, nil
	}

	t1 := time.Now()
	if len(subscription.Status.FailedActivation) == 0 {
		// it's the first time
		subscription.Status.FailedActivation = t1.Format(time.RFC3339)
		return false, nil
	}

	// check the timeout
	if t0, er := time.Parse(time.RFC3339, subscription.Status.FailedActivation); er != nil {
		err = er
	} else if t1.Sub(t0) > timeoutRetryActiveEmsStatus {
		err = fmt.Errorf("timeout waiting for the subscription to be active: %v", subscription.Name)
	}

	return false, err
}

// checkLastFailedDelivery checks if LastFailedDelivery exists and if it happened after LastSuccessfulDelivery
func (r *Reconciler) checkLastFailedDelivery(subscription *eventingv1alpha1.Subscription) (bool, error) {
	if len(subscription.Status.EmsSubscriptionStatus.LastFailedDelivery) > 0 {
		var lastFailedDeliveryTime, LastSuccessfulDeliveryTime time.Time
		var err error
		if lastFailedDeliveryTime, err = time.Parse(time.RFC3339, subscription.Status.EmsSubscriptionStatus.LastFailedDelivery); err != nil {
			r.namedLogger().Errorw("parse LastFailedDelivery failed", "error", err)
			return true, err
		}
		if len(subscription.Status.EmsSubscriptionStatus.LastSuccessfulDelivery) > 0 {
			if LastSuccessfulDeliveryTime, err = time.Parse(time.RFC3339, subscription.Status.EmsSubscriptionStatus.LastSuccessfulDelivery); err != nil {
				r.namedLogger().Errorw("parse LastSuccessfulDelivery failed", "error", err)
				return true, err
			}
		}
		if lastFailedDeliveryTime.After(LastSuccessfulDeliveryTime) {
			return true, nil
		}
	}
	return false, nil
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}

// isValidScheme returns true if the sink scheme is http or https, otherwise returns false.
func isValidScheme(sink string) bool {
	return strings.HasPrefix(sink, "http://") || strings.HasPrefix(sink, "https://")
}
