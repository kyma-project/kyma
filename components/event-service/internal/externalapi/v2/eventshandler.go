package v2

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv2 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v2"
	busv2 "github.com/kyma-project/kyma/components/event-service/internal/events/bus/v2"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"

	log "github.com/sirupsen/logrus"
)

const (
	requestBodyTooLargeErrorMessage = "http: request body too large"
)

var (
	isValidEventTypeVersion = regexp.MustCompile(shared.AllowedEventTypeVersionChars).MatchString
	isValidEventID          = regexp.MustCompile(shared.AllowedEventIDChars).MatchString
	traceHeaderKeys         = []string{"x-request-id", "x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags", "x-ot-span-context"}
	specVersion             = "0.3"
)

type maxBytesHandler struct {
	next  http.Handler
	limit int64
}

func (h *maxBytesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(rw, r.Body, h.limit)
	h.next.ServeHTTP(rw, r)
}

// NewEventsHandler creates an http.Handler to handle the events endpoint
func NewEventsHandler(maxRequestSize int64) http.Handler {
	return &maxBytesHandler{next: http.HandlerFunc(handleEvents), limit: maxRequestSize}
}

// FilterCEHeaders filters Cloud Events Headers
func FilterCEHeaders(req *http.Request) map[string][]string {
	//forward `ce-` headers only
	headers := make(map[string][]string)
	for k := range req.Header {
		if strings.HasPrefix(strings.ToUpper(k), "CE-") {
			headers[k] = req.Header[k]
		}
	}
	return headers
}

// handleEvents handles "/v2/events" requests
func handleEvents(w http.ResponseWriter, req *http.Request) {
	if req.Body == nil || req.ContentLength == 0 {
		resp := shared.ErrorResponseBadRequest(shared.ErrorMessageBadPayload)
		writeJSONResponse(w, resp)
		return
	}
	var err error
	parameters := &apiv2.PublishEventParametersV2{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&parameters.EventRequestV2)
	if err != nil {
		var resp *api.PublishEventResponses
		if err.Error() == requestBodyTooLargeErrorMessage {
			resp = shared.ErrorResponseRequestBodyTooLarge(err.Error())
		} else {
			resp = shared.ErrorResponseBadRequest(err.Error())
		}
		writeJSONResponse(w, resp)
		return
	}
	resp := &api.PublishEventResponses{}

	traceHeaders := getTraceHeaders(req)

	forwardHeaders := FilterCEHeaders(req)

	err = handleEvent(parameters, resp, traceHeaders, &forwardHeaders)
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

var handleEvent = func(publishRequest *apiv2.PublishEventParametersV2, publishResponse *api.PublishEventResponses,
	traceHeaders *map[string]string, forwardHeaders *map[string][]string) (err error) {
	checkResp := checkParameters(publishRequest)
	if checkResp.Error != nil {
		publishResponse.Error = checkResp.Error
		return
	}
	// add source to the incoming request
	sendRequest, err := busv2.AddSource(publishRequest)
	if err != nil {
		return err
	}
	// send the event
	sendEventResponse, err := bus.SendEvent(sendRequest, traceHeaders, forwardHeaders)
	if err != nil {
		return err
	}
	publishResponse.Ok = sendEventResponse.Ok
	publishResponse.Error = sendEventResponse.Error
	return err
}

func checkParameters(parameters *apiv2.PublishEventParametersV2) (response *api.PublishEventResponses) {
	if parameters == nil {
		return shared.ErrorResponseBadRequest(shared.ErrorMessageBadPayload)
	}
	if len(parameters.EventRequestV2.EventID) == 0 {
		return ErrorResponseMissingFieldEventID()
	}
	if len(parameters.EventRequestV2.EventType) == 0 {
		return ErrorResponseMissingFieldEventType()
	}
	if len(parameters.EventRequestV2.EventTypeVersion) == 0 {
		return ErrorResponseMissingFieldEventTypeVersion()
	}
	if !isValidEventTypeVersion(parameters.EventRequestV2.EventTypeVersion) {
		return ErrorResponseWrongEventTypeVersion()
	}
	if len(parameters.EventRequestV2.EventTime) == 0 {
		return ErrorResponseMissingFieldEventTime()
	}
	if _, err := time.Parse(time.RFC3339, parameters.EventRequestV2.EventTime); err != nil {
		return ErrorResponseWrongEventTime()
	}
	if len(parameters.EventRequestV2.EventID) > 0 && !isValidEventID(parameters.EventRequestV2.EventID) {
		return ErrorResponseWrongEventID()
	}
	if parameters.EventRequestV2.SpecVersion != specVersion {
		return ErrorResponseWrongSpecVersion()
	}
	if parameters.EventRequestV2.Data == nil {
		return ErrorResponseMissingFieldData()
	} else if d, ok := (parameters.EventRequestV2.Data).(string); ok && d == "" {
		return ErrorResponseMissingFieldData()
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
