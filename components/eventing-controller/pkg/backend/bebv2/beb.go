package bebv2

import (
	"fmt"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

type HTTPStatusError struct {
	StatusCode int
}

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("%v", e.StatusCode)
}

func (e *HTTPStatusError) Is(target error) bool {
	t, ok := target.(*HTTPStatusError)
	if !ok {
		return false
	}
	return e.StatusCode == t.StatusCode
}

// Perform a compile time check.
var _ Backend = &BEB{}

type Backend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(cfg env.Config) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of messaging backend system.
	// It should return true if Kyma eventing subscription status was changed during this synchronization process.
	SyncSubscription(subscription *eventingv1alpha2.Subscription, cleaner eventtype.Cleaner, apiRule *apigatewayv1beta1.APIRule) (bool, error)

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha2.Subscription) error
}

type OAuth2ClientCredentials struct {
	ClientID     string
	ClientSecret string
}

func NewBEB(credentials *OAuth2ClientCredentials, mapper backendutils.NameMapper, logger *logger.Logger) *BEB {
	return &BEB{
		OAth2credentials: credentials,
		logger:           logger,
		SubNameMapper:    mapper,
	}
}

type BEB struct {
	Client           client.PublisherManager
	WebhookAuth      *types.WebhookAuth
	ProtocolSettings *eventingv1alpha2.ProtocolSettings
	Namespace        string
	OAth2credentials *OAuth2ClientCredentials
	SubNameMapper    backendutils.NameMapper
	logger           *logger.Logger
}

type Response struct {
	StatusCode int
	Error      error
}

func (b *BEB) Initialize(cfg env.Config) error {
	return nil
}

// SyncSubscription synchronize the EV2 subscription with the EMS subscription. It returns true, if the EV2 subscription status was changed.
func (b *BEB) SyncSubscription(_ *eventingv1alpha2.Subscription, _ eventtype.Cleaner, _ *apigatewayv1beta1.APIRule) (bool, error) {
	return false, nil
}

// DeleteSubscription deletes the corresponding EMS subscription.
func (b *BEB) DeleteSubscription(subscription *eventingv1alpha2.Subscription) error {
	return nil
}
