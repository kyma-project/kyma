package externalapi

import (
	"encoding/json"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
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

	/*
		EventType        string       `json:"event-type,omitempty"`
		EventTypeVersion string       `json:"event-type-version,omitempty"`
		EventID          string       `json:"event-id,omitempty"`
		EventTime        string       `json:"event-time,omitempty"`
		Data             api.AnyValue `json:"data,omitempty"`
	*/

	/*
		validate http request:
		- method
		- body
		- content type
	*/

	/*
		parse request body to PublishRequestV1
	*/

	/*
		generate an event id if there is none
	*/

	/*
		validate the PublishRequestV1 for missing / incoherent values
	*/

	/*
		convert PublishRequestV1 to CE
	*/

	/*
		extract the context from the HTTP req
	*/
	context := req.Context()
	log.Debugf("Received Context: %+v", context)

	// TODO(marcobebway) make sure that the CE headers are forwarded with the context (old filterCEHeaders func)

	/*
		send CE using mesh
	*/
	cloudEvent, err := convertPublishRequestToCloudEvent(apiv1.PublishRequestV1{})
	if err != nil {
		//TODO(marcobebway) return error
	}
	response, err := mesh.SendEvent(context, cloudEvent)

	/*
		prepare the proper response
	*/
	if err != nil {
		//TODO(marcobebway) return error
	}
	writeJSONResponse(w, response)
}

func checkParameters(parameters *apiv1.PublishEventParametersV1) (response *api.PublishEventResponses) {
	return
}

// TODO(marcobebway) is this still relevant or not
func writeJSONResponse(w http.ResponseWriter, resp *api.SendEventResponse) {
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

// TODO(marcobebway) does this still relevant or not
func getTraceHeaders(req *http.Request) *map[string]string {
	traceHeaders := make(map[string]string)
	for _, key := range traceHeaderKeys {
		if value := req.Header.Get(key); len(value) > 0 {
			traceHeaders[key] = value
		}
	}
	return &traceHeaders
}


