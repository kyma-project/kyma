package v2

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	kymaevent "github.com/kyma-project/kyma/components/event-service/pkg/event"
	"io/ioutil"
	"net/http"
	// TODO(k15r): get rid off publish import
	"github.com/kyma-project/kyma/components/event-bus/api/publish"

	ce "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go"
)

const (
	requestBodyTooLargeErrorMessage = "http: request body too large"
)

var (
	traceHeaderKeys         = []string{"x-request-id", "x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags", "x-ot-span-context"}
)

type maxBytesHandler struct {
	next  http.Handler
	limit int64
}

type CloudEventsHandler struct {
	Transport *cehttp.Transport
	Client    cloudevents.Client
}

func (h *maxBytesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Body != http.NoBody {
		r.Body = http.MaxBytesReader(rw, r.Body, h.limit)
	}
	h.next.ServeHTTP(rw, r)
}

// NewEventsHandler creates an http.Handler to handle the events endpoint
func NewEventsHandler(maxRequestSize int64) http.Handler {
	t, err := ce.NewHTTPTransport()
	if err != nil {
		return nil
	}

	handler := CloudEventsHandler{Transport: t}

	return &maxBytesHandler{next: &handler, limit: maxRequestSize}
}


func writeJSONResponse(w http.ResponseWriter, resp *api.PublishEventResponse) {
	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", httpconsts.ContentTypeApplicationJSON)
	if resp.Error != nil {
		w.WriteHeader(resp.Error.Status)
		// TODO(nachtmaar): handle error
		encoder.Encode(resp.Error)
	} else {
		// TODO(nachtmaar): handle error
		encoder.Encode(resp.Ok)
	}
	return
}

func getTraceHeaders(ctx context.Context) *map[string]string {
	tctx := cehttp.TransportContextFrom(ctx)

	traceHeaders := make(map[string]string)
	for _, key := range traceHeaderKeys {
		if value := tctx.Header.Get(key); len(value) > 0 {
			traceHeaders[key] = value
		}
	}
	return &traceHeaders
}

func (h *CloudEventsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if bus.CheckConf() != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx := req.Context()
	ctx = cehttp.WithTransportContext(ctx, cehttp.NewTransportContext(req))
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		// check if request is too large
		if err.Error() == requestBodyTooLargeErrorMessage {
			if err := kymaevent.RespondWithError(w, api.Error{
				Status:   http.StatusRequestEntityTooLarge,
				Type:     publish.ErrorTypeBadRequest,
				Message:  publish.ErrorMessageBadRequest,
				MoreInfo: "",
				Details:  nil,
			}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		// handle all other read errors
		if err := kymaevent.RespondWithError(w, api.Error{
			Status:   http.StatusBadRequest,
			Type:     publish.ErrorTypeBadRequest,
			Message:  publish.ErrorMessageBadRequest,
			MoreInfo: "",
			Details:  nil,
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	message := cehttp.Message{
		Header: req.Header,
		Body:   body,
	}

	e, apiError, err := kymaevent.FromMessage(h.Transport, ctx, message, bus.Conf.SourceID)
	if err != nil {
		if err := kymaevent.RespondWithError(w, api.Error{
			Status:   http.StatusBadRequest,
			Type:     publish.ErrorTypeBadRequest,
			Message:  publish.ErrorMessageBadRequest,
			MoreInfo: "",
			Details:  nil,
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if apiError != nil {
		if err := kymaevent.RespondWithError(w, *apiError); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	log.Debug("received event", log.WithField("event", e))
	msg, err := kymaevent.ToMessage(ctx, *e, cehttp.BinaryV03)

	log.Debug("message from event", log.WithField("message", msg))
	if err != nil {
	}

	resp, err := bus.SendEventV2(*e, *getTraceHeaders(ctx))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	log.Debug("bus response", log.WithField("response", resp))
	if resp.Ok != nil || resp.Error != nil {
		tmp := api.PublishEventResponse(*resp)
		writeJSONResponse(w, &tmp)
		return
	}

	log.Println("Cannot process event")
	http.Error(w, "Cannot process event", http.StatusInternalServerError)
	return

}
