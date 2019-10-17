package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/publisher"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
)

type CloudEventHandler struct {
	KnativePublisher *publisher.KnativePublisher
	KnativeLib       *knative.KnativeLib
	Transport        *cloudeventshttp.Transport
}

// Receive finally handles the decoded event
func (handler *CloudEventHandler) HandleEvent(ctx context.Context, event cloudevents.Event) (*api.Error, error) {
	codec := cloudeventshttp.CodecV03{
		DefaultEncoding: cloudeventshttp.StructuredV03,
	}

	m, err := codec.Encode(ctx, event)
	// err is the error from EventContext.Validate
	//if err != nil {
	//	if marshalError, ok := err.(json.MarshalerError); !ok {
	//		marshalError.Error()
	//
	//	}

	// TODO(nachtmaar): create api.ErrorDetails
	//	return nil, err
	//}

	message, ok := m.(*cloudeventshttp.Message)
	if !ok {
		return nil, fmt.Errorf("expected type http message, but got type: %v", reflect.TypeOf(m))
	}

	fmt.Printf("%v", message)

	etv, err := event.Context.GetExtension("event-type-version")

	var etvstring string

	if rawmessage, ok := etv.(json.RawMessage); ok {

		err := json.Unmarshal(rawmessage, &etvstring)
		if err != nil {
			panic(err)
		}
	}

	if err != nil {
		return nil, err
	}
	ns := knative.GetDefaultChannelNamespace()
	header := map[string][]string(message.Header)

	publishError, namespace, channelname := (*handler.KnativePublisher).Publish(handler.KnativeLib, &ns, &header, &message.Body, event.Source(), event.Type(), etvstring)
	fmt.Printf("%+v\n\n%+v\n\n%+v\n\n", publishError, namespace, channelname)

	b, err := json.Marshal(publishError)
	if err != nil {
		return publishError, err
	}
	fmt.Printf("publishError: %s", b)
	return publishError, nil
}

// ServeHTTP implements http.Handler
// TODO(nachtmaar) add tracing and
func (handler *CloudEventHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Add the transport context to ctx.
	ctx := req.Context()
	ctx = cloudeventshttp.WithTransportContext(ctx, cloudeventshttp.NewTransportContext(req))
	logger := cecontext.LoggerFrom(ctx)
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		// TODO(nachtmaar)
		logger.Errorw("failed to handle request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"Invalid request"}`))
		//r.Error()
		return
	}

	event, err := handler.Transport.MessageToEvent(ctx, &cloudeventshttp.Message{
		Header: req.Header,
		Body:   body,
	})
	if err != nil {
		logger.Errorw("failed to convert message to event", zap.Error(err))
	}
	//errorDetails := validateKymaSpecific(event)
	if err != nil {
		logger.Errorw("failed validating kyma specifics", zap.Error(err))
	}

	apiError, err := handler.HandleEvent(ctx, *event)
	if apiError != nil {
		// TODO(nachtmaar)
		//if handler.Transport.Req != nil {
		//	copyHeaders(handler.Transport.Req.Header, w.Header())
		//}
		//if len(apiError.Message.Header) > 0 {
		//	copyHeaders(apiError.Message.Header, w.Header())
		//}
		status := apiError.Status
		w.WriteHeader(status)

		w.Header().Add("Content-Length", strconv.Itoa(0))
		if len(apiError.Message) > 0 {

			apiErrorBytes, err := json.Marshal(apiError)
			w.Header().Add("Content-Length", strconv.Itoa(len(apiErrorBytes)))
			if err != nil {
				// TODO(nachtmaar) which format to use to return error ?
				status := http.StatusInternalServerError
				w.WriteHeader(status)
				//_, _ = w.Write([]byte(`{"error":"Invalid request"}`))
				//logger.Errorw("unable to marshal response", zap.Error(err))
				return
			}
			if _, err := w.Write(apiErrorBytes); err != nil {
				logger.Errorw("unable to write response, error: %s", zap.Error(err))
				//r.Error()
				return
			}
		}
		// TODO(nachtmaar) write actual response in case of no error

		//r.OK()
		return
	}

	w.WriteHeader(http.StatusNoContent)
	//r.OK()
}

func validateKymaSpecific(event *cloudevents.Event) []api.ErrorDetail {
	var errors []api.ErrorDetail
	eventBytes, err := event.DataBytes()
	if err != nil {
		errors = append(errors, api.ErrorDetail{
			Field:   "data",
			Type:    api.ErrorTypeBadPayload,
			Message: err.Error(),
		})
	}
	if len(eventBytes) == 0 {
		errors = append(errors, api.ErrorDetail{
			Field:   "data",
			Type:    api.ErrorTypeBadPayload,
			Message: "payload is missing",
		})
	}
	_, err = event.Context.GetExtension(api.FieldEventTypeVersion)
	if err != nil {
		errors = append(errors, api.ErrorDetail{
			Field:   api.FieldEventTypeVersion,
			Type:    api.ErrorTypeMissingField,
			Message: api.ErrorMessageMissingField,
		})
	}

	return errors
}
