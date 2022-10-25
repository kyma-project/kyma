//nolint:gosec
package v2

import (
	"context"
	"fmt"
	"net/url"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"

	"go.uber.org/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/pkg/errors"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

func GetExposedURLFromAPIRule(apiRule *apigatewayv1beta1.APIRule, targetURL string) (string, error) {
	// @TODO: Move this method to backend/eventmesh/utils.go once old BEB backend is depreciated
	scheme := "https://"
	path := ""

	sURL, err := url.ParseRequestURI(targetURL)
	if err != nil {
		return "", err
	}
	sURLPath := sURL.Path
	if sURL.Path == "" {
		sURLPath = "/"
	}
	for _, rule := range apiRule.Spec.Rules {
		if rule.Path == sURLPath {
			path = rule.Path
			break
		}
	}
	return fmt.Sprintf("%s%s%s", scheme, *apiRule.Spec.Host, path), nil
}

// UpdateSubscriptionStatus updates the status of all Kyma subscriptions on k8s.
func UpdateSubscriptionStatus(ctx context.Context, dClient dynamic.Interface,
	sub *eventingv1alpha2.Subscription) error {
	unstructuredObj, err := sub.ToUnstructuredSub()
	if err != nil {
		return errors.Wrap(err, "convert subscription to unstructured failed")
	}
	_, err = dClient.
		Resource(eventingv1alpha2.SubscriptionGroupVersionResource()).
		Namespace(sub.Namespace).
		UpdateStatus(ctx, unstructuredObj, metav1.UpdateOptions{})

	return err
}

// LoggerWithSubscription returns a logger with the given subscription (v1alpha2) details.
func LoggerWithSubscription(log *zap.SugaredLogger,
	subscription *eventingv1alpha2.Subscription) *zap.SugaredLogger {
	return log.With(
		"kind", subscription.GetObjectKind().GroupVersionKind().Kind,
		"version", subscription.GetGeneration(),
		"namespace", subscription.GetNamespace(),
		"name", subscription.GetName(),
	)
}
