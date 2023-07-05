package eventmesh

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/internal/featureflags"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/auth"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/httpclient"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

const (
	eventMeshHandlerName      = "event-mesh-handler"
	MaxSubscriptionNameLength = 50
	eventTypeSegmentsLimit    = 7
	subscriptionNameLogKey    = "eventMeshSubscriptionName"
	errorLogKey               = "error"
)

// Perform a compile time check.
var _ Backend = &EventMesh{}

type Backend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(cfg env.Config) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of messaging backend system.
	// It should return true if Kyma eventing subscription status was changed during this synchronization process.
	SyncSubscription(subscription *eventingv1alpha2.Subscription, cleaner cleaner.Cleaner, apiRule *apigatewayv1beta1.APIRule) (bool, error)

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha2.Subscription) error
}

func NewEventMesh(credentials *OAuth2ClientCredentials, mapper backendutils.NameMapper,
	logger *logger.Logger) *EventMesh {
	return &EventMesh{
		oAuth2credentials: credentials,
		logger:            logger,
		SubNameMapper:     mapper,
	}
}

type EventMesh struct {
	client            client.PublisherManager
	webhookAuth       *types.WebhookAuth
	protocolSettings  *backendutils.ProtocolSettings
	namespace         string
	eventMeshPrefix   string
	oAuth2credentials *OAuth2ClientCredentials
	SubNameMapper     backendutils.NameMapper
	logger            *logger.Logger
}

func (em *EventMesh) Initialize(cfg env.Config) error {
	if em.client == nil {
		authenticatedClient := auth.NewAuthenticatedClient(cfg)
		httpClient, err := httpclient.NewHTTPClient(cfg.BEBAPIURL, authenticatedClient)
		if err != nil {
			return err
		}
		em.client = client.NewClient(httpClient)
		em.webhookAuth = getWebHookAuth(em.oAuth2credentials)
		em.protocolSettings = &backendutils.ProtocolSettings{
			ContentMode:     &cfg.ContentMode,
			ExemptHandshake: &cfg.ExemptHandshake,
			Qos:             &cfg.Qos,
		}
		em.namespace = cfg.BEBNamespace
		em.eventMeshPrefix = cfg.EventTypePrefix
	}
	return nil
}

// getWebHookAuth returns the webhook auth config from the given env config
// or returns an error if the env config contains invalid grant type or auth type.
func getWebHookAuth(credentials *OAuth2ClientCredentials) *types.WebhookAuth {
	return &types.WebhookAuth{
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
		TokenURL:     credentials.TokenURL,
		Type:         types.AuthTypeClientCredentials,
		GrantType:    types.GrantTypeClientCredentials,
	}
}

