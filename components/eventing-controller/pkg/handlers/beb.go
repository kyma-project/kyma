package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/auth"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/httpclient"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

const (
	bebHandlerName               = "beb-handler"
	MaxBEBSubscriptionNameLength = 50
	BEBSubscriptionNameLogKey    = "bebSubscriptionName"
	ErrorLogKey                  = "error"
)

// compile time check
var _ BEBBackend = &BEB{}

type BEBBackend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(cfg env.Config) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of messaging backend system.
	// It should return true if Kyma eventing subscription status was changed during this synchronization process.
	SyncSubscription(subscription *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner, apiRule *apigatewayv1alpha1.APIRule) (bool, error)

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha1.Subscription) error
}

type OAuth2ClientCredentials struct {
	ClientID     string
	ClientSecret string
}

func NewBEB(credentials *OAuth2ClientCredentials, mapper NameMapper, logger *logger.Logger) *BEB {
	return &BEB{
		OAth2credentials: credentials,
		logger:           logger,
		SubNameMapper:    mapper,
	}
}

type BEB struct {
	Client           client.PublisherManager
	WebhookAuth      *types.WebhookAuth
	ProtocolSettings *eventingv1alpha1.ProtocolSettings
	Namespace        string
	OAth2credentials *OAuth2ClientCredentials
	SubNameMapper    NameMapper
	logger           *logger.Logger
}

type BEBResponse struct {
	StatusCode int
	Error      error
}

func (b *BEB) Initialize(cfg env.Config) error {
	if b.Client == nil {
		authenticatedClient := auth.NewAuthenticatedClient(cfg)
		httpClient, err := httpclient.NewHTTPClient(cfg.BEBAPIURL, authenticatedClient)
		if err != nil {
			return err
		}
		b.Client = client.NewClient(httpClient)
		b.WebhookAuth = getWebHookAuth(cfg, b.OAth2credentials)
		b.ProtocolSettings = &eventingv1alpha1.ProtocolSettings{
			ContentMode:     &cfg.ContentMode,
			ExemptHandshake: &cfg.ExemptHandshake,
			Qos:             &cfg.Qos,
		}
		b.Namespace = cfg.BEBNamespace
	}
	return nil
}

// getWebHookAuth returns the webhook auth config from the given env config
// or returns an error if the env config contains invalid grant type or auth type.
func getWebHookAuth(cfg env.Config, credentials *OAuth2ClientCredentials) *types.WebhookAuth {
	return &types.WebhookAuth{
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
		TokenURL:     cfg.WebhookTokenEndpoint,
		Type:         types.AuthTypeClientCredentials,
		GrantType:    types.GrantTypeClientCredentials,
	}
}

