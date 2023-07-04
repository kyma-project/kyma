package eventmesh

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"

	recerrors "github.com/kyma-project/kyma/components/eventing-controller/controllers/errors"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"golang.org/x/xerrors"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"

	"go.uber.org/zap"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

// Reconciler reconciles a Subscription object.
type Reconciler struct {
	ctx context.Context
	client.Client
	logger            *logger.Logger
	recorder          record.EventRecorder
	Backend           eventmesh.Backend
	Domain            string
	cleaner           cleaner.Cleaner
	oauth2credentials *eventmesh.OAuth2ClientCredentials
	// nameMapper is used to map the Kyma subscription name to a subscription name on EventMesh.
	nameMapper    backendutils.NameMapper
	sinkValidator sink.Validator
}

const (
	suffixLength                = 10
	externalHostPrefix          = "web"
	externalSinkScheme          = "https"
	apiRuleNamePrefix           = "webhook-"
	reconcilerName              = "eventMesh-subscription-reconciler"
	timeoutRetryActiveEmsStatus = time.Second * 30
	requeueAfterDuration        = time.Second * 2
)

func NewReconciler(ctx context.Context, client client.Client, logger *logger.Logger, recorder record.EventRecorder,
	cfg env.Config, cleaner cleaner.Cleaner, eventMeshBackend eventmesh.Backend,
	credential *eventmesh.OAuth2ClientCredentials, mapper backendutils.NameMapper, validator sink.Validator) *Reconciler {
	if err := eventMeshBackend.Initialize(cfg); err != nil {
		logger.WithContext().Errorw("Failed to start reconciler", "name",
			reconcilerName, "error", err)
		panic(err)
	}
	return &Reconciler{
		ctx:               ctx,
		Client:            client,
		logger:            logger,
		recorder:          recorder,
		Backend:           eventMeshBackend,
		Domain:            cfg.Domain,
		cleaner:           cleaner,
		oauth2credentials: credential,
		nameMapper:        mapper,
		sinkValidator:     validator,
	}
}

// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch
// Generate required RBAC to emit kubernetes events in the controller.
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules,verbs=get;list;watch;create;update;patch;delete
// Generated required RBAC to list Applications (required by event type cleaner).
// +kubebuilder:rbac:groups="applicationconnector.kyma-project.io",resources=applications,verbs=get;list;watch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// fetch current subscription object and ensure the object was not deleted in the meantime
	currentSubscription := &eventingv1alpha2.Subscription{}
	if err := r.Client.Get(ctx, req.NamespacedName, currentSubscription); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// copy the subscription object, so we don't modify the source object
	sub := currentSubscription.DeepCopy()

	// bind fields to logger
	log := backendutils.LoggerWithSubscription(r.namedLogger(), sub)
	log.Debugw("Received new reconcile request")

	// instantiate a return object
	result := ctrl.Result{}

	// handle deletion of the subscription
	if isInDeletion(sub) {
		return r.handleDeleteSubscription(ctx, sub, log)
	}

	// sync the initial Subscription status
	r.syncInitialStatus(sub)

	// sync Finalizers, ensure the finalizer is set
	if err := r.syncFinalizer(sub, log); err != nil {
		if updateErr := r.updateSubscription(ctx, sub, log); updateErr != nil {
			return ctrl.Result{}, xerrors.Errorf(updateErr.Error()+": %v", err)
		}
		return ctrl.Result{}, xerrors.Errorf("failed to sync finalizer: %v", err)
	}

	// sync APIRule for the desired subscription
	apiRule, err := r.syncAPIRule(ctx, sub, log)
	// sync the condition: ConditionAPIRuleStatus
	sub.Status.SetConditionAPIRuleStatus(err)
	if !recerrors.IsSkippable(err) {
		if updateErr := r.updateSubscription(ctx, sub, log); updateErr != nil {
			return ctrl.Result{}, xerrors.Errorf(updateErr.Error()+": %v", err)
		}
		return ctrl.Result{}, err
	}

	// sync the EventMesh Subscription with the Subscription CR
	ready, err := r.syncEventMeshSubscription(sub, apiRule, log)
	if err != nil {
		if updateErr := r.updateSubscription(ctx, sub, log); updateErr != nil {
			return ctrl.Result{}, xerrors.Errorf(updateErr.Error()+": %v", err)
		}
		return ctrl.Result{}, err
	}
	// if eventMesh subscription is not ready, then requeue
	if !ready {
		log.Debugw("Requeuing reconciliation because EventMesh subscription is not ready")
		result.RequeueAfter = requeueAfterDuration
	}

	// update the subscription if modified
	if err := r.updateSubscription(ctx, sub, log); err != nil {
		return ctrl.Result{}, err
	}

	return result, nil
}

