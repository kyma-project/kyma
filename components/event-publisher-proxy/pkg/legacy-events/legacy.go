package legacy

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	apiv1 "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
)

var (
	isValidEventTypeVersion = regexp.MustCompile(AllowedEventTypeVersionChars).MatchString
	isValidEventID          = regexp.MustCompile(AllowedEventIDChars).MatchString
	// eventTypePrefixFormat is driven by BEB specification.
	// An EventType must have at least 4 segments separated by dots in the form of:
	// <domainNamespace>.<businessObjectName>.<operation>.<version>
	eventTypePrefixFormat = "%s.%s.%s.%s"
)

const (
	requestBodyTooLargeErrorMessage = "http: request body too large"
	eventTypeVersionExtensionKey    = "eventtypeversion"
)

type Transformer struct {
	bebNamespace      string
	eventTypePrefix   string
	applicationLister *application.Lister
}

func NewTransformer(bebNamespace string, eventTypePrefix string, applicationLister *application.Lister) *Transformer {
	return &Transformer{
		bebNamespace:      bebNamespace,
		eventTypePrefix:   eventTypePrefix,
		applicationLister: applicationLister,
	}
}

// CheckParameters validates the parameters in the request and sends error responses if found invalid
func (t Transformer) checkParameters(parameters *apiv1.PublishEventParametersV1) (response *apiv1.PublishEventResponses) {
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

func (t Transformer) TransformLegacyRequestsToCE(writer http.ResponseWriter, request *http.Request) *cev2event.Event {
	// parse request body to PublishRequestV1
	if request.Body == nil || request.ContentLength == 0 {
		resp := ErrorResponseBadRequest(ErrorMessageBadPayload)
		writeJSONResponse(writer, resp)
		return nil
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
		return nil
	}
	log.Errorf("Transforming event %s", string(parameters.PublishrequestV1.EventType))
	// validate the PublishRequestV1 for missing / incoherent values
	checkResp := t.checkParameters(parameters)
	if checkResp.Error != nil {
		writeJSONResponse(writer, checkResp)
		return nil
	}

	// clean the application name form non-alphanumeric characters
	applicationName := ParseApplicationNameFromPath(request.URL.Path)
	app, err := t.applicationLister.Get(applicationName)
	if err != nil {
		response := ErrorResponse(http.StatusInternalServerError, err)
		writeJSONResponse(writer, response)
		return nil
	}
	applicationName = application.CleanName(app)

	event, err := t.convertPublishRequestToCloudEvent(applicationName, parameters)
	if err != nil {
		response := ErrorResponse(http.StatusInternalServerError, err)
		writeJSONResponse(writer, response)
		return nil
	}

	return event
}

func (t Transformer) TransformsCEResponseToLegacyResponse(writer http.ResponseWriter, statusCode int, event *cev2event.Event, msg string) {
	response := &apiv1.PublishEventResponses{}
	// Fail
	if !is2XXStatusCode(statusCode) {
		response.Error = &apiv1.Error{
			Status:  statusCode,
			Message: msg,
		}
		writeJSONResponse(writer, response)
		return
	}

	// Success
	response.Ok = &apiv1.PublishResponse{EventID: event.ID()}
	writeJSONResponse(writer, response)
	return
}

// convertPublishRequestToCloudEvent converts the given publish request to a CloudEvent.
func (t Transformer) convertPublishRequestToCloudEvent(appName string, publishRequest *apiv1.PublishEventParametersV1) (*cev2event.Event, error) {
	if !application.IsCleanName(appName) {
		return nil, errors.New("application name should be cleaned from none-alphanumeric characters")
	}

	event := cev2event.New(cev2event.CloudEventsVersionV1)

	evTime, err := time.Parse(time.RFC3339, publishRequest.PublishrequestV1.EventTime)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse time from the external publish request")
	}
	event.SetTime(evTime)

	if err := event.SetData(ContentTypeApplicationJSON, publishRequest.PublishrequestV1.Data); err != nil {
		return nil, errors.Wrap(err, "failed to set data to CE data field")
	}

	// set the event id from the request if it is available
	// otherwise generate a new one
	if len(publishRequest.PublishrequestV1.EventID) > 0 {
		event.SetID(publishRequest.PublishrequestV1.EventID)
	} else {
		event.SetID(uuid.New().String())
	}

	eventType := formatEventType4BEB(t.eventTypePrefix, appName, publishRequest.PublishrequestV1.EventType, publishRequest.PublishrequestV1.EventTypeVersion)
	event.SetSource(t.bebNamespace)
	event.SetType(eventType)
	event.SetExtension(eventTypeVersionExtensionKey, publishRequest.PublishrequestV1.EventTypeVersion)
	event.SetDataContentType(ContentTypeApplicationJSON)
	return &event, nil
}
