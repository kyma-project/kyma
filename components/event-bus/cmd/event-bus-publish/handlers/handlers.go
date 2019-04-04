package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/controllers"
	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/publish"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	requestBodyTooLargeErrorMessage = "http: request body too large"
)

// WithRequestSizeLimiting creates a new request size limiting HandlerFunc
func WithRequestSizeLimiting(next http.HandlerFunc, limit int64) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(rw, r.Body, limit)
		next.ServeHTTP(rw, r)
	}
}

// GetPublishHandler is a factory for publish events handler
// TODO research a better way for dependency injection
func GetPublishHandler(publisher *controllers.Publisher, tracer *trace.Tracer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handlePublishRequest(w, r, publisher, tracer)
	}
}

var status common.StatusReady

// GetReadinessHandler is a factory for publish events handler
func GetReadinessHandler(publisher *controllers.Publisher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if publisher != nil {
			if ok, err := (*publisher).IsReady(); ok && err == nil {
				if status.SetReady() {
					log.Printf("GetReadinessHandler :: Status: READY")
				}
				w.WriteHeader(http.StatusOK)
			} else {
				status.SetNotReady()
				log.Printf("GetReadinessHandler :: Status: NOT_READY")
				w.WriteHeader(http.StatusBadGateway)
				go func() {
					(*publisher).Stop()
					(*publisher).Start()
				}()
			}
		} else {
			status.SetNotReady()
			log.Printf("GetReadinessHandler :: statusReadyHandler :: Status: NOT_READY")
			w.WriteHeader(http.StatusBadGateway)
		}
	}
}

/* TODO:
* Log to different levels
 */
func handlePublishRequest(w http.ResponseWriter, r *http.Request, publisher *controllers.Publisher, tracer *trace.Tracer) {
	log.Println("PublishHandler :: handlePublishRequest ::  Handling request.")
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	log.Println("PublishHandler :: handlePublishRequest :: Creating publish-to-internal-broker trace span")
	var publishSpan *opentracing.Span
	var traceContext *api.TraceContext
	if (*tracer).Started() {
		spanContext := trace.ReadTraceHeaders(&r.Header)
		publishSpan = trace.StartSpan(spanContext, &(*tracer).Options().OperationName, ext.SpanKindProducer)
		traceContext = trace.WriteSpan(publishSpan)
		defer trace.FinishSpan(publishSpan)
	}

	if err != nil {
		log.Printf("PublishHandler :: handlePublishRequest :: Unexpected error while reading request body. Error: %v", err)
		var apiError *api.Error
		if err.Error() == requestBodyTooLargeErrorMessage {
			apiError = api.ErrorResponseRequestBodyTooLarge()
		} else {
			apiError = api.ErrorResponseInternalServer()
		}
		publish.SendJSONError(w, apiError)
		trace.TagSpanAsError(publishSpan, "error while reading request body", err.Error())
		return
	}
	if r.Method != "POST" {
		log.Println("PublishHandler :: handlePublishRequest :: Invalid request. Got a non POST request")
		publish.SendJSONError(w, api.ErrorResponseBadRequest())
		trace.TagSpanAsError(publishSpan, "error got a non POST request", "")
		return
	}
	if r.Body == nil {
		log.Println("PublishHandler :: handlePublishRequest :: Invalid request. Got a null body")
		publish.SendJSONError(w, api.ErrorResponseBadRequest())
		trace.TagSpanAsError(publishSpan, "error got a null request body", "")
		return
	}
	publishRequest, err := parseRequest(body)

	if err != nil {
		log.Printf("PublishHandler :: handlePublishRequest :: Error while parsing request :: Error: %v", err)
		publish.SendJSONError(w, api.ErrorResponseBadPayload())
		trace.TagSpanAsError(publishSpan, "error while parsing request", err.Error())
		return
	}

	if len(publishRequest.SourceID) == 0 {
		setSourceIdFromHeader(publishRequest, &r.Header)
	}

	errResponse := api.ValidatePublish(publishRequest, api.GetDefaultEventOptions())
	if errResponse != nil {
		log.Printf("PublishHandler :: handlePublishRequest :: Request validation failed. :: Error: %v", *errResponse)
		publish.SendJSONError(w, errResponse)
		trace.TagSpanAsError(publishSpan, errResponse.Message, "")
		return
	}
	if len(publishRequest.EventID) == 0 {
		log.Println("PublishHandler :: handlePublishRequest :: Generating event ID.")
		eventID, err := generateEventID()
		if err != nil {
			log.Printf("PublishHandler :: handlePublishRequest :: Event ID generation failed. :: Error: %v", err)
			publish.SendJSONError(w, api.ErrorResponseInternalServer())
			return
		}
		publishRequest.EventID = eventID
	}

	cloudEvent := buildCloudEvent(publishRequest, traceContext)

	log.Println("PublishHandler :: handlePublishRequest :: Constructing event body")
	body, err = json.Marshal(cloudEvent)
	if err != nil {
		log.Printf("PublishHandler :: handlePublishRequest :: Error constructing event body :: Error: %v", err)
		publish.SendJSONError(w, api.ErrorResponseInternalServer())
		trace.TagSpanAsError(publishSpan, "error constructing the event", err.Error())
		return
	}

	addSpanTagsForCloudEvent(publishSpan, &cloudEvent)

	log.Println("PublishHandler :: handlePublishRequest :: Publishing event.")
	publishResponse, err := publishEvent(publishRequest, string(body), publisher)
	if err != nil {
		log.Printf("PublishHandler :: handlePublishRequest :: Error while publishing event :: Error: %v", err)
		publish.SendJSONError(w, api.ErrorResponseInternalServer())
		trace.TagSpanAsError(publishSpan, "error while publishing the event", err.Error())
		return
	}
	log.Println("PublishHandler :: handlePublishRequest :: Event published, sending response.")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(*publishResponse)

	log.Println("PublishHandler :: handlePublishRequest :: OK.")
}