// updateSubscription updates the subscription changes to k8s.
func (r *Reconciler) updateSubscription(ctx context.Context, sub *eventingv1alpha2.Subscription, logger *zap.SugaredLogger) error {
	namespacedName := &k8stypes.NamespacedName{
		Name:      sub.Name,
		Namespace: sub.Namespace,
	}

	// fetch the latest subscription object, to avoid k8s conflict errors
	latestSubscription := &eventingv1alpha2.Subscription{}
	if err := r.Client.Get(ctx, *namespacedName, latestSubscription); err != nil {
		return err
	}

	// copy new changes to the latest object
	newSubscription := latestSubscription.DeepCopy()
	newSubscription.Status = sub.Status
	newSubscription.ObjectMeta.Finalizers = sub.ObjectMeta.Finalizers

	// emit the condition events if needed
	r.emitConditionEvents(latestSubscription, newSubscription, logger)

	// sync sub status with k8s
	if err := r.updateStatus(ctx, latestSubscription, newSubscription, logger); err != nil {
		return err
	}

	// update the subscription object in k8s
	if !reflect.DeepEqual(latestSubscription.ObjectMeta.Finalizers, newSubscription.ObjectMeta.Finalizers) {
		if err := r.Update(ctx, newSubscription); err != nil {
			return xerrors.Errorf("failed to remove finalizer name '%s': %v", eventingv1alpha2.Finalizer, err)
		}
		logger.Debugw("Updated subscription meta for finalizers", "oldFinalizers", latestSubscription.ObjectMeta.Finalizers, "newFinalizers", newSubscription.ObjectMeta.Finalizers)
	}

	return nil
}

// emitConditionEvents check each condition, if the condition is modified then emit an event.
func (r *Reconciler) emitConditionEvents(oldSubscription, newSubscription *eventingv1alpha2.Subscription, logger *zap.SugaredLogger) {
	for _, condition := range newSubscription.Status.Conditions {
		oldCondition := oldSubscription.Status.FindCondition(condition.Type)
		if oldCondition != nil && eventingv1alpha2.ConditionEquals(*oldCondition, condition) {
			continue
		}
		// condition is modified, so emit an event
		r.emitConditionEvent(newSubscription, condition)
		logger.Debug("Emitted condition event", condition)
	}
}

// updateStatus updates the status to k8s if modified.
func (r *Reconciler) updateStatus(ctx context.Context, oldSubscription, newSubscription *eventingv1alpha2.Subscription, logger *zap.SugaredLogger) error {
	// compare the status taking into consideration lastTransitionTime in conditions
	if object.IsSubscriptionStatusEqual(oldSubscription.Status, newSubscription.Status) {
		return nil
	}

	// update the status for subscription in k8s
	if err := r.Status().Update(ctx, newSubscription); err != nil {
		return xerrors.Errorf("failed to update subscription status: %v", err)
	}
	logger.Debugw("Updated subscription status", "oldStatus", oldSubscription.Status, "newStatus", newSubscription.Status)

	return nil
}

// syncFinalizer sets the finalizer in the Subscription.
func (r *Reconciler) syncFinalizer(subscription *eventingv1alpha2.Subscription, logger *zap.SugaredLogger) error {
	// Check if finalizer is already set
	if isFinalizerSet(subscription) {
		return nil
	}

	return addFinalizer(subscription, logger)
}

