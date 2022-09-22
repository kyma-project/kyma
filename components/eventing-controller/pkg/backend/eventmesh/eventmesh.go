package eventmesh

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendbebv1 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/auth"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/httpclient"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

const (
	eventMeshHandlerName               = "event-mesh-handler"
	MaxEventMeshSubscriptionNameLength = 50
	SubscriptionNameLogKey             = "eventMeshSubscriptionName"
	ErrorLogKey                        = "error"
)

// Perform a compile time check.
var _ Backend = &EventMesh{}

type Backend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(cfg env.Config) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of messaging backend system.
	// It should return true if Kyma eventing subscription status was changed during this synchronization process.
	SyncSubscription(subscription *eventingv1alpha2.Subscription, cleaner eventtype.Cleaner, apiRule *apigatewayv1beta1.APIRule) (bool, error)

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha2.Subscription) error
}

func NewEventMesh(credentials *backendbebv1.OAuth2ClientCredentials, mapper backendutils.NameMapper, logger *logger.Logger) *EventMesh {
	return &EventMesh{
		OAth2credentials: credentials,
		logger:           logger,
		SubNameMapper:    mapper,
	}
}

type EventMesh struct {
	Client           client.PublisherManager
	WebhookAuth      *types.WebhookAuth
	ProtocolSettings *eventingv1alpha2.ProtocolSettings
	Namespace        string
	OAth2credentials *backendbebv1.OAuth2ClientCredentials
	SubNameMapper    backendutils.NameMapper
	logger           *logger.Logger
}

func (em *EventMesh) Initialize(cfg env.Config) error {
	if em.Client == nil {
		authenticatedClient := auth.NewAuthenticatedClient(cfg)
		httpClient, err := httpclient.NewHTTPClient(cfg.BEBAPIURL, authenticatedClient)
		if err != nil {
			return err
		}
		em.Client = client.NewClient(httpClient)
		em.WebhookAuth = getWebHookAuth(cfg, em.OAth2credentials)
		em.ProtocolSettings = &eventingv1alpha2.ProtocolSettings{
			ContentMode:     &cfg.ContentMode,
			ExemptHandshake: &cfg.ExemptHandshake,
			Qos:             &cfg.Qos,
		}
		em.Namespace = cfg.BEBNamespace
	}
	return nil
}

// getWebHookAuth returns the webhook auth config from the given env config
// or returns an error if the env config contains invalid grant type or auth type.
func getWebHookAuth(cfg env.Config, credentials *backendbebv1.OAuth2ClientCredentials) *types.WebhookAuth {
	return &types.WebhookAuth{
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
		TokenURL:     cfg.WebhookTokenEndpoint,
		Type:         types.AuthTypeClientCredentials,
		GrantType:    types.GrantTypeClientCredentials,
	}
}

