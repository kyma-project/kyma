package legacy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	apiv1 "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	isValidEventTypeVersion = regexp.MustCompile(AllowedEventTypeVersionChars).MatchString
	isValidEventID          = regexp.MustCompile(AllowedEventIDChars).MatchString
	// eventTypePrefixFormat is driven by BEB specification.
	// An eventtype must have atleast 4 segments separated by dots in the form of:
	// <domainNamespace>.<businessObjectName>.<operation>.<version>
	eventTypePrefixFormat = "%s.%s.%s.%s"
)

type Transformer struct {
	bebNamespace    string
	eventTypePrefix string
}

func NewTransformer(bebNamespace string, eventTypePrefix string) *Transformer {
	return &Transformer{
		bebNamespace:    bebNamespace,
		eventTypePrefix: eventTypePrefix,
	}
}

const (
	requestBodyTooLargeErrorMessage = "http: request body too large"
	eventTypeVersionExtensionKey    = "eventtypeversion"
)

func (t Transformer) TransformsLegacyRequestsToCE(writer http.ResponseWriter, request *http.Request) *cev2event.Event {

	// parse request body to PublishRequestV1
	if request.Body == nil || request.ContentLength == 0 {
		resp := ErrorResponseBadRequest(ErrorMessageBadPayload)
		writeJSONResponse(writer, resp)
		return nil
	}
	appName := parseApplicationNameFromPath(request.URL.Path)
	var err error
	parameters := &apiv1.PublishEventParametersV1{}
	decoder := json.NewDecoder(request.Body)
	err = decoder.Decode(&parameters.PublishrequestV1)
	if err != nil {
		var resp *apiv1.PublishEventResponses
		if err.Error() == requestBodyTooLargeErrorMessage {
			resp = ErrorResponseRequestBodyTooLarge(err.Error())
		} else {
			resp = ErrorResponseBadRequest(err.Error())
		}
		writeJSONResponse(writer, resp)
		return nil
	}

	// validate the PublishRequestV1 for missing / incoherent values
	checkResp := checkParameters(parameters)
	if checkResp.Error != nil {
		writeJSONResponse(writer, checkResp)
		return nil
	}

	response := &apiv1.PublishEventResponses{}

	event, err := t.convertPublishRequestToCloudEvent(appName, parameters)
	if err != nil {
		response.Error = &apiv1.Error{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
		writeJSONResponse(writer, response)
		return nil
	}

	return event
}

func parseApplicationNameFromPath(path string) string {
	// Assumption: Clients(e.g. ApplicationConnectivity validator which has a flag for
	// the path(https://github.com/kyma-project/kyma/blob/master/components/application-connectivity-validator/cmd/applicationconnectivityvalidator/applicationconnectivityvalidator.go#L49)
	// using this endpoint must be sending request to path /:application/v1/events
	// Hence it should be safe to return 0th index as the application name
	pathSegments := make([]string, 0)
	for _, segment := range strings.Split(path, "/") {
		if strings.TrimSpace(segment) != "" {
			pathSegments = append(pathSegments, segment)
		}
	}
	return pathSegments[0]
}

func checkParameters(parameters *apiv1.PublishEventParametersV1) (response *apiv1.PublishEventResponses) {
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

// is2XXStatusCode checks whether status code is a 2XX status code
func is2XXStatusCode(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

func writeJSONResponse(w http.ResponseWriter, resp *apiv1.PublishEventResponses) {
	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", ContentTypeApplicationJSON)

	if resp.Error != nil {
		w.WriteHeader(resp.Error.Status)
		_ = encoder.Encode(resp.Error)
		return
	}

	if resp.Ok != nil {
		_ = encoder.Encode(resp.Ok)
		return
	}

	log.Errorf("received an empty response")
}

// convertPublishRequestToCloudEvent converts the given publish request to a CloudEvent.
func (t Transformer) convertPublishRequestToCloudEvent(appName string, publishRequest *apiv1.PublishEventParametersV1) (*cev2event.Event, error) {
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

// formatEventType4BEB format eventType as per BEB spec
func formatEventType4BEB(eventTypePrefix, app, eventType, version string) string {
	eventType4BEB := fmt.Sprintf(eventTypePrefixFormat, eventTypePrefix, app, eventType, version)
	return strings.ReplaceAll(eventType4BEB, "-", ".")
}
