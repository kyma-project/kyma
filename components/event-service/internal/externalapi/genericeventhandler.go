package externalapi

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	"github.com/kyma-project/kyma/components/event-service/internal/events/mesh"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
	"time"
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

// TODO(marcobebway) do we still need this
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

func handleEvents(w http.ResponseWriter, req *http.Request) {
	// parse request body to PublishRequestV1
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

	// validate the PublishRequestV1 for missing / incoherent values
	checkResp := checkParameters(parameters)
	if checkResp.Error != nil {
		writeJSONResponse(w, checkResp)
		return
	}

	// extract the context from the HTTP req
	context := req.Context()
	log.Infof("Received Context: %+v", context)

	// TODO(marcobebway) make sure that the CE headers are forwarded with the context:
	//  - filterCEHeaders
	//  - getTraceHeaders

	// send publishRequest to meshclient, this would convert the legacy publish request to CloudEvent
	// and send it to the event mesh using cloudevent go-sdk's httpclient
	response, err := mesh.SendEvent(context, parameters)

	writeJSONResponse(w, response)
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
		return shared.ErrorResponseWrongEventTime()
	}
	if len(parameters.PublishrequestV1.EventID) > 0 && !isValidEventID(parameters.PublishrequestV1.EventID) {
		return shared.ErrorResponseWrongEventID()
	}
	if parameters.PublishrequestV1.Data == nil {
		return shared.ErrorResponseMissingFieldData()
	}
	if d, ok := (parameters.PublishrequestV1.Data).(string); ok && len(d) == 0 {
		return shared.ErrorResponseMissingFieldData()
	}
	// OK
	return &api.PublishEventResponses{}
}

func writeJSONResponse(w http.ResponseWriter, resp *api.PublishEventResponses) {
	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", httpconsts.ContentTypeApplicationJSON)

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

// TODO(marcobebway) do we still need this
func getTraceHeaders(req *http.Request) *map[string]string {
	traceHeaders := make(map[string]string)
	for _, key := range traceHeaderKeys {
		if value := req.Header.Get(key); len(value) > 0 {
			traceHeaders[key] = value
		}
	}
	return &traceHeaders
}
