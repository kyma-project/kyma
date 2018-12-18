package externalapi

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	log "github.com/sirupsen/logrus"
)

var (
	isValidEventTypeVersion = regexp.MustCompile(shared.AllowedEventTypeVersionChars).MatchString
	isValidEventID          = regexp.MustCompile(shared.AllowedEventIDChars).MatchString
	traceHeaderKeys         = []string{"x-request-id", "x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags", "x-ot-span-context"}
)

// NewEventsHandler creates an http.Handler to handle the events endpoint
func NewEventsHandler() http.Handler {
	return http.HandlerFunc(handleEvents)
}

// handleEvents handles "/v1/events" requests
func handleEvents(w http.ResponseWriter, req *http.Request) {
	if req.Body == nil || req.ContentLength == 0 {
		resp := shared.ErrorResponseBadRequest(shared.ErrorMessageBadPayload)
		writeJSONResponse(w, resp)
		return
	}
	var err error
	parameters := &api.PublishEventParameters{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&parameters.Publishrequest)
	if err != nil {
		resp := shared.ErrorResponseBadRequest(err.Error())
		writeJSONResponse(w, resp)
		return
	}
	resp := &api.PublishEventResponses{}

	traceHeaders := getTraceHeaders(req)

	err = handleEvent(parameters, resp, traceHeaders)
	if err == nil {
		if resp.Ok != nil || resp.Error != nil {
			writeJSONResponse(w, resp)
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
	if len(parameters.Publishrequest.EventID) > 0 && !isValidEventID(parameters.Publishrequest.EventID) {
		return shared.ErrorResponseWrongEventID()
	}
	if parameters.Publishrequest.Data == nil {
		return shared.ErrorResponseMissingFieldData()
	} else if d, ok := (parameters.Publishrequest.Data).(string); ok && d == "" {
		return shared.ErrorResponseMissingFieldData()
	}
	// OK
	return &api.PublishEventResponses{Ok: nil, Error: nil}
}

func writeJSONResponse(w http.ResponseWriter, resp *api.PublishEventResponses) {
	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", httpconsts.ContentTypeApplicationJSON)
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