// SyncSubscription synchronize the EV2 subscription with the EMS subscription. It returns true, if the EV2 subscription status was changed.
func (em *EventMesh) SyncSubscription(subscription *eventingv1alpha2.Subscription, cleaner eventtype.Cleaner, apiRule *apigatewayv1beta1.APIRule) (bool, error) {
	// Format logger
	log := backendutils.LoggerWithSubscriptionV1AlphaV2(em.namedLogger(), subscription)

	// define flag to track if status is updated
	var statusChanged = false

	// process event types
	typesInfo, err := em.GetProcessedEventTypes(subscription, cleaner)
	if err != nil {
		log.Errorw("Failed to process types", ErrorLogKey, err)
		return false, err
	}

	// convert Kyma Sub to EventMesh sub
	eventMeshSub, err := backendutils.ConvertKymaSubToEventMeshSub(subscription, typesInfo, apiRule, em.WebhookAuth, em.ProtocolSettings, em.Namespace, em.SubNameMapper)
	if err != nil {
		log.Errorw("Failed to get Kyma subscription internal view", ErrorLogKey, err)
		return false, err
	}

	// First, check if Kyma Subscription was modified
	// compare hashes
	// if hashes are different --> delete and recreate EventMesh subs

	isEventMeshSubModified, err := backendutils.IsEventMeshSubModified(eventMeshSub, subscription.Status.Backend.Ev2hash)
	if err != nil {
		return false, err
	}

	if isEventMeshSubModified {
		// delete subscription from EventMesh server
		if err := em.deleteSubscription(subscription.Name); err != nil {
			log.Errorw("Failed to delete subscription on EventMesh", ErrorLogKey, err)
			return false, err
		}
	}

	var eventMeshServerSub *types.Subscription
	if !isEventMeshSubModified {
		eventMeshServerSub, err = em.getSubscription(eventMeshSub.Name)
		if err != nil {
			// throw error if it is not a NotFound exception.
			httpStatusNotFoundError := backendbebv1.HTTPStatusError{StatusCode: http.StatusNotFound}
			if !errors.Is(err, httpStatusNotFoundError) {
				log.Errorw("Failed to get BEB subscription", SubscriptionNameLogKey, eventMeshSub.Name, ErrorLogKey, err)
				return false, err
			}
		}
	}

	if eventMeshServerSub != nil {
		// get the internal view for the EMS subscription
		cleanedEventMeshServerSub := backendutils.GetCleanedEventMeshSubscription(eventMeshServerSub)
		isEventMeshServerSubModified, err := backendutils.IsEventMeshSubModified(cleanedEventMeshServerSub, subscription.Status.Backend.Emshash)
		if err != nil {
			return false, err
		}

		if isEventMeshServerSubModified {
			// delete subscription from EventMesh server
			if err := em.deleteSubscription(subscription.Name); err != nil {
				log.Errorw("Failed to delete subscription on EventMesh", ErrorLogKey, err)
				return false, err
			}
			// remove the eventMeshServerSub local instance
			eventMeshServerSub = nil
		}
	}

	// check if we should create subscription on EventMesh server
	if eventMeshServerSub == nil {
		// reset the cleanEventTypes
		subscription.Status.InitializeCleanEventTypes()

		// create the new EMS subscription
		eventMeshServerSub, err = em.createAndGetSubscription(eventMeshSub)
		if err != nil {
			log.Errorw("Failed to get subscription from EventMesh", ErrorLogKey, err)
			return false, err
		}
		// update flag for status update
		statusChanged = true
	}

	// Update status.types
	subscription.Status.Types = statusCleanEventTypes(typesInfo)

	// Update status.backend.types
	// @TODO: check where to put this information in status, the EventMesh subject
	// would be different from cleaned type because we add prefix
	// for testing, putting it in backend.types
	subscription.Status.Backend.Types = statusFinalEventTypes(typesInfo)

	// Update hashes in status
	if err = em.updateHashesInStatus(subscription, eventMeshSub, eventMeshServerSub); err != nil {
		log.Errorw("Failed to update hashes in subscription status", ErrorLogKey, err)
		return false, err
	}

	// update EventMesh sub status in kyma sub status
	statusChanged = em.setEmsSubscriptionStatus(subscription, eventMeshServerSub) || statusChanged

	return statusChanged, nil
}

// DeleteSubscription deletes the corresponding EventMesh subscription.
func (em *EventMesh) DeleteSubscription(subscription *eventingv1alpha2.Subscription) error {
	return em.deleteSubscription(em.SubNameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace))
}

// GetProcessedEventTypes returns the processed types after cleaning and prefixing as required by EventMesh specifications.
func (em *EventMesh) GetProcessedEventTypes(kymaSubscription *eventingv1alpha2.Subscription, cleaner eventtype.Cleaner) ([]backendutils.EventTypeInfo, error) {
	// deduplicate event types
	uniqueTypes := kymaSubscription.GetUniqueTypes()

	// process types including cleaning, appending prefixes
	result := make([]backendutils.EventTypeInfo, 0, len(uniqueTypes))
	for _, t := range uniqueTypes {
		if kymaSubscription.Spec.TypeMatching == eventingv1alpha2.EXACT {
			// not do any processing if TypeMatching is exact.
			result = append(result, backendutils.EventTypeInfo{OriginalType: t, CleanType: t, ProcessedType: t})
			continue
		}

		// clean type and source
		cleanedSource := kymaSubscription.Spec.Source // @TODO: clean
		cleanedType, err := cleaner.Clean(t)
		if err != nil {
			return nil, err
		}
		eventMeshSubject := em.GetEventMeshSubject(cleanedSource, cleanedType)
		result = append(result, backendutils.EventTypeInfo{OriginalType: t, CleanType: cleanedType, ProcessedType: eventMeshSubject})
	}

	return result, nil
}

// GetEventMeshSubject appends the prefix to subject.
func (em *EventMesh) GetEventMeshSubject(source, subject string) string {
	// @TODO: Update it to use event type prefix and source
	return fmt.Sprintf("%s.%s.%s", "sap.kyma.custom", source, subject)
}

func (em *EventMesh) updateHashesInStatus(kymaSubscription *eventingv1alpha2.Subscription, eventMeshLocalSubscription *types.Subscription, eventMeshServerSubscription *types.Subscription) error {
	if err := em.setEventMeshLocalSubHashInStatus(kymaSubscription, eventMeshLocalSubscription); err != nil {
		return err
	}
	if err := em.setEventMeshServerSubHashInStatus(kymaSubscription, eventMeshServerSubscription); err != nil {
		return err
	}
	return nil
}