func (r *Reconciler) handleDeleteSubscription(ctx context.Context, subscription *eventingv1alpha2.Subscription,
	logger *zap.SugaredLogger) (ctrl.Result, error) {
	// delete EventMesh subscriptions
	logger.Debug("Deleting subscription on EventMesh")
	if err := r.Backend.DeleteSubscription(subscription); err != nil {
		return ctrl.Result{}, err
	}

	// update condition in subscription status
	condition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscribed,
		eventingv1alpha2.ConditionReasonSubscriptionDeleted, corev1.ConditionFalse, "")
	r.replaceStatusCondition(subscription, condition)

	// remove finalizers from subscription
	removeFinalizer(subscription)

	// update subscription CR with changes
	if err := r.updateSubscription(ctx, subscription, logger); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

// syncEventMeshSubscription delegates the subscription synchronization to the backend client. It returns true if the subscription is ready.
func (r *Reconciler) syncEventMeshSubscription(subscription *eventingv1alpha2.Subscription, apiRule *apigatewayv1beta1.APIRule, logger *zap.SugaredLogger) (bool, error) {
	logger.Debug("Syncing subscription with EventMesh")

	if apiRule == nil {
		return false, errors.Errorf("APIRule is required")
	}

	if _, err := r.Backend.SyncSubscription(subscription, r.cleaner, apiRule); err != nil {
		r.syncConditionSubscribed(subscription, err)
		return false, err
	}

	// check if the eventMesh subscription is active
	isActive, err := r.checkStatusActive(subscription)
	if err != nil {
		return false, xerrors.Errorf("reached retry timeout: %v", err)
	}

	// sync the condition: ConditionSubscribed
	r.syncConditionSubscribed(subscription, nil)

	// sync the condition: ConditionSubscriptionActive
	r.syncConditionSubscriptionActive(subscription, isActive, logger)

	// sync the condition: WebhookCallStatus
	r.syncConditionWebhookCallStatus(subscription)

	return isActive, nil
}

// syncConditionSubscribed syncs the condition ConditionSubscribed.
func (r *Reconciler) syncConditionSubscribed(subscription *eventingv1alpha2.Subscription, err error) {
	// Include the EventMesh subscription ID in the Condition message
	message := eventingv1alpha2.CreateMessageForConditionReasonSubscriptionCreated(r.nameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace))
	condition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscribed, eventingv1alpha2.ConditionReasonSubscriptionCreated, corev1.ConditionTrue, message)
	if err != nil {
		message = err.Error()
		condition = eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscribed, eventingv1alpha2.ConditionReasonSubscriptionCreationFailed, corev1.ConditionFalse, message)
	}

	r.replaceStatusCondition(subscription, condition)
}

// syncConditionSubscriptionActive syncs the condition ConditionSubscribed.
func (r *Reconciler) syncConditionSubscriptionActive(subscription *eventingv1alpha2.Subscription, isActive bool, logger *zap.SugaredLogger) {
	condition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscriptionActive,
		eventingv1alpha2.ConditionReasonSubscriptionActive,
		corev1.ConditionTrue,
		"")
	if !isActive {
		logger.Debugw("Waiting for subscription to be active",
			"name",
			subscription.Name,
			"status",
			subscription.Status.Backend.EventMeshSubscriptionStatus.Status)

		message := "Waiting for subscription to be active"
		condition = eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscriptionActive,
			eventingv1alpha2.ConditionReasonSubscriptionNotActive,
			corev1.ConditionFalse,
			message)
	}
	r.replaceStatusCondition(subscription, condition)
}

// syncConditionWebhookCallStatus syncs the condition WebhookCallStatus
// checks if the last webhook call returned an error.
func (r *Reconciler) syncConditionWebhookCallStatus(subscription *eventingv1alpha2.Subscription) {
	condition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionWebhookCallStatus, eventingv1alpha2.ConditionReasonWebhookCallStatus, corev1.ConditionFalse, "")
	if isWebhookCallError, err := r.checkLastFailedDelivery(subscription); err != nil {
		condition.Message = err.Error()
	} else if isWebhookCallError {
		condition.Message = subscription.Status.Backend.EventMeshSubscriptionStatus.LastFailedDeliveryReason
	} else {
		condition.Status = corev1.ConditionTrue
	}
	r.replaceStatusCondition(subscription, condition)
}