// SyncSubscription synchronize the EV2 subscription with the EMS subscription.
// It returns true, if the EV2 subscription status was changed.
func (em *EventMesh) SyncSubscription(subscription *eventingv1alpha2.Subscription, cleaner cleaner.Cleaner,
	apiRule *apigatewayv1beta1.APIRule) (bool, error) { //nolint:funlen,gocognit
	// Format logger
	log := backendutils.LoggerWithSubscription(em.namedLogger(), subscription)

	// process event types
	typesInfo, err := em.getProcessedEventTypes(subscription, cleaner)
	if err != nil {
		log.Errorw("Failed to process types", errorLogKey, err)
		return false, err
	}

	// convert Kyma Subscription to EventMesh Subscription object
	eventMeshSub, err := backendutils.ConvertKymaSubToEventMeshSub(subscription, typesInfo, apiRule, em.webhookAuth,
		em.protocolSettings, em.namespace, em.SubNameMapper)
	if err != nil {
		log.Errorw("Failed to get Kyma subscription internal view", errorLogKey, err)
		return false, err
	}

	// check and handle if Kyma subscription or EventMesh subscription is modified
	isKymaSubModified := false
	isEventMeshSubModified := false

	// check if Kyma Subscription was modified.
	isKymaSubModified, err = em.handleKymaSubModified(eventMeshSub, subscription)
	if err != nil {
		log.Errorw("Failed to handle kyma subscription modified", errorLogKey, err)
		return false, err
	}

	// fetch the existing subscription from EventMesh.
	var eventMeshServerSub *types.Subscription
	if !isKymaSubModified {
		eventMeshServerSub, err = em.getSubscriptionIgnoreNotFound(eventMeshSub.Name)
		if err != nil {
			log.Errorw("Failed to get EventMesh subscription", subscriptionNameLogKey,
				eventMeshSub.Name, errorLogKey, err)
			return false, err
		}
	}

	// check if the EventMesh subscription was modified by EventMesh server.
	if eventMeshServerSub != nil {
		isEventMeshSubModified, err = em.handleEventMeshSubModified(eventMeshServerSub, subscription)
		if err != nil {
			log.Errorw("Failed to handle EventMesh subscription modified", errorLogKey, err)
			return false, err
		}

		// Make sure the EventMesh subscription was not deleted by checking
		// the isEventMeshSubModified flag to be false.
		if featureflags.IsEventingWebhookAuthEnabled() && !isEventMeshSubModified {
			if err = em.handleWebhookAuthChange(eventMeshSub, subscription); err != nil {
				log.Errorw("Failed to handle WebhookAuth Change", errorLogKey, err)
				return false, err
			}
		}
	}

	// create a new subscription on EventMesh server
	if isKymaSubModified || isEventMeshSubModified || eventMeshServerSub == nil {
		// create the new EMS subscription
		eventMeshServerSub, err = em.handleCreateEventMeshSub(eventMeshSub, subscription)
		if err != nil {
			log.Errorw("Failed to handle create EventMesh subscription", errorLogKey, err)
			return false, err
		}
	}

	// Update status in kyma subscription
	isUpdated, err := em.handleKymaSubStatusUpdate(eventMeshServerSub, eventMeshSub, subscription, typesInfo)
	if err != nil {
		return false, err
	}

	// check if the status is updated
	isStatusUpdated := isKymaSubModified || isEventMeshSubModified || isUpdated

	return isStatusUpdated, nil
}

// DeleteSubscription deletes the corresponding EventMesh subscription.
func (em *EventMesh) DeleteSubscription(subscription *eventingv1alpha2.Subscription) error {
	return em.deleteSubscription(em.SubNameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace))
}

// getProcessedEventTypes returns the processed types after cleaning
// and prefixing as required by EventMesh specifications.
func (em *EventMesh) getProcessedEventTypes(kymaSubscription *eventingv1alpha2.Subscription,
	cleaner cleaner.Cleaner) ([]backendutils.EventTypeInfo, error) {
	// deduplicate event types
	uniqueTypes := kymaSubscription.GetUniqueTypes()

	// process types including cleaning, appending prefixes
	result := make([]backendutils.EventTypeInfo, 0, len(uniqueTypes))
	for _, t := range uniqueTypes {
		if kymaSubscription.Spec.TypeMatching == eventingv1alpha2.TypeMatchingExact {
			// not do any processing if TypeMatching is exact.
			result = append(result, backendutils.EventTypeInfo{OriginalType: t, CleanType: t, ProcessedType: t})
			continue
		}

		// clean type and source
		cleanedSource, err := cleaner.CleanSource(kymaSubscription.Spec.Source)
		if err != nil {
			return nil, err
		}

		cleanedType, err := cleaner.CleanEventType(t)
		if err != nil {
			return nil, err
		}
		eventMeshSubject := getEventMeshSubject(cleanedSource, cleanedType, em.eventMeshPrefix)

		if isEventTypeSegmentsOverLimit(eventMeshSubject) {
			return nil, fmt.Errorf("EventMesh subject exceeds the limit of segments, "+
				"max number of segements allowed: %d", eventTypeSegmentsLimit)
		}

		result = append(result, backendutils.EventTypeInfo{OriginalType: t, CleanType: cleanedType,
			ProcessedType: eventMeshSubject})
	}

	return result, nil
}

// handleKymaSubModified checks if the Kyma subscription is modified.
// If modified, then it deletes the corresponding subscription on EventMesh and returns true.
func (em *EventMesh) handleKymaSubModified(eventMeshSub *types.Subscription, kymaSub *eventingv1alpha2.Subscription) (bool, error) {
	var (
		isModified bool
		hash       int64
		err        error
	)

	if featureflags.IsEventingWebhookAuthEnabled() {
		hash = kymaSub.Status.Backend.EventMeshLocalHash
	} else {
		hash = kymaSub.Status.Backend.Ev2hash
	}

	if isModified, err = backendutils.IsEventMeshSubModified(eventMeshSub, hash); err != nil {
		return false, err
	}

	if isModified {
		// delete subscription from EventMesh server, so it will be re-created later.
		if err := em.deleteSubscription(eventMeshSub.Name); err != nil {
			return false, fmt.Errorf("failed to delete subscription on EventMesh: %w", err)
		}
		return true, nil
	}
	return false, nil
}

