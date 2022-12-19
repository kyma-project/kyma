package legacy

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	ce "github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/builder"
	apiv1 "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/tracing"
)

var (
	isValidEventTypeVersion = regexp.MustCompile(AllowedEventTypeVersionChars).MatchString
	isValidEventID          = regexp.MustCompile(AllowedEventIDChars).MatchString
)

const (
	requestBodyTooLargeErrorMessage = "http: request body too large"
	eventTypeVersionExtensionKey    = "eventtypeversion"
)

type RequestToCETransformer interface {
	TransformLegacyRequestsToEvent(http.ResponseWriter, *http.Request) (*builder.Event, string)
	TransformsCEResponseToLegacyResponse(http.ResponseWriter, int, *ce.Event, string)
}

type Transformer struct {
	namespace         string
	eventTypePrefix   string
	applicationLister *application.Lister
}

func NewTransformer(namespace string, eventTypePrefix string, applicationLister *application.Lister) *Transformer {
	return &Transformer{
		namespace:         namespace,
		eventTypePrefix:   eventTypePrefix,
		applicationLister: applicationLister,
	}
}

// CheckParameters validates the parameters in the request and sends error responses if found invalid
func (t *Transformer) checkParameters(parameters *apiv1.PublishEventParametersV1) (response *apiv1.PublishEventResponses) {
	if parameters == nil {
		return ErrorResponseBadRequest(ErrorMessageBadPayload)
	}
	if len(parameters.PublishrequestV1.EventType) == 0 {
		return ErrorResponseMissingFieldEventType()
	}
	if len(parameters.PublishrequestV1.EventTypeVersion) == 0 {
		return ErrorResponseMissingFieldEventTypeVersion()
	}
	if !isValidEventTypeVersion(parameters.PublishrequestV1.EventTypeVersion) {
		return ErrorResponseWrongEventTypeVersion()
	}
	if len(parameters.PublishrequestV1.EventTime) == 0 {

		return ErrorResponseMissingFieldEventTime()
	}
	if _, err := time.Parse(time.RFC3339, parameters.PublishrequestV1.EventTime); err != nil {
		return ErrorResponseWrongEventTime()
	}
	if len(parameters.PublishrequestV1.EventID) > 0 && !isValidEventID(parameters.PublishrequestV1.EventID) {
		return ErrorResponseWrongEventID()
	}
	if parameters.PublishrequestV1.Data == nil {
		return ErrorResponseMissingFieldData()
	}
	if d, ok := (parameters.PublishrequestV1.Data).(string); ok && len(d) == 0 {
		return ErrorResponseMissingFieldData()
	}
	// OK
	return &apiv1.PublishEventResponses{}
}

// TransformLegacyRequestsToEvent transforms the http request containing a legacy event
// to a Event from the given request. It's second return type is a string that holds
// the original Type without any cleanup.
func (t *Transformer) TransformLegacyRequestsToEvent(writer http.ResponseWriter, request *http.Request) (*builder.Event, string) {
	// Parse request body to PublishRequestV1.
	if request.Body == nil || request.ContentLength == 0 {
		resp := ErrorResponseBadRequest(ErrorMessageBadPayload)
		writeJSONResponse(writer, resp)
		return nil, ""
	}

	parameters := &apiv1.PublishEventParametersV1{}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&parameters.PublishrequestV1); err != nil {
		var resp *apiv1.PublishEventResponses
		if err.Error() == requestBodyTooLargeErrorMessage {
			resp = ErrorResponseRequestBodyTooLarge(err.Error())
		} else {
			resp = ErrorResponseBadRequest(err.Error())
		}
		writeJSONResponse(writer, resp)
		return nil, ""
	}

	// Validate the PublishRequestV1 for missing or incoherent values.
	checkResp := t.checkParameters(parameters)
	if checkResp.Error != nil {
		writeJSONResponse(writer, checkResp)
		return nil, ""
	}

	// Clean the application name from non-alphanumeric characters.
	originalAppName := ParseApplicationNameFromPath(request.URL.Path)
	appName := originalAppName
	if appObj, err := t.applicationLister.Get(appName); err == nil {
		// Handle existing applications.
		appName = application.GetCleanTypeOrName(appObj)
	} else {
		// Handle non-existing applications.
		appName = application.GetCleanName(appName)
	}

	event, err := t.convertPublishRequestToEvent(appName, parameters)
	if err != nil {
		response := ErrorResponse(http.StatusInternalServerError, err)
		writeJSONResponse(writer, response)
		return nil, ""
	}

	// Add tracing context to cloud events.
	tracing.AddTracingContextToCEExtensions(request.Header, &event.Event)

	// Prepare the original event-type without cleanup.
	originalEventType := formatEventType(t.eventTypePrefix, originalAppName, parameters.PublishrequestV1.EventType, parameters.PublishrequestV1.EventTypeVersion)

	return event, originalEventType
}

func (t *Transformer) TransformsCEResponseToLegacyResponse(writer http.ResponseWriter, statusCode int, event *ce.Event, msg string) {
	response := &apiv1.PublishEventResponses{}
	// Fail.
	if !is2XXStatusCode(statusCode) {
		response.Error = &apiv1.Error{
			Status:  statusCode,
			Message: msg,
		}
		writeJSONResponse(writer, response)
		return
	}

	// Success.
	response.Ok = &apiv1.PublishResponse{EventID: event.ID()}
	writeJSONResponse(writer, response)
}

// convertPublishRequestToCloudEvent converts the given publish request to a Event (which itself is a wrapper
// around a CloudEvent).
func (t *Transformer) convertPublishRequestToEvent(appName string, publishRequest *apiv1.PublishEventParametersV1) (*builder.Event, error) {
	if !application.IsCleanName(appName) {
		return nil, errors.New("application name should be cleaned from none-alphanumeric characters")
	}

	cloudEvent := ce.New(ce.CloudEventsVersionV1)

	evTime, err := time.Parse(time.RFC3339, publishRequest.PublishrequestV1.EventTime)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse time from the external publish request")
	}
	cloudEvent.SetTime(evTime)

	if err := cloudEvent.SetData(internal.ContentTypeApplicationJSON, publishRequest.PublishrequestV1.Data); err != nil {
		return nil, errors.Wrap(err, "failed to set data to CloudEvent data field")
	}

	// Set the event id from the request if it is available
	// or otherwise generate a new one.
	if len(publishRequest.PublishrequestV1.EventID) > 0 {
		cloudEvent.SetID(publishRequest.PublishrequestV1.EventID)
	} else {
		cloudEvent.SetID(uuid.New().String())
	}

	// Create a new Event from the CloudEvent.
	event, err := builder.NewEvent(
		builder.WithCloudEvent(&cloudEvent),
		builder.WithPrefix(t.eventTypePrefix),
		builder.WithApp(appName),
		builder.WithName(publishRequest.PublishrequestV1.EventType),
		builder.WithVersion(publishRequest.PublishrequestV1.EventTypeVersion),
		builder.WithRemoveNonAlphanumericsFromType(),
		builder.WithEventSource(t.namespace),
		builder.WithEventExtension(eventTypeVersionExtensionKey, publishRequest.PublishrequestV1.EventTypeVersion),
		builder.WithEventDataContentType(internal.ContentTypeApplicationJSON),
	)
	if err != nil {
		return nil, err
	}
	return event, nil
}