// syncAPIRule validate the given subscription sink URL and sync its APIRule.
func (r *Reconciler) syncAPIRule(ctx context.Context, subscription *eventingv1alpha2.Subscription,
	logger *zap.SugaredLogger) (*apigatewayv1beta1.APIRule, error) {
	if err := r.sinkValidator.Validate(subscription); err != nil {
		return nil, err
	}

	sURL, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed,
			"Parse sink URI failed %s", subscription.Spec.Sink)
		return nil, recerrors.NewSkippable(xerrors.Errorf("failed to parse sink URL: %v", err))
	}

	apiRule, err := r.createOrUpdateAPIRule(ctx, subscription, *sURL, logger)
	if err != nil {
		return nil, xerrors.Errorf("failed to create or update APIRule: %v", err)
	}

	if apiRule != nil {
		subscription.Status.Backend.APIRuleName = apiRule.Name
	}

	// check if the apiRule is ready
	apiRuleReady := computeAPIRuleReadyStatus(apiRule)

	// set subscription sink only if the APIRule is ready
	if apiRuleReady {
		if err := setSubscriptionStatusExternalSink(subscription, apiRule); err != nil {
			return apiRule, xerrors.Errorf("failed to set subscription status externalSink "+
				"namespace=%s, name=%s : %v", subscription.Namespace, subscription.Name, err)
		}
		return apiRule, nil
	}

	return apiRule, recerrors.NewSkippable(errors.Errorf("apiRule %s is not ready", apiRule.Name))
}

// createOrUpdateAPIRule create new or update existing APIRule for the given subscription.
func (r *Reconciler) createOrUpdateAPIRule(ctx context.Context, subscription *eventingv1alpha2.Subscription,
	sink url.URL, logger *zap.SugaredLogger) (*apigatewayv1beta1.APIRule, error) {
	svcNs, svcName, err := getSvcNsAndName(sink.Host)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse svc name and ns in create or update APIRule: %v", err)
	}
	labels := map[string]string{
		constants.ControllerServiceLabelKey:  svcName,
		constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
	}

	svcPort, err := utils.GetPortNumberFromURL(sink)
	if err != nil {
		return nil, xerrors.Errorf("failed to convert URL port to APIRule port: %v", err)
	}
	var reusableAPIRule *apigatewayv1beta1.APIRule
	existingAPIRules, err := r.getAPIRulesForASvc(ctx, labels, svcNs)
	if err != nil {
		return nil, xerrors.Errorf("failed to fetch APIRule for labels=%v : %v", labels, err)
	}
	if existingAPIRules != nil {
		reusableAPIRule = r.filterAPIRulesOnPort(existingAPIRules, svcPort)
	}

	// Get all subscriptions valid for the cluster-local subscriber
	subscriptions, err := r.getSubscriptionsForASvc(ctx, svcNs, svcName)
	if err != nil {
		return nil, xerrors.Errorf("failed to fetch subscriptions for subscriber namespace=%s, name=%s : %v", svcNs, svcName, err)
	}
	filteredSubscriptions := r.filterSubscriptionsOnPort(subscriptions, svcPort)

	desiredAPIRule := r.makeAPIRule(svcNs, svcName, labels, filteredSubscriptions, svcPort)
	if err != nil {
		return nil, xerrors.Errorf("failed to make APIRule: %v", err)
	}

	// update or remove the previous APIRule if it is not used by other subscriptions
	if err := r.handlePreviousAPIRule(ctx, subscription, reusableAPIRule); err != nil {
		return nil, err
	}

	// no APIRule to reuse, create a new one
	if reusableAPIRule == nil {
		if err := r.Client.Create(ctx, desiredAPIRule, &client.CreateOptions{}); err != nil {
			events.Warn(r.recorder, subscription, events.ReasonCreateFailed, "Create APIRule failed %s", desiredAPIRule.Name)
			return nil, xerrors.Errorf("failed to create APIRule: %v", err)
		}

		events.Normal(r.recorder, subscription, events.ReasonCreate, "Create APIRule succeeded %s", desiredAPIRule.Name)
		return desiredAPIRule, nil
	}
	logger.Debugw("Reusing APIRule", "namespace", svcNs, "name", reusableAPIRule.Name, "service", svcName)

	object.ApplyExistingAPIRuleAttributes(reusableAPIRule, desiredAPIRule)
	if object.Semantic.DeepEqual(reusableAPIRule, desiredAPIRule) {
		return reusableAPIRule, nil
	}
	err = r.Client.Update(ctx, desiredAPIRule, &client.UpdateOptions{})
	if err != nil {
		events.Warn(r.recorder, subscription, events.ReasonUpdateFailed, "Update APIRule failed %s", desiredAPIRule.Name)
		return nil, xerrors.Errorf("failed to update APIRule: %v", err)
	}
	events.Normal(r.recorder, subscription, events.ReasonUpdate, "Update APIRule succeeded %s", desiredAPIRule.Name)

	return desiredAPIRule, nil
}

