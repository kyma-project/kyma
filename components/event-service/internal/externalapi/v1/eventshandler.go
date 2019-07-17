package v1

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
	busV1 "github.com/kyma-project/kyma/components/event-service/internal/events/bus/v1"
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

func filterCEHeaders(req *http.Request) map[string][]string {
	//forward `ce-` headers only
	headers := make(map[string][]string)
	for k := range req.Header {
		if strings.HasPrefix(strings.ToUpper(k), "CE-") {
			headers[k] = req.Header[k]
		}
	}
	return headers
}

// handleEvents handles "/v1/events" requests
func handleEvents(w http.ResponseWriter, req *http.Request) {
	if req.Body == nil || req.ContentLength == 0 {
		resp := shared.ErrorResponseBadRequest(shared.ErrorMessageBadPayload)
		writeJSONResponse(w, resp)
		return
	}
	var err error
	parameters := &apiv1.PublishEventParametersV1{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&parameters.PublishrequestV1)
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

	forwardHeaders := filterCEHeaders(req)

	err = handleEvent(parameters, resp, traceHeaders, &forwardHeaders)
	if err == nil {
		if resp.Ok != nil || resp.Error != nil {
			writeJSONResponse(w, resp)
			return
		}
		log.Errorf("cannot process event: %v", err)
		http.Error(w, "Cannot process event", http.StatusInternalServerError)
		return
	}
	log.Printf("Internal Error: %s\n", err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

var handleEvent = func(publishRequest *apiv1.PublishEventParametersV1, publishResponse *api.PublishEventResponses,
	traceHeaders *map[string]string, forwardHeaders *map[string][]string) (err error) {
	checkResp := checkParameters(publishRequest)
	if checkResp.Error != nil {
		publishResponse.Error = checkResp.Error
		return
	}
	// add source to the incoming request
	sendRequest, err := busV1.AddSource(publishRequest)
	if err != nil {
		return err
	}
	// send the event
	sendEventResponse, err := bus.SendEvent("v1", sendRequest, traceHeaders, forwardHeaders)
	if err != nil {
		return err
	}
	publishResponse.Ok = sendEventResponse.Ok
	publishResponse.Error = sendEventResponse.Error
	return err
}

func checkParameters(parameters *apiv1.PublishEventParametersV1) (response *api.PublishEventResponses) {
	if parameters == nil {
		return shared.ErrorResponseBadRequest(shared.ErrorMessageBadPayload)
	}
	if len(parameters.PublishrequestV1.EventType) == 0 {
		return shared.ErrorResponseMissingFieldEventType()
	}
	if len(parameters.PublishrequestV1.EventTypeVersion) == 0 {
		return shared.ErrorResponseMissingFieldEventTypeVersion()
	}
	if !isValidEventTypeVersion(parameters.PublishrequestV1.EventTypeVersion) {
		return shared.ErrorResponseWrongEventTypeVersion()
	}
	if len(parameters.PublishrequestV1.EventTime) == 0 {
		return shared.ErrorResponseMissingFieldEventTime()
	}
	if _, err := time.Parse(time.RFC3339, parameters.PublishrequestV1.EventTime); err != nil {
		return shared.ErrorResponseWrongEventTime(err)
	}
	if len(parameters.PublishrequestV1.EventID) > 0 && !isValidEventID(parameters.PublishrequestV1.EventID) {
		return shared.ErrorResponseWrongEventID()
	}
	if parameters.PublishrequestV1.Data == nil {
		return shared.ErrorResponseMissingFieldData()
	} else if d, ok := (parameters.PublishrequestV1.Data).(string); ok && d == "" {
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