// handleEventMeshSubModified checks if the EventMesh subscription is modified.
// If modified, then it deletes the subscription on EventMesh and returns true.
func (em *EventMesh) handleEventMeshSubModified(eventMeshSub *types.Subscription, kymaSub *eventingv1alpha2.Subscription) (bool, error) {
	// get the cleaned EMS subscription for comparing the hash (EventMeshHash)
	cleanedEventMeshServerSub := backendutils.GetCleanedEventMeshSubscription(eventMeshSub)
	isEventMeshServerSubModified, err := backendutils.IsEventMeshSubModified(cleanedEventMeshServerSub,
		kymaSub.Status.Backend.EventMeshHash)
	if err != nil {
		return false, err
	}

	if isEventMeshServerSubModified {
		// delete subscription from EventMesh server
		if err := em.deleteSubscription(eventMeshSub.Name); err != nil {
			return false, fmt.Errorf("failed to delete subscription on EventMesh: %w", err)
		}
		return true, nil
	}
	return false, nil
}

// handleCreateEventMeshSub handles if a new EventMesh subscription needs to be created.
func (em *EventMesh) handleCreateEventMeshSub(eventMeshSub *types.Subscription, kymaSub *eventingv1alpha2.Subscription) (*types.Subscription, error) {
	// reset the cleanEventTypes
	kymaSub.Status.InitializeEventTypes()

	// create the new EMS subscription
	eventMeshServerSub, err := em.createAndGetSubscription(eventMeshSub)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription from EventMesh: %w", err)
	}

	return eventMeshServerSub, nil
}

// handleWebhookAuthChange handles the EventMesh subscription WebhookAuth changes.
// It uses the PATCH request API provided by EventMesh to update the subscription WebhookAuth.
func (em *EventMesh) handleWebhookAuthChange(eventMeshSub *types.Subscription,
	kymaSub *eventingv1alpha2.Subscription) error {
	hash, err := backendutils.GetWebhookAuthHash(eventMeshSub.WebhookAuth)
	if err != nil {
		return fmt.Errorf("failed to get the EventMesh WebhookAuth hash: %w", err)
	}

	// skip if the WebhookAuth did not change
	if hash == kymaSub.Status.Backend.WebhookAuthHash {
		return nil
	}

	// pause subscription
	em.namedLogger().Debugf("Pausing EventMmesh subscription: %s", eventMeshSub.Name)
	state := types.State{Action: types.StateActionPause}
	resp, err := em.client.UpdateState(eventMeshSub.Name, state)
	if err != nil {
		return fmt.Errorf("failed to pause EventMesh subscription: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		em.namedLogger().Warnf("failed to pause EventMesh subscription: %s, not found", eventMeshSub.Name)
		return nil
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to pause EventMesh subscription: %w; %v",
			HTTPStatusError{StatusCode: resp.StatusCode}, resp.Message)
	}

	// update webhook auth config
	em.namedLogger().Debugf("Updating WebhookAuth config for EventMmesh subscription: %s", eventMeshSub.Name)
	updateResp, err := em.client.Update(eventMeshSub.Name, eventMeshSub.WebhookAuth)
	if err != nil {
		return fmt.Errorf("failed to update webhook auth config: %w", err)
	}
	if updateResp.StatusCode == http.StatusNotFound {
		em.namedLogger().Warnf("failed to update webhook auth config: %s, not found", eventMeshSub.Name)
		return nil
	}
	if updateResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to update webhook auth config: %w; %v",
			HTTPStatusError{StatusCode: updateResp.StatusCode}, updateResp.Message)
	}

	// resume subscription
	em.namedLogger().Debugf("Resuming EventMmesh subscription: %s", eventMeshSub.Name)
	state = types.State{Action: types.StateActionResume}
	resp, err = em.client.UpdateState(eventMeshSub.Name, state)
	if err != nil {
		return fmt.Errorf("failed to resume EventMesh subscription: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		em.namedLogger().Warnf("failed to resume EventMesh subscription: %s, not found", eventMeshSub.Name)
		return nil
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to resume EventMesh subscription: %w; %v",
			HTTPStatusError{StatusCode: resp.StatusCode}, resp.Message)
	}

	return nil
}