// handlePreviousAPIRule computes the OwnerReferences list for the previous subscription APIRule (if any)
// if the OwnerReferences list is empty, then the APIRule will be deleted
// else if the OwnerReferences list length was decreased, then the APIRule will be updated.
func (r *Reconciler) handlePreviousAPIRule(ctx context.Context, subscription *eventingv1alpha2.Subscription,
	reusableAPIRule *apigatewayv1beta1.APIRule) error {
	// subscription does not have a previous APIRule
	if len(subscription.Status.Backend.APIRuleName) == 0 {
		return nil
	}

	// the previous APIRule for the subscription is the current one no need to update it
	if reusableAPIRule != nil && subscription.Status.Backend.APIRuleName == reusableAPIRule.Name {
		return nil
	}

	// get the previous APIRule
	previousAPIRule := &apigatewayv1beta1.APIRule{}
	key := k8stypes.NamespacedName{Namespace: subscription.Namespace, Name: subscription.Status.Backend.APIRuleName}
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
		namespaceSubscriptions := &eventingv1alpha2.SubscriptionList{}
		if err := r.Client.List(ctx, namespaceSubscriptions, &client.ListOptions{Namespace: previousAPIRule.Namespace}); err != nil {
			return err
		}

		// build a new subscription list and exclude the current subscription from the list
		subscriptions := make([]eventingv1alpha2.Subscription, 0, len(namespaceSubscriptions.Items))
		for _, namespaceSubscription := range namespaceSubscriptions.Items {
			// skip the current subscription
			if namespaceSubscription.UID == subscription.UID {
				continue
			}

			// skip not relevant subscriptions to the previous APIRule
			if namespaceSubscription.Status.Backend.APIRuleName != previousAPIRule.Name {
				continue
			}

			subscriptions = append(subscriptions, namespaceSubscription)
		}

		// update the APIRule OwnerReferences list and Spec Rules
		object.WithOwnerReference(subscriptions)(previousAPIRule)
		object.WithRules(
			r.oauth2credentials.CertsURL,
			subscriptions,
			*previousAPIRule.Spec.Service,
			http.MethodPost,
			http.MethodOptions,
		)(previousAPIRule)

		if err := r.Client.Update(ctx, previousAPIRule); err != nil {
			return err
		}
	}

	return nil
}