// SyncSubscription synchronize the EV2 subscription with the EMS subscription. It returns true, if the EV2 subscription status was changed
func (b *BEB) SyncSubscription(subscription *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner, apiRule *apigatewayv1alpha1.APIRule) (bool, error) {
	// Format logger
	log := utils.LoggerWithSubscription(b.namedLogger(), subscription)

	// get the internal view for the ev2 subscription
	var statusChanged = false
	sEv2, err := getInternalView4Ev2(subscription, apiRule, b.WebhookAuth, b.ProtocolSettings, b.Namespace, b.SubNameMapper)
	if err != nil {
		log.Errorw("get Kyma subscription internal view failed", ErrorLogKey, err)
		return false, err
	}

	newEv2Hash, err := getHash(sEv2)
	if err != nil {
		log.Errorw("get Kyma subscription hash failed", ErrorLogKey, err)
		return false, err
	}

	var bebSubscription *types.Subscription
	// check the hash values for ev2 and EMS
	if newEv2Hash != subscription.Status.Ev2hash {
		// delete & create a new EMS subscription
		var newEMSHash int64
		bebSubscription, newEMSHash, err = b.deleteCreateAndHashSubscription(sEv2, cleaner, log)
		if err != nil {
			return false, err
		}
		subscription.Status.Ev2hash = newEv2Hash
		subscription.Status.Emshash = newEMSHash
		statusChanged = true
	} else {
		// check if EMS subscription is the same as in the past
		bebSubscription, err = b.getSubscription(sEv2.Name)
		if err != nil {
			log.Errorw("get BEB subscription failed", BEBSubscriptionNameLogKey, sEv2.Name, ErrorLogKey, err)
			httpStatusNotFoundError := errors.New(strconv.Itoa(http.StatusNotFound))
			if errors.As(err, &httpStatusNotFoundError) {
				log.Infow("Recreate the BEB subscription", BEBSubscriptionNameLogKey, sEv2.Name)
				bebSubscription, err = b.createAndGetSubscription(sEv2, cleaner, log)
				if err != nil {
					return false, err
				}
			} else {
				return false, err
			}
		}
		// get the internal view for the EMS subscription
		sEms := getInternalView4Ems(bebSubscription)
		newEmsHash, err := getHash(sEms)
		if err != nil {
			log.Errorw("get BEB subscription hash failed", ErrorLogKey, err)
			return false, err
		}
		if newEmsHash != subscription.Status.Emshash {
			// delete & create a new EMS subscription
			bebSubscription, newEmsHash, err = b.deleteCreateAndHashSubscription(sEv2, cleaner, log)
			if err != nil {
				return false, err
			}
			subscription.Status.Emshash = newEmsHash
			statusChanged = true
		}
	}
	// set the status of bebSubscription in ev2Subscription
	statusChanged = b.setEmsSubscriptionStatus(subscription, bebSubscription) || statusChanged

	// get the clean event types
	subscription.Status.CleanEventTypes = statusCleanEventTypes(bebSubscription.Events)

	return statusChanged, nil
}

// DeleteSubscription deletes the corresponding EMS subscription
func (b *BEB) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	return b.deleteSubscription(b.SubNameMapper.MapSubscriptionName(subscription))
}

func (b *BEB) deleteCreateAndHashSubscription(subscription *types.Subscription, cleaner eventtype.Cleaner, log *zap.SugaredLogger) (*types.Subscription, int64, error) {
	log = log.With(BEBSubscriptionNameLogKey, subscription.Name)
	// delete EMS subscription
	if err := b.deleteSubscription(subscription.Name); err != nil {
		log.Errorw("delete BEB subscription failed", ErrorLogKey, err)
		return nil, 0, err
	}

	// clean the application name segment in the subscription event-types from none-alphanumeric characters
	if err := cleanEventTypes(subscription, cleaner); err != nil {
		log.Errorw("clean application name in the subscription event-types failed", ErrorLogKey, err)
		return nil, 0, err
	}

	// create a new EMS subscription
	if err := b.createSubscription(subscription, log); err != nil {
		log.Errorw("create BEB subscription failed", ErrorLogKey, err)
		return nil, 0, err
	}

	// get the new EMS subscription
	bebSubscription, err := b.getSubscription(subscription.Name)
	if err != nil {
		log.Errorw("get BEB subscription failed", ErrorLogKey, err)
		return nil, 0, err
	}

	// get the new hash
	sEMS := getInternalView4Ems(bebSubscription)
	if err != nil {
		log.Errorw("get BEB subscription internal view failed", ErrorLogKey, err)
	}
	newEmsHash, err := getHash(sEMS)
	if err != nil {
		log.Errorw("get BEB subscription hash failed", ErrorLogKey, err)
		return nil, 0, err
	}

	return bebSubscription, newEmsHash, nil
}

func (b *BEB) createAndGetSubscription(subscription *types.Subscription, cleaner eventtype.Cleaner, log *zap.SugaredLogger) (*types.Subscription, error) {
	// clean the application name segment in the subscription event-types from none-alphanumeric characters
	if err := cleanEventTypes(subscription, cleaner); err != nil {
		log.Errorw("clean application name in the subscription event-types failed", ErrorLogKey, err)
		return nil, err
	}

	log = log.With(BEBSubscriptionNameLogKey, subscription.Name)
	// create a new EMS subscription
	if err := b.createSubscription(subscription, log); err != nil {
		log.Errorw("create BEB subscription failed", ErrorLogKey, err)
		return nil, err
	}

	// get the new EMS subscription
	bebSubscription, err := b.getSubscription(subscription.Name)
	if err != nil {
		log.Errorw("get BEB subscription failed", ErrorLogKey, err)
		return nil, err
	}

	return bebSubscription, nil
}

