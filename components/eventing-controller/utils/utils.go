package utils

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// GetPortNumberFromURL converts string port from url.URL to uint32 port.
func GetPortNumberFromURL(u url.URL) (uint32, error) {
	port := uint32(0)
	sinkPort := u.Port()
	if sinkPort != "" {
		u64, err := strconv.ParseUint(sinkPort, 10, 32)
		if err != nil {
			return port, errors.Wrapf(err, "convert port failed %s", u.Port())
		}
		port = uint32(u64)
	}
	if port == uint32(0) {
		switch strings.ToLower(u.Scheme) {
		case "https":
			port = uint32(443)
		default:
			port = uint32(80)
		}
	}
	return port, nil
}

// Helper functions to check and remove string from a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func BoolPtr(b bool) *bool {
	return &b
}

func Int32Ptr(i int32) *int32 {
	return &i
}

func BoolPtrEqual(b1, b2 *bool) bool {
	if b1 == nil && b2 == nil {
		return true
	}

	if b1 != nil && b2 != nil {
		return *b1 == *b2
	}

	return false
}

// LoggerWithSubscription returns a logger with the given subscription details.
func LoggerWithSubscription(log *zap.SugaredLogger, subscription *eventingv1alpha1.Subscription) *zap.SugaredLogger {
	return log.With(
		"kind", subscription.GetObjectKind().GroupVersionKind().Kind,
		"version", subscription.GetGeneration(),
		"namespace", subscription.GetNamespace(),
		"name", subscription.GetName(),
	)
}
