package controllers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

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

const ExternalSinkScheme = "https"

// SetupWithManager TODO ...
func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&apigatewayv1alpha1.APIRule{}).Complete(r)
}

// Reconcile TODO ...
func (r *APIRuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	apiRule := &apigatewayv1alpha1.APIRule{}
	if err := r.Client.Get(ctx, req.NamespacedName, apiRule); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !isRelevantAPIRule(apiRule) {
		return ctrl.Result{}, nil
	}

	return r.syncAPIRuleSubscriptionsStatus(apiRule, ctx)
}

// syncAPIRuleSubscriptionsStatus TODO ...
func (r *APIRuleReconciler) syncAPIRuleSubscriptionsStatus(apiRule *apigatewayv1alpha1.APIRule, ctx context.Context) (ctrl.Result, error) {
	// format log
	log := r.Log.WithValues(
		"kind", apiRule.GetObjectKind().GroupVersionKind().Kind,
		"name", apiRule.GetName(),
		"namespace", apiRule.GetNamespace(),
		"version", apiRule.GetGeneration(),
	)

	apiRuleReady := computeAPIRuleReadyStatus(apiRule)

	// update the statuses of the APIRule dependant subscriptions
	for _, ownerRef := range apiRule.ObjectMeta.OwnerReferences {
		subscription := &eventingv1alpha1.Subscription{}
		lookupKey := k8stypes.NamespacedName{Name: ownerRef.Name, Namespace: apiRule.Namespace}

		if err := r.Cache.Get(ctx, lookupKey, subscription); err != nil {
			log.Error(err, "Subscription not found", "Name", ownerRef.Name)
			return ctrl.Result{}, err
		}

		// Won't do anything if the subscription is marked for deletion
		if subscription.DeletionTimestamp != nil {
			continue
		}

		// work on a copy
		subscriptionCopy := subscription.DeepCopy()

		// set subscription status APIRule status condition
		subscriptionCopy.Status.SetConditionAPIRuleStatus(apiRuleReady)

		subscriptionCopy.Status.ExternalSink = ""

		// set subscription status externalSink if the APIRule status is ready
		if apiRuleReady {
			if err := setSubscriptionStatusExternalSink(subscriptionCopy, apiRule); err != nil {
				log.Error(err, "Failed to set Subscription status externalSink", "Subscription", subscription.Name, "Namespace", subscription.Namespace)
				return ctrl.Result{}, err
			}
		}

		if err := r.Client.Status().Update(ctx, subscriptionCopy); err != nil {
			log.Error(err, "Failed to update Subscription status", "Subscription", subscription.Name, "Namespace", subscription.Namespace)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// isRelevantAPIRule TODO ...
func isRelevantAPIRule(apiRule *apigatewayv1alpha1.APIRule) bool {
	if v, ok := apiRule.Labels[ControllerIdentityLabelKey]; ok && v == ControllerIdentityLabelValue {
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
