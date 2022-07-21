package sink

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	clusterLocalURLSuffix = "svc.cluster.local"
	MissingSchemeErrMsg   = "sink URL scheme should be 'http' or 'https'"
)

type Validator interface {
	Validate(subscription *v1alpha1.Subscription) error
}

// ValidatorFunc implements the Validator interface.
type ValidatorFunc func(*v1alpha1.Subscription) error

func (vf ValidatorFunc) Validate(sub *v1alpha1.Subscription) error {
	return vf(sub)
}

type defaultSinkValidator struct {
	ctx      context.Context
	client   client.Client
	recorder record.EventRecorder
	logger   *logger.Logger
}

// compile-time check
var _ Validator = &defaultSinkValidator{}

func NewValidator(ctx context.Context, client client.Client, recorder record.EventRecorder, logger *logger.Logger) Validator {
	return &defaultSinkValidator{ctx: ctx, client: client, recorder: recorder, logger: logger}
}

func (s defaultSinkValidator) Validate(subscription *v1alpha1.Subscription) error {
	if !isValidScheme(subscription.Spec.Sink) {
		events.Warn(s.recorder, subscription, events.ReasonValidationFailed, "Sink URL scheme should be HTTP or HTTPS: %s", subscription.Spec.Sink)
		return fmt.Errorf(MissingSchemeErrMsg)
	}

	sURL, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		events.Warn(s.recorder, subscription, events.ReasonValidationFailed, "Not able to parse Sink URL with error: %s", err.Error())
		return fmt.Errorf("not able to parse sink url with error: %s", err.Error())
	}

	// Validate sink URL is a cluster local URL
	trimmedHost := strings.Split(sURL.Host, ":")[0]
	if !strings.HasSuffix(trimmedHost, clusterLocalURLSuffix) {
		events.Warn(s.recorder, subscription, events.ReasonValidationFailed, "Sink does not contain suffix: %s", clusterLocalURLSuffix)
		return fmt.Errorf("sink does not contain suffix: %s in the URL", clusterLocalURLSuffix)
	}

	// we expected a sink in the format "service.namespace.svc.cluster.local"
	subDomains := strings.Split(trimmedHost, ".")
	if len(subDomains) != 5 {
		events.Warn(s.recorder, subscription, events.ReasonValidationFailed, "Sink should contain 5 sub-domains: %s", trimmedHost)
		return fmt.Errorf("sink should contain 5 sub-domains: %s", trimmedHost)
	}

	// Assumption: Subscription CR and Subscriber should be deployed in the same namespace
	svcNs := subDomains[1]
	if subscription.Namespace != svcNs {
		events.Warn(s.recorder, subscription, events.ReasonValidationFailed, "natsNamespace of subscription: %s and the subscriber: %s are different", subscription.Namespace, svcNs)
		return fmt.Errorf("namespace of subscription: %s and the namespace of subscriber: %s are different", subscription.Namespace, svcNs)
	}

	// Validate svc is a cluster-local one
	svcName := subDomains[0]
	if _, err := getClusterLocalService(s.ctx, s.client, svcNs, svcName); err != nil {
		if k8serrors.IsNotFound(err) {
			events.Warn(s.recorder, subscription, events.ReasonValidationFailed, "Sink does not correspond to a valid cluster local svc")
			return fmt.Errorf("sink is not a valid cluster local svc, failed with error: %w", err)
		}

		events.Warn(s.recorder, subscription, events.ReasonValidationFailed, "Fetch cluster-local svc failed namespace %s name %s", svcNs, svcName)
		return fmt.Errorf("fetch cluster-local svc failed namespace:%s name:%s with error: %w", svcNs, svcName, err)
	}

	return nil
}

func getClusterLocalService(ctx context.Context, client client.Client, svcNs, svcName string) (*corev1.Service, error) {
	svcLookupKey := k8stypes.NamespacedName{Name: svcName, Namespace: svcNs}
	svc := &corev1.Service{}
	if err := client.Get(ctx, svcLookupKey, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

// isValidScheme returns true if the sink scheme is http or https, otherwise returns false.
func isValidScheme(sink string) bool {
	return strings.HasPrefix(sink, "http://") || strings.HasPrefix(sink, "https://")
}
