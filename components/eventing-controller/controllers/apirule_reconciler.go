package controllers

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
)

const ExternalSinkScheme = "https"

// APIRuleReconciler reconciles an APIRule object
type APIRuleReconciler struct {
	client.Client
	cache.Cache
	Log      logr.Logger
	recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// NewAPIRuleReconciler TODO ...
func NewAPIRuleReconciler(client client.Client,
	cache cache.Cache,
	log logr.Logger,
	recorder record.EventRecorder,
	scheme *runtime.Scheme) *APIRuleReconciler {
	return &APIRuleReconciler{
		Client:   client,
		Cache:    cache,
		Log:      log,
		recorder: recorder,
		Scheme:   scheme,
	}
}

// SetupWithManager TODO ...
func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&apigatewayv1alpha1.APIRule{}).Complete(r)
}

// Reconcile TODO ...
func (r *APIRuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	apiRule := &apigatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	// handle delete APIRule
	if err := r.Client.Get(ctx, req.NamespacedName, apiRule); err != nil {
		return r.handleAPIRuleDelete(ctx, apiRule)
	}

	// handle add/update APIRule
	return r.handleAPIRuleAddOrUpdate(ctx, apiRule)
}

// handleAPIRuleDelete TODO ...
func (r *APIRuleReconciler) handleAPIRuleDelete(ctx context.Context, apiRule *apigatewayv1alpha1.APIRule) (ctrl.Result, error) {
	if !isRelevantAPIRuleName(apiRule.Name) {
		return ctrl.Result{}, nil
	}

	// format log
	log := r.Log.WithValues(
		"kind", "APIRule",
		"name", apiRule.ObjectMeta.Name,
		"namespace", apiRule.ObjectMeta.Namespace,
		"mode", "Delete",
	)

	// list all namespace subscriptions
	namespaceSubscriptions := &eventingv1alpha1.SubscriptionList{}
	if err := r.Client.List(ctx, namespaceSubscriptions, client.InNamespace(apiRule.Namespace)); err != nil {
		log.Error(err, "Failed to list namespace Subscriptions")
		return ctrl.Result{}, err
	}

	// filter namespace subscriptions that are relevant to the current APIRule
	apiRuleSubscriptions := make([]eventingv1alpha1.Subscription, 0, len(namespaceSubscriptions.Items))
	for _, subscription := range namespaceSubscriptions.Items {
		// skip if the subscription is marked for deletion
		if subscription.DeletionTimestamp != nil {
			continue
		}

		// skip if APIRule name does not match
		if subscription.Status.APIRuleName != apiRule.Name {
			continue
		}

		apiRuleSubscriptions = append(apiRuleSubscriptions, subscription)
	}

	return r.syncSubscriptionsStatus(ctx, apiRule, apiRuleSubscriptions, log)
}

// handleAPIRuleAddOrUpdate TODO ...
func (r *APIRuleReconciler) handleAPIRuleAddOrUpdate(ctx context.Context, apiRule *apigatewayv1alpha1.APIRule) (ctrl.Result, error) {
	if !isRelevantAPIRule(apiRule) {
		return ctrl.Result{}, nil
	}

	// format log
	log := r.Log.WithValues(
		"kind", "APIRule",
		"name", apiRule.ObjectMeta.Name,
		"namespace", apiRule.ObjectMeta.Namespace,
		"version", apiRule.GetGeneration(),
		"mode", "Add/Update",
	)

	// get subscriptions that are relevant to the current APIRule
	apiRuleSubscriptions := make([]eventingv1alpha1.Subscription, 0, len(apiRule.ObjectMeta.OwnerReferences))
	for _, ownerRef := range apiRule.ObjectMeta.OwnerReferences {
		subscription := &eventingv1alpha1.Subscription{}
		lookupKey := k8stypes.NamespacedName{Name: ownerRef.Name, Namespace: apiRule.Namespace}

		if err := r.Client.Get(ctx, lookupKey, subscription); err != nil {
			if k8serrors.IsNotFound(err) {
				// The subscription is deleted so nothing to do
				return ctrl.Result{}, nil
			}
			log.Error(err, "Subscription not found", "Name", ownerRef.Name)
			return ctrl.Result{}, err
		}

		// skip if the subscription is marked for deletion
		if subscription.DeletionTimestamp != nil {
			continue
		}

		apiRuleSubscriptions = append(apiRuleSubscriptions, *subscription)
	}

	return r.syncSubscriptionsStatus(ctx, apiRule, apiRuleSubscriptions, log)
}

