package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"net/url"
	"strings"

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

func NewAPIRuleReconciler(client client.Client, cache cache.Cache, log logr.Logger, recorder record.EventRecorder, scheme *runtime.Scheme) *APIRuleReconciler {
	return &APIRuleReconciler{Client: client, Cache: cache, Log: log, recorder: recorder, Scheme: scheme}
}

func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&apigatewayv1alpha1.APIRule{}).Complete(r)
}

func (r *APIRuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	// get the APIRule
	apiRule := &apigatewayv1alpha1.APIRule{}
	if err := r.Client.Get(ctx, req.NamespacedName, apiRule); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// work on a copy
	apiRuleCopy := apiRule.DeepCopy()

	// format log
	log := r.Log.WithValues(
		"kind", apiRuleCopy.GetObjectKind().GroupVersionKind().Kind,
		"name", apiRuleCopy.GetName(),
		"namespace", apiRuleCopy.GetNamespace(),
		"version", apiRuleCopy.GetGeneration(),
	)

	// filter out APIRules not created by this controller
	if !isRelevantAPIRule(apiRule) {
		return ctrl.Result{}, nil
	}

	return r.syncAPIRuleSubscriptionsStatus(apiRuleCopy, ctx, log)
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
	apiRuleStatus := apiRule.Status.APIRuleStatus.Code == apigatewayv1alpha1.StatusOK
	accessRuleStatus := apiRule.Status.AccessRuleStatus.Code == apigatewayv1alpha1.StatusOK
	virtualServiceStatus := apiRule.Status.VirtualServiceStatus.Code == apigatewayv1alpha1.StatusOK
	return apiRuleStatus && accessRuleStatus && virtualServiceStatus
}

// syncAPIRuleSubscriptionsStatus TODO ...
func (r *APIRuleReconciler) syncAPIRuleSubscriptionsStatus(apiRule *apigatewayv1alpha1.APIRule, ctx context.Context, log logr.Logger) (ctrl.Result, error) {
	// compute the APIRule status
	apiRuleReady := computeAPIRuleReadyStatus(apiRule)

	// update the statuses of the APIRule dependant subscriptions
	for _, ownerRef := range apiRule.ObjectMeta.OwnerReferences {
		subscription := &eventingv1alpha1.Subscription{}
		lookupKey := k8stypes.NamespacedName{Name: ownerRef.Name, Namespace: apiRule.Namespace}

		if err := r.Cache.Get(ctx, lookupKey, subscription); err != nil {
			log.Error(err, "Subscription not found", "Name", ownerRef.Name)

			// TODO discuss return err or continue
			return ctrl.Result{}, err
		}

		// work on a copy
		subscriptionCopy := subscription.DeepCopy()

		// set subscription externalSink
		if err := setSubscriptionExternalSink(subscriptionCopy); err != nil {
			log.Error(err, "Failed to set Subscription externalSink")

			// TODO discuss return err or continue
			return ctrl.Result{}, err
		}

		subscriptionCopy.Status.APIRuleReady = apiRuleReady
		if err := r.Status().Update(ctx, subscriptionCopy); err != nil {
			log.Error(err, "Failed to update the Subscription status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// setSubscriptionExternalSink TODO ...
func setSubscriptionExternalSink(subscription *eventingv1alpha1.Subscription) error {
	u, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(fmt.Sprintf("subscription: [%s/%s] has invalid sink", subscription.Name, subscription.Namespace)))
	}

	parts := strings.Split(u.Host, ".")
	if len(parts) < 5 || !strings.HasSuffix(strings.Split(u.Host, ":")[0], ClusterLocalURLSuffix) {
		return fmt.Errorf("subscription: [%s/%s] has invalid cluster local sink", subscription.Name, subscription.Namespace)
	}

	// TODO
	//  - get the domain from env
	//  - move logic to a common pkg
	subscription.Status.ExternalSink = fmt.Sprintf("%s://%s.%s.%s%s", u.Scheme, parts[0], parts[1], "domain", u.Path)

	return nil
}