func addSpanTagsForCloudEvent(publishSpan *opentracing.Span, cloudEvent *api.CloudEvent) {
	tags := trace.CreateTraceTagsFromCloudEvent(cloudEvent)
	trace.SetSpanTags(publishSpan, &tags)
}

func parseRequest(b []byte) (*api.PublishRequest, error) {
	// Unmarshal
	log.Println("PublishHandler :: parseRequest :: Unmarsharlling request")
	var publishRequest api.PublishRequest
	err := json.Unmarshal(b, &publishRequest)
	return &publishRequest, err
}

func setSourceIdFromHeader(publishRequest *api.PublishRequest, header *http.Header) {
	sourceId := header.Get(api.HeaderSourceId)
	if len(sourceId) != 0 {
		log.Println("Setting source id from header")
		publishRequest.SourceID = sourceId
		publishRequest.SourceIdFromHeader = true
	}
}

func publishEvent(r *api.PublishRequest, body string, publisher *controllers.Publisher) (*api.PublishResponse, error) {
	subj := encodeSubject(r)
	log.Printf("PublishHandler :: publishEvent :: Publish to Subject: %s\n", subj)

	if err := (*publisher).Publish(subj, body); err != nil {
		log.Printf("PublishHandler :: publishEvent :: Error publishing message: %v\n", err)
		return nil, fmt.Errorf("error publishing message: %v", err)
	}
	return &api.PublishResponse{EventID: r.EventID}, nil
}

func encodeSubject(r *api.PublishRequest) string {
	return common.FromPublishRequest(r).Encode()
}

func generateEventID() (string, error) {
	uid, err := uuid.NewV4()
	return uid.String(), err
}

func buildCloudEvent(publishRequest *api.PublishRequest, traceContext *api.TraceContext) api.CloudEvent {
	cloudEvent := api.CloudEvent{}
	cloudEvent.PublishRequest = *publishRequest
	if traceContext != nil {
		cloudEvent.Extensions = make(api.Extensions)
		cloudEvent.Extensions[api.FieldTraceContext] = *traceContext
	}
	return cloudEvent
}