// cleanEventTypes cleans the application name segment in the subscription event-types from none-alphanumeric characters
// note: the given subscription instance will be updated with the cleaned event-types
func cleanEventTypes(subscription *types.Subscription, cleaner eventtype.Cleaner) error {
	events := make(types.Events, 0, len(subscription.Events))
	for _, event := range subscription.Events {
		eventType, err := cleaner.Clean(event.Type)
		if err != nil {
			return err
		}
		events = append(events, types.Event{Source: event.Source, Type: eventType})
	}
	subscription.Events = events
	return nil
}

// setEmsSubscriptionStatus sets the status of bebSubscription in ev2Subscription
func (b *BEB) setEmsSubscriptionStatus(subscription *eventingv1alpha1.Subscription, bebSubscription *types.Subscription) bool {
	var statusChanged = false
	if subscription.Status.EmsSubscriptionStatus.SubscriptionStatus != string(bebSubscription.SubscriptionStatus) {
		subscription.Status.EmsSubscriptionStatus.SubscriptionStatus = string(bebSubscription.SubscriptionStatus)
		statusChanged = true
	}
	if subscription.Status.EmsSubscriptionStatus.SubscriptionStatusReason != bebSubscription.SubscriptionStatusReason {
		subscription.Status.EmsSubscriptionStatus.SubscriptionStatusReason = bebSubscription.SubscriptionStatusReason
		statusChanged = true
	}
	if subscription.Status.EmsSubscriptionStatus.LastSuccessfulDelivery != bebSubscription.LastSuccessfulDelivery {
		subscription.Status.EmsSubscriptionStatus.LastSuccessfulDelivery = bebSubscription.LastSuccessfulDelivery
		statusChanged = true
	}
	if subscription.Status.EmsSubscriptionStatus.LastFailedDelivery != bebSubscription.LastFailedDelivery {
		subscription.Status.EmsSubscriptionStatus.LastFailedDelivery = bebSubscription.LastFailedDelivery
		statusChanged = true
	}
	if subscription.Status.EmsSubscriptionStatus.LastFailedDeliveryReason != bebSubscription.LastFailedDeliveryReason {
		subscription.Status.EmsSubscriptionStatus.LastFailedDeliveryReason = bebSubscription.LastFailedDeliveryReason
		statusChanged = true
	}
	return statusChanged
}

func (b *BEB) getSubscription(name string) (*types.Subscription, error) {
	bebSubscription, resp, err := b.Client.Get(name)
	if err != nil {
		return nil, fmt.Errorf("get subscription failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get subscription failed: %w; %v", errors.New(strconv.Itoa(resp.StatusCode)), resp.Message)
	}
	return bebSubscription, nil
}

func statusCleanEventTypes(events types.Events) []string {
	var cleanEventTypes []string
	for _, e := range events {
		cleanEventTypes = append(cleanEventTypes, e.Type)
	}
	return cleanEventTypes
}

func (b *BEB) deleteSubscription(name string) error {
	resp, err := b.Client.Delete(name)
	if err != nil {
		return fmt.Errorf("delete subscription failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete subscription failed: %w; %v", errors.New(strconv.Itoa(resp.StatusCode)), resp.Message)
	}
	return nil
}

func (b *BEB) createSubscription(subscription *types.Subscription, log *zap.SugaredLogger) error {
	createResponse, err := b.Client.Create(subscription)
	if err != nil {
		return fmt.Errorf("create subscription failed: %v", err)
	}
	if createResponse.StatusCode > http.StatusAccepted && createResponse.StatusCode != http.StatusConflict {
		return fmt.Errorf("create subscription failed: %w; %v", errors.New(strconv.Itoa(createResponse.StatusCode)), createResponse.Message)
	}
	log.Debug("create subscription succeeded")
	return nil
}

func (b *BEB) namedLogger() *zap.SugaredLogger {
	return b.logger.WithContext().Named(bebHandlerName)
}