// getSubscriptionsForASvc returns a list of Subscriptions which are valid for the subscriber in focus.
func (r *Reconciler) getSubscriptionsForASvc(ctx context.Context, svcNs, svcName string) ([]eventingv1alpha2.Subscription, error) {
	subscriptions := &eventingv1alpha2.SubscriptionList{}
	relevantSubs := make([]eventingv1alpha2.Subscription, 0)
	err := r.Client.List(ctx, subscriptions, &client.ListOptions{
		Namespace: svcNs,
	})
	if err != nil {
		return []eventingv1alpha2.Subscription{}, err
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

// filterSubscriptionsOnPort returns a list of Subscriptions which matches a particular port.
func (r *Reconciler) filterSubscriptionsOnPort(subList []eventingv1alpha2.Subscription, svcPort uint32) []eventingv1alpha2.Subscription {
	filteredSubs := make([]eventingv1alpha2.Subscription, 0)
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

func (r *Reconciler) makeAPIRule(svcNs, svcName string, labels map[string]string, subs []eventingv1alpha2.Subscription, port uint32) *apigatewayv1beta1.APIRule {
	randomSuffix := utils.GetRandString(suffixLength)
	hostName := fmt.Sprintf("%s-%s.%s", externalHostPrefix, randomSuffix, r.Domain)
	svc := object.GetService(svcName, port)
	apiRule := object.NewAPIRule(svcNs, apiRuleNamePrefix,
		object.WithLabels(labels),
		object.WithOwnerReference(subs),
		object.WithService(hostName, svcName, port),
		object.WithGateway(constants.ClusterLocalAPIGateway),
		object.WithRules(r.oauth2credentials.CertsURL, subs, svc, http.MethodPost, http.MethodOptions))
	return apiRule
}

func (r *Reconciler) getAPIRulesForASvc(ctx context.Context, labels map[string]string, svcNs string) ([]apigatewayv1beta1.APIRule, error) {
	existingAPIRules := &apigatewayv1beta1.APIRuleList{}
	err := r.Client.List(ctx, existingAPIRules, &client.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(labels),
		Namespace:     svcNs,
	})
	if err != nil {
		return nil, err
	}
	return existingAPIRules.Items, nil
}

func (r *Reconciler) filterAPIRulesOnPort(existingAPIRules []apigatewayv1beta1.APIRule, port uint32) *apigatewayv1beta1.APIRule {
	// Assumption: there will be one APIRule for an svc with the labels injected by the controller hence trusting the first match
	for _, apiRule := range existingAPIRules {
		if *apiRule.Spec.Service.Port == port {
			return &apiRule
		}
	}
	return nil
}

// syncInitialStatus determines the desired initial status and updates it accordingly (if conditions changed).
func (r *Reconciler) syncInitialStatus(subscription *eventingv1alpha2.Subscription) {
	if subscription.Status.Types == nil {
		subscription.Status.InitializeEventTypes()
	}

	expectedStatus := eventingv1alpha2.SubscriptionStatus{}
	expectedStatus.InitializeConditions()

	// case: conditions are already initialized and there is no change in the Ready status
	if eventingv1alpha2.ContainSameConditionTypes(subscription.Status.Conditions, expectedStatus.Conditions) &&
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
	subscription.Status.Backend.APIRuleName = ""
	subscription.Status.Backend.ExternalSink = ""
}

// getRequiredConditions removes the non-required conditions from the subscription  and adds any missing required-conditions.
func getRequiredConditions(subscriptionConditions, expectedConditions []eventingv1alpha2.Condition) []eventingv1alpha2.Condition {
	var requiredConditions []eventingv1alpha2.Condition
	expectedConditionsMap := make(map[eventingv1alpha2.ConditionType]eventingv1alpha2.Condition)
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
// So make sure you always use this method then changing a condition.
func (r *Reconciler) replaceStatusCondition(subscription *eventingv1alpha2.Subscription, condition eventingv1alpha2.Condition) bool {
	// the subscription is ready if all conditions are fulfilled
	isReady := true

	// compile list of desired conditions
	desiredConditions := make([]eventingv1alpha2.Condition, 0)
	for _, c := range subscription.Status.Conditions {
		var chosenCondition eventingv1alpha2.Condition
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
	if eventingv1alpha2.ConditionsEquals(subscription.Status.Conditions, desiredConditions) && subscription.Status.Ready == isReady {
		return false
	}

	// update the status
	subscription.Status.Conditions = desiredConditions
	subscription.Status.Ready = isReady
	return true
}

// emitConditionEvent emits a kubernetes event and sets the event type based on the Condition status.
func (r *Reconciler) emitConditionEvent(subscription *eventingv1alpha2.Subscription, condition eventingv1alpha2.Condition) {
	eventType := corev1.EventTypeNormal
	if condition.Status == corev1.ConditionFalse {
		eventType = corev1.EventTypeWarning
	}
	r.recorder.Event(subscription, eventType, string(condition.Reason), condition.Message)
}

// SetupUnmanaged creates a controller under the client control.
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged(reconcilerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return xerrors.Errorf("failed to create unmanaged controller: %v", err)
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha2.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return xerrors.Errorf("failed to watch subscriptions: %v", err)
	}

	apiRuleEventHandler :=
		&handler.EnqueueRequestForOwner{OwnerType: &eventingv1alpha2.Subscription{}, IsController: false}
	if err := ctru.Watch(&source.Kind{Type: &apigatewayv1beta1.APIRule{}}, apiRuleEventHandler); err != nil {
		return xerrors.Errorf("failed to watch APIRule: %v", err)
	}

	go func(r *Reconciler, c controller.Controller) {
		if err := c.Start(r.ctx); err != nil {
			r.namedLogger().Fatalw("Failed to start controller",
				"name", reconcilerName, "error", err)
		}
	}(r, ctru)

	return nil
}

// checkStatusActive checks if the subscription is active and if not, sets a timer for retry.
func (r *Reconciler) checkStatusActive(subscription *eventingv1alpha2.Subscription) (active bool, err error) {
	if subscription.Status.Backend.EventMeshSubscriptionStatus == nil {
		return false, nil
	}

	// check if the EMS subscription status is active
	if subscription.Status.Backend.EventMeshSubscriptionStatus.Status == string(types.SubscriptionStatusActive) {
		if len(subscription.Status.Backend.FailedActivation) > 0 {
			subscription.Status.Backend.FailedActivation = ""
		}
		return true, nil
	}

	t1 := time.Now()
	if len(subscription.Status.Backend.FailedActivation) == 0 {
		// it's the first time
		subscription.Status.Backend.FailedActivation = t1.Format(time.RFC3339)
		return false, nil
	}

	// check the timeout
	if t0, er := time.Parse(time.RFC3339, subscription.Status.Backend.FailedActivation); er != nil {
		err = er
	} else if t1.Sub(t0) > timeoutRetryActiveEmsStatus {
		err = xerrors.Errorf("timeout waiting for the subscription to be active: %s", subscription.Name)
	}

	return false, err
}

// checkLastFailedDelivery checks if LastFailedDelivery exists and if it happened after LastSuccessfulDelivery.
func (r *Reconciler) checkLastFailedDelivery(subscription *eventingv1alpha2.Subscription) (bool, error) {
	// Check if LastFailedDelivery exists.
	lastFailed := subscription.Status.Backend.EventMeshSubscriptionStatus.LastFailedDelivery
	if len(lastFailed) == 0 {
		return false, nil
	}

	// Try to parse LastFailedDelivery.
	var err error
	var lastFailedDeliveryTime time.Time
	if lastFailedDeliveryTime, err = time.Parse(time.RFC3339, lastFailed); err != nil {
		return true, xerrors.Errorf("failed to parse LastFailedDelivery: %v", err)
	}

	// Check if LastSuccessfulDelivery exists. If not, LastFailedDelivery happened last.
	lastSuccessful := subscription.Status.Backend.EventMeshSubscriptionStatus.LastSuccessfulDelivery
	if len(lastSuccessful) == 0 {
		return true, nil
	}

	// Try to parse LastSuccessfulDelivery.
	var lastSuccessfulDeliveryTime time.Time
	if lastSuccessfulDeliveryTime, err = time.Parse(time.RFC3339, lastSuccessful); err != nil {
		return true, xerrors.Errorf("failed to parse LastSuccessfulDelivery: %v", err)
	}

	return lastFailedDeliveryTime.After(lastSuccessfulDeliveryTime), nil
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}
