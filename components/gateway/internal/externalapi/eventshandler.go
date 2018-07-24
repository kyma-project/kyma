package externalapi

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/kyma-project/kyma/components/gateway/internal/events/api"
	"github.com/kyma-project/kyma/components/gateway/internal/events/bus"
	"github.com/kyma-project/kyma/components/gateway/internal/events/shared"
	log "github.com/sirupsen/logrus"
)

var (
	isValidEventTypeVersion = regexp.MustCompile(shared.AllowedEventTypeVersionChars).MatchString
	isValidEventId          = regexp.MustCompile(shared.AllowedEventIdChars).MatchString
	traceHeaderKeys         = []string{"x-request-id", "x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags", "x-ot-span-context"}
)

func NewEventsHandler() http.Handler {
	return http.HandlerFunc(handleEvents)
}

// EventsHandler handles "/v1/events" requests
func handleEvents(w http.ResponseWriter, req *http.Request) {
	if req.Body == nil || req.ContentLength == 0 {
		resp := shared.ErrorResponseBadRequest(shared.ErrorMessageBadPayload)
		writeJsonResponse(w, resp)
		return
	}
	var err error
	parameters := &api.PublishEventParameters{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&parameters.Publishrequest)
	if err != nil {
		resp := shared.ErrorResponseBadRequest(err.Error())
		writeJsonResponse(w, resp)
		return
	}
	resp := &api.PublishEventResponses{}

	traceHeaders := getTraceHeaders(req)

	err = handleEvent(parameters, resp, traceHeaders)
	if err == nil {
		if resp.Ok != nil || resp.Error != nil {
			writeJsonResponse(w, resp)
			return
		}
		log.Println("Cannot process event")
		http.Error(w, "Cannot process event", http.StatusInternalServerError)
		return
	}
	log.Printf("Internal Error: %s\n", err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

var handleEvent = func(publishRequest *api.PublishEventParameters, publishResponse *api.PublishEventResponses, traceHeaders *map[string]string) (err error) {
	checkResp := checkParameters(publishRequest)
	if checkResp.Error != nil {
		publishResponse.Error = checkResp.Error
		return
	}
	// add source to the incoming request
	sendRequest, err := bus.AddSource(publishRequest)
	if err != nil {
		return err
	}
	// send the event
	sendEventResponse, err := bus.SendEvent(sendRequest, traceHeaders)
	if err != nil {
		return err
	}
	publishResponse.Ok = sendEventResponse.Ok
	publishResponse.Error = sendEventResponse.Error
	return err
}

func checkParameters(parameters *api.PublishEventParameters) (response *api.PublishEventResponses) {
	if parameters == nil {
		return shared.ErrorResponseBadRequest(shared.ErrorMessageBadPayload)
	}
	if len(parameters.Publishrequest.EventType) == 0 {
		return shared.ErrorResponseMissingFieldEventType()
	}
	if len(parameters.Publishrequest.EventTypeVersion) == 0 {
		return shared.ErrorResponseMissingFieldEventTypeVersion()
	}
	if !isValidEventTypeVersion(parameters.Publishrequest.EventTypeVersion) {
		return shared.ErrorResponseWrongEventTypeVersion()
	}
	if len(parameters.Publishrequest.EventTime) == 0 {
		return shared.ErrorResponseMissingFieldEventTime()
	}
	if _, err := time.Parse(time.RFC3339, parameters.Publishrequest.EventTime); err != nil {
		return shared.ErrorResponseWrongEventTime(err)
	}
	if len(parameters.Publishrequest.EventId) > 0 && !isValidEventId(parameters.Publishrequest.EventId) {
		return shared.ErrorResponseWrongEventId()
	}
	if parameters.Publishrequest.Data == nil {
		return shared.ErrorResponseMissingFieldData()
	} else if d, ok := (parameters.Publishrequest.Data).(string); ok && d == "" {
		return shared.ErrorResponseMissingFieldData()
	}
	// OK
	return &api.PublishEventResponses{Ok: nil, Error: nil}
}

func writeJsonResponse(w http.ResponseWriter, resp *api.PublishEventResponses) {
	encoder := json.NewEncoder(w)
	if resp.Error != nil {
		w.WriteHeader(resp.Error.Status)
		encoder.Encode(resp.Error)
	} else {
		encoder.Encode(resp.Ok)
	}
	return
}

func getTraceHeaders(req *http.Request) *map[string]string {
	traceHeaders := make(map[string]string)
	for _, key := range traceHeaderKeys {
		if value := req.Header.Get(key); len(value) > 0 {
			traceHeaders[key] = value
		}
	}
	return &traceHeaders
}