// syncSubscriptionsStatus TODO ...
func (r *APIRuleReconciler) syncSubscriptionsStatus(ctx context.Context, apiRule *apigatewayv1alpha1.APIRule, subscriptions []eventingv1alpha1.Subscription, log logr.Logger) (ctrl.Result, error) {
	apiRuleReady := computeAPIRuleReadyStatus(apiRule)

	// update the statuses of the APIRule dependant subscriptions
	for _, subscription := range subscriptions {
		// skip if the subscription is marked for deletion
		if subscription.DeletionTimestamp != nil {
			continue
		}

		// work on a copy
		subscriptionCopy := subscription.DeepCopy()

		// set subscription initial status
		subscriptionCopy.Status.ExternalSink = ""
		subscriptionCopy.Status.SetConditionAPIRuleStatus(apiRuleReady)

		// set subscription status externalSink if the APIRule status is ready
		if apiRuleReady {
			if err := setSubscriptionStatusExternalSink(subscriptionCopy, apiRule); err != nil {
				log.Error(err, "Failed to set Subscription status externalSink", "Subscription", subscription.Name, "Namespace", subscription.Namespace)
				return ctrl.Result{}, err
			}
		}

		// skip updating the status if nothing changed
		if subscription.Status.GetConditionAPIRuleStatus() == apiRuleReady && subscription.Status.ExternalSink == subscriptionCopy.Status.ExternalSink {
			continue
		}

		if err := r.Client.Status().Update(ctx, subscriptionCopy); err != nil {
			log.Error(err, "Failed to update Subscription status", "Subscription", subscription.Name, "Namespace", subscription.Namespace)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// isRelevantAPIRuleName TODO ...
func isRelevantAPIRuleName(name string) bool {
	return strings.HasPrefix(name, SinkURLPrefix)
}

// isRelevantAPIRule TODO ...
func isRelevantAPIRule(apiRule *apigatewayv1alpha1.APIRule) bool {
	if v, ok := apiRule.Labels[constants.ControllerIdentityLabelKey]; ok && v == constants.ControllerIdentityLabelValue {
		return true
	}
	return false
}

// computeAPIRuleReadyStatus TODO ...
func computeAPIRuleReadyStatus(apiRule *apigatewayv1alpha1.APIRule) bool {
	if apiRule.Status.APIRuleStatus == nil || apiRule.Status.AccessRuleStatus == nil || apiRule.Status.VirtualServiceStatus == nil {
		return false
	}
	apiRuleStatus := apiRule.Status.APIRuleStatus.Code == apigatewayv1alpha1.StatusOK
	accessRuleStatus := apiRule.Status.AccessRuleStatus.Code == apigatewayv1alpha1.StatusOK
	virtualServiceStatus := apiRule.Status.VirtualServiceStatus.Code == apigatewayv1alpha1.StatusOK
	return apiRuleStatus && accessRuleStatus && virtualServiceStatus
}

// setSubscriptionStatusExternalSink TODO ...
func setSubscriptionStatusExternalSink(subscription *eventingv1alpha1.Subscription, apiRule *apigatewayv1alpha1.APIRule) error {
	u, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(fmt.Sprintf("subscription: [%s/%s] has invalid sink", subscription.Name, subscription.Namespace)))
	}

	path := u.Path
	if u.Path == "" {
		path = "/"
	}

	subscription.Status.ExternalSink = fmt.Sprintf("%s://%s%s", ExternalSinkScheme, *apiRule.Spec.Service.Host, path)

	return nil
}
