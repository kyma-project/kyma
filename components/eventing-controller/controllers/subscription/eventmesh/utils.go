package eventmesh

import (
	"fmt"
	"net/url"
	"strings"

	apigatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// isInDeletion checks if the Subscription shall be deleted.
func isInDeletion(subscription *eventingv1alpha2.Subscription) bool {
	return !subscription.DeletionTimestamp.IsZero()
}

// isFinalizerSet checks if a finalizer is set on the Subscription which belongs to this controller.
func isFinalizerSet(sub *eventingv1alpha2.Subscription) bool {
	// Check if finalizer is already set
	for _, finalizer := range sub.ObjectMeta.Finalizers {
		if finalizer == eventingv1alpha2.Finalizer {
			return true
		}
	}
	return false
}

// addFinalizer adds eventingv1alpha2 finalizer to the Subscription.
func addFinalizer(sub *eventingv1alpha2.Subscription, logger *zap.SugaredLogger) error {
	sub.ObjectMeta.Finalizers = append(sub.ObjectMeta.Finalizers, eventingv1alpha2.Finalizer)
	logger.Debug("Added finalizer to subscription")
	return nil
}

// removeFinalizer removes eventingv1alpha2 finalizer from the Subscription.
func removeFinalizer(sub *eventingv1alpha2.Subscription) {
	var finalizers []string

	// Build finalizer list without the one the controller owns
	for _, finalizer := range sub.ObjectMeta.Finalizers {
		if finalizer == eventingv1alpha2.Finalizer {
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	sub.ObjectMeta.Finalizers = finalizers
}

// getSvcNsAndName returns namespace and name of the svc from the URL.
func getSvcNsAndName(url string) (string, string, error) {
	parts := strings.Split(url, ".")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid sinkURL for cluster local svc: %s", url)
	}
	return parts[1], parts[0], nil
}

// computeAPIRuleReadyStatus returns true if all APIRule statuses is ok, otherwise returns false.
func computeAPIRuleReadyStatus(apiRule *apigatewayv1beta1.APIRule) bool {
	if apiRule == nil || apiRule.Status.APIRuleStatus == nil || apiRule.Status.AccessRuleStatus == nil || apiRule.Status.VirtualServiceStatus == nil {
		return false
	}
	apiRuleStatus := apiRule.Status.APIRuleStatus.Code == apigatewayv1beta1.StatusOK
	accessRuleStatus := apiRule.Status.AccessRuleStatus.Code == apigatewayv1beta1.StatusOK
	virtualServiceStatus := apiRule.Status.VirtualServiceStatus.Code == apigatewayv1beta1.StatusOK
	return apiRuleStatus && accessRuleStatus && virtualServiceStatus
}

// setSubscriptionStatusExternalSink sets the subscription external sink based on the given APIRule service host.
func setSubscriptionStatusExternalSink(subscription *eventingv1alpha2.Subscription, apiRule *apigatewayv1beta1.APIRule) error {
	if apiRule.Spec.Service == nil {
		return errors.Errorf("APIRule has nil service")
	}

	if apiRule.Spec.Host == nil {
		return errors.Errorf("APIRule has nil host")
	}

	u, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		return xerrors.Errorf("invalid sink for subscription namespace=%s name=%s : %v", subscription.Namespace, subscription.Name, err)
	}

	path := u.Path
	if u.Path == "" {
		path = "/"
	}

	subscription.Status.Backend.ExternalSink = fmt.Sprintf("%s://%s%s", externalSinkScheme, *apiRule.Spec.Host, path)

	return nil
}