// setEventMeshLocalSubHashInStatus sets the hash for EventMesh local sub in Kyma Sub status.
func (em *EventMesh) setEventMeshLocalSubHashInStatus(kymaSubscription *eventingv1alpha2.Subscription, eventMeshSubscription *types.Subscription) error {
	// generate hash
	newHash, err := backendutils.GetHash(eventMeshSubscription)
	if err != nil {
		return err
	}

	// set hash in status
	kymaSubscription.Status.Backend.Ev2hash = newHash
	return nil
}

// setEventMeshServerSubHashInStatus sets the hash for EventMesh local sub in Kyma Sub status.
func (em *EventMesh) setEventMeshServerSubHashInStatus(kymaSubscription *eventingv1alpha2.Subscription, eventMeshSubscription *types.Subscription) error {
	// clean up the server sub object from extra info
	cleanedEventMeshSub := backendutils.GetCleanedEventMeshSubscription(eventMeshSubscription)
	// generate hash
	newHash, err := backendutils.GetHash(cleanedEventMeshSub)
	if err != nil {
		return err
	}

	// set hash in status
	kymaSubscription.Status.Backend.Emshash = newHash
	return nil
}

// setEmsSubscriptionStatus sets the status of bebSubscription in ev2Subscription.
func (em *EventMesh) setEmsSubscriptionStatus(subscription *eventingv1alpha2.Subscription, eventMeshSubscription *types.Subscription) bool {
	var statusChanged = false
	if subscription.Status.Backend.EmsSubscriptionStatus == nil {
		subscription.Status.Backend.EmsSubscriptionStatus = &eventingv1alpha2.EmsSubscriptionStatus{}
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.Status != string(eventMeshSubscription.SubscriptionStatus) {
		subscription.Status.Backend.EmsSubscriptionStatus.Status = string(eventMeshSubscription.SubscriptionStatus)
		statusChanged = true
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.StatusReason != eventMeshSubscription.SubscriptionStatusReason {
		subscription.Status.Backend.EmsSubscriptionStatus.StatusReason = eventMeshSubscription.SubscriptionStatusReason
		statusChanged = true
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.LastSuccessfulDelivery != eventMeshSubscription.LastSuccessfulDelivery {
		subscription.Status.Backend.EmsSubscriptionStatus.LastSuccessfulDelivery = eventMeshSubscription.LastSuccessfulDelivery
		statusChanged = true
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.LastFailedDelivery != eventMeshSubscription.LastFailedDelivery {
		subscription.Status.Backend.EmsSubscriptionStatus.LastFailedDelivery = eventMeshSubscription.LastFailedDelivery
		statusChanged = true
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.LastFailedDeliveryReason != eventMeshSubscription.LastFailedDeliveryReason {
		subscription.Status.Backend.EmsSubscriptionStatus.LastFailedDeliveryReason = eventMeshSubscription.LastFailedDeliveryReason
		statusChanged = true
	}
	return statusChanged
}

func (em *EventMesh) getSubscription(name string) (*types.Subscription, error) {
	bebSubscription, resp, err := em.Client.Get(name)
	if err != nil {
		return nil, fmt.Errorf("get subscription failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get subscription failed: %w; %v", backendbebv1.HTTPStatusError{StatusCode: resp.StatusCode}, resp.Message)
	}
	return bebSubscription, nil
}

func statusCleanEventTypes(typeInfos []backendutils.EventTypeInfo) []eventingv1alpha2.EventType {
	var cleanEventTypes []eventingv1alpha2.EventType
	for _, i := range typeInfos {
		cleanEventTypes = append(cleanEventTypes, eventingv1alpha2.EventType{OriginalType: i.OriginalType, CleanType: i.CleanType})
	}
	return cleanEventTypes
}

func statusFinalEventTypes(typeInfos []backendutils.EventTypeInfo) []eventingv1alpha2.JetStreamTypes {
	var finalEventTypes []eventingv1alpha2.JetStreamTypes
	for _, i := range typeInfos {
		finalEventTypes = append(finalEventTypes, eventingv1alpha2.JetStreamTypes{OriginalType: i.OriginalType, ConsumerName: i.ProcessedType})
	}
	return finalEventTypes
}

func (em *EventMesh) deleteSubscription(name string) error {
	resp, err := em.Client.Delete(name)
	if err != nil {
		return fmt.Errorf("delete subscription failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete subscription failed: %w; %v", backendbebv1.HTTPStatusError{StatusCode: resp.StatusCode}, resp.Message)
	}
	return nil
}

func (em *EventMesh) createSubscription(subscription *types.Subscription) error {
	createResponse, err := em.Client.Create(subscription)
	if err != nil {
		return fmt.Errorf("create subscription failed: %v", err)
	}
	if createResponse.StatusCode > http.StatusAccepted && createResponse.StatusCode != http.StatusConflict {
		return fmt.Errorf("create subscription failed: %w; %v", backendbebv1.HTTPStatusError{StatusCode: createResponse.StatusCode}, createResponse.Message)
	}
	return nil
}

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
