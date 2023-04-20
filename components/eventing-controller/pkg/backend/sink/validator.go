package sink

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	"golang.org/x/xerrors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
)

type Validator interface {
	Validate(subscription *v1alpha2.Subscription) error
}

// ValidatorFunc implements the Validator interface.
type ValidatorFunc func(*v1alpha2.Subscription) error

func (vf ValidatorFunc) Validate(sub *v1alpha2.Subscription) error {
	return vf(sub)
}

type defaultSinkValidator struct {
	ctx      context.Context
	client   client.Client
	recorder record.EventRecorder
}

// Perform a compile-time check.
var _ Validator = &defaultSinkValidator{}

func NewValidator(ctx context.Context, client client.Client, recorder record.EventRecorder) Validator {
	return &defaultSinkValidator{ctx: ctx, client: client, recorder: recorder}
}

func (s defaultSinkValidator) Validate(subscription *v1alpha2.Subscription) error {
	_, subDomains, err := utils.GetSinkData(subscription.Spec.Sink)
	if err != nil {
		return err
	}
	svcNs := subDomains[1]
	svcName := subDomains[0]

	// Validate svc is a cluster-local one
	if _, err := GetClusterLocalService(s.ctx, s.client, svcNs, svcName); err != nil {
		if k8serrors.IsNotFound(err) {
			events.Warn(s.recorder, subscription, events.ReasonValidationFailed, "Sink does not correspond to a valid cluster local svc")
			return xerrors.Errorf("failed to validate subscription sink URL. It is not a valid cluster local svc: %v", err)
		}

		events.Warn(s.recorder, subscription, events.ReasonValidationFailed, "Fetch cluster-local svc failed namespace %s name %s", svcNs, svcName)
		return xerrors.Errorf("failed to fetch cluster-local svc for namespace '%s' and name '%s': %v", svcNs, svcName, err)
	}

	return nil
}

func GetClusterLocalService(ctx context.Context, client client.Client, svcNs, svcName string) (*corev1.Service, error) {
	svcLookupKey := k8stypes.NamespacedName{Name: svcName, Namespace: svcNs}
	svc := &corev1.Service{}
	if err := client.Get(ctx, svcLookupKey, svc); err != nil {
		return nil, err
	}
	return svc, nil
}
