package apirule

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler"
)

// externalSinkScheme the scheme used for external sink.
const externalSinkScheme = "https"

// Reconciler reconciles an APIRule object.
type Reconciler struct {
	client.Client
	cache.Cache
	Log      logr.Logger
	recorder record.EventRecorder
}

// NewReconciler returns a new APIRule reconciler instance.
func NewReconciler(client client.Client, cache cache.Cache, log logr.Logger, recorder record.EventRecorder) *Reconciler {
	return &Reconciler{
		Client:   client,
		Cache:    cache,
		Log:      log,
		recorder: recorder,
	}
}

// SetupWithManager sets the controller manager object type to be the APIRule.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&apigatewayv1alpha1.APIRule{}).Complete(r)
}

// Reconcile handles the APIRule add/update/delete.
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
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

// handleAPIRuleDelete handles the APIRule delete.
func (r *Reconciler) handleAPIRuleDelete(ctx context.Context, apiRule *apigatewayv1alpha1.APIRule) (ctrl.Result, error) {
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

// handleAPIRuleAddOrUpdate handles the APIRule add/update.
func (r *Reconciler) handleAPIRuleAddOrUpdate(ctx context.Context, apiRule *apigatewayv1alpha1.APIRule) (ctrl.Result, error) {
	if !hasRelevantAPIRuleLabels(apiRule.Labels) {
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

// syncSubscriptionsStatus updates the subscription status for each item in the given subscriptions list.
func (r *Reconciler) syncSubscriptionsStatus(ctx context.Context, apiRule *apigatewayv1alpha1.APIRule, subscriptions []eventingv1alpha1.Subscription, log logr.Logger) (ctrl.Result, error) {
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

// isRelevantAPIRuleName returns true if the given name matches the APIRule name pattern
// used by the eventing-controller, otherwise returns false.
func isRelevantAPIRuleName(name string) bool {
	return strings.HasPrefix(name, reconciler.ApiRuleNamePrefix)
}

// hasRelevantAPIRuleLabels returns true if the given APIRule labels matches the APIRule Labels
// used by the eventing-controller, otherwise returns false.
func hasRelevantAPIRuleLabels(labels map[string]string) bool {
	if v, ok := labels[constants.ControllerIdentityLabelKey]; ok && v == constants.ControllerIdentityLabelValue {
		return true
	}
	return false
}

// computeAPIRuleReadyStatus returns true if all APIRule statuses is ok, otherwise returns false.
func computeAPIRuleReadyStatus(apiRule *apigatewayv1alpha1.APIRule) bool {
	if apiRule.Status.APIRuleStatus == nil || apiRule.Status.AccessRuleStatus == nil || apiRule.Status.VirtualServiceStatus == nil {
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
