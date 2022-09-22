//nolint:gosec
package utils

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

type EventTypeInfo struct {
	OriginalType  string
	CleanType     string
	ProcessedType string
}

// NameMapper is used to map Kyma-specific resource names to their corresponding name on other
// (external) systems, e.g. on different eventing backends, the same Kyma subscription name
// could map to a different name.
type NameMapper interface {
	MapSubscriptionName(subscriptionName, subscriptionNamespace string) string
}

// produces a name+hash which is not longer than maxLength. If necessary, shortens name, not the hash.
// Requires maxLength >= len(hash).
func shortenNameAndAppendHash(name string, hash string, maxLength int) string {
	if len(hash) > maxLength {
		// This shouldn't happen!
		panic(fmt.Sprintf("max name length (%d) used for EventMesh subscription mapper is not large enough to hold the hash (%s)", maxLength, hash))
	}
	maxNameLen := maxLength - len(hash)
	// keep the first maxNameLen characters of the name
	if maxNameLen <= 0 {
		return hash
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}
	return name + hash
}

func getExposedURLFromAPIRule(apiRule *apigatewayv1beta1.APIRule, targetUrl string) (string, error) {
	scheme := "https://"
	path := ""

	sURL, err := url.ParseRequestURI(targetUrl)
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

// UpdateSubscriptionV1Alpha2Status updates the status of all Kyma subscriptions on k8s
func UpdateSubscriptionV1Alpha2Status(ctx context.Context, dClient dynamic.Interface, sub *eventingv1alpha2.Subscription) error {
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

// LoggerWithSubscriptionV1AlphaV2 returns a logger with the given subscription (v1alphav2) details.
func LoggerWithSubscriptionV1AlphaV2(log *zap.SugaredLogger, subscription *eventingv1alpha2.Subscription) *zap.SugaredLogger {
	return log.With(
		"kind", subscription.GetObjectKind().GroupVersionKind().Kind,
		"version", subscription.GetGeneration(),
		"namespace", subscription.GetNamespace(),
		"name", subscription.GetName(),
	)
}