// handleKymaSubStatusUpdate updates the status in Kyma subscription.
// Returns true if status is updated.
func (em *EventMesh) handleKymaSubStatusUpdate(eventMeshServerSub *types.Subscription,
	eventMeshSub *types.Subscription, kymaSub *eventingv1alpha2.Subscription,
	typesInfo []backendutils.EventTypeInfo) (bool, error) {
	// Update status.types
	kymaSub.Status.Types = statusCleanEventTypes(typesInfo)

	// Update status.backend.emsTypes
	kymaSub.Status.Backend.EmsTypes = statusFinalEventTypes(typesInfo)

	// Update hashes in status
	if err := updateHashesInStatus(kymaSub, eventMeshSub, eventMeshServerSub); err != nil {
		return false, fmt.Errorf("failed to update hashes in subscription status: %w", err)
	}

	// update EventMesh sub status in kyma sub status
	statusChanged := setEmsSubscriptionStatus(kymaSub, eventMeshServerSub)

	return statusChanged, nil
}

func (em *EventMesh) getSubscriptionIgnoreNotFound(name string) (*types.Subscription, error) {
	httpStatusNotFoundError := HTTPStatusError{StatusCode: http.StatusNotFound}

	// fetch the existing subscription from EventMesh.
	eventMeshServerSub, err := em.getSubscription(name)
	if err != nil && !errors.Is(err, httpStatusNotFoundError) {
		// throw error if it is not a NotFound exception.
		return nil, err
	}
	return eventMeshServerSub, nil
}

// getSubscription fetches the subscription from EventMesh.
func (em *EventMesh) getSubscription(name string) (*types.Subscription, error) {
	eventMeshSubscription, resp, err := em.client.Get(name)
	if err != nil {
		return nil, fmt.Errorf("get subscription failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get subscription failed: %w; %v",
			HTTPStatusError{StatusCode: resp.StatusCode}, resp.Message)
	}
	return eventMeshSubscription, nil
}

// deleteSubscription deletes the subscription on EventMesh.
func (em *EventMesh) deleteSubscription(name string) error {
	em.namedLogger().Debugf("Deleting EventMmesh subscription: %s", name)
	resp, err := em.client.Delete(name)
	if err != nil {
		return fmt.Errorf("delete subscription failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete subscription failed: %w; %v",
			HTTPStatusError{StatusCode: resp.StatusCode}, resp.Message)
	}
	return nil
}

// createSubscription creates a subscription on EventMesh.
func (em *EventMesh) createSubscription(subscription *types.Subscription) error {
	createResponse, err := em.client.Create(subscription)
	if err != nil {
		return fmt.Errorf("create subscription failed: %v", err)
	}
	if createResponse.StatusCode > http.StatusAccepted && createResponse.StatusCode != http.StatusConflict {
		return fmt.Errorf("create subscription failed: %w; %v",
			HTTPStatusError{StatusCode: createResponse.StatusCode}, createResponse.Message)
	}
	return nil
}

// createSubscription creates and returns the subscription from EventMesh.
func (em *EventMesh) createAndGetSubscription(subscription *types.Subscription) (*types.Subscription, error) {
	// create a new EMS subscription
	if err := em.createSubscription(subscription); err != nil {
		return nil, err
	}

	// get the new EMS subscription
	eventMeshSub, err := em.getSubscription(subscription.Name)
	if err != nil {
		return nil, err
	}

	return eventMeshSub, nil
}

func (em *EventMesh) namedLogger() *zap.SugaredLogger {
	return em.logger.WithContext().Named(eventMeshHandlerName)
}

// SetCredentials sets the WebhookAuth credentials.
// WARNING: This functions should be used for testing purposes only.
func (em *EventMesh) SetCredentials(credentials *OAuth2ClientCredentials) {
	em.namedLogger().Warn("This logic should be used for testing purposes only")
	em.webhookAuth = getWebHookAuth(credentials)
}
