package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	cetypes "github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/publisher"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"go.uber.org/zap"
)

type CloudEventHandler struct {
	KnativePublisher *publisher.KnativePublisher
	KnativeLib       *knative.KnativeLib
	Transport        *cehttp.Transport
}

// HandleEvent finally handles the decoded event
func (handler *CloudEventHandler) HandleEvent(ctx context.Context, event cloudevents.Event) (*api.Response, *api.Error, error) {
	// make sure to get v1 event

	//TODO(k15r): should we make this configurable
	codec := cehttp.CodecV1{
		DefaultEncoding: cehttp.StructuredV1,
	}

	m, err := codec.Encode(ctx, event)
	if err != nil {
		return nil, nil, err
	}

	message, ok := m.(*cehttp.Message)
	if !ok {
		return nil, nil, fmt.Errorf("expected type http message, but got type: %v", reflect.TypeOf(m))
	}

	fmt.Printf("%v", message)

	var etv string
	var ex interface{}
	if ex, ok = event.Context.GetExtensions()[api.FieldEventTypeVersion]; !ok {
		return nil, nil, fmt.Errorf("this should never happen, sine the event has been already validated. Hence the extension should not be missing.")
	}

	// extension can have a different type depending on CE version
	if event.SpecVersion() == cloudevents.VersionV1 {
		etv, err = cetypes.ToString(ex)
		if err != nil {
			return nil, nil, err
		}
	} else if event.SpecVersion() == cloudevents.VersionV03 {
		if err := json.Unmarshal(ex.(json.RawMessage), &etv); err != nil {
			return nil, nil, err
		}
	}

	ns := knative.GetDefaultChannelNamespace()
	header := map[string][]string(message.Header)

	publishError, status, _ := (*handler.KnativePublisher).Publish(handler.KnativeLib, &ns, &header, &message.Body, event.Source(), event.Type(), etv)
	if publishError != nil {
		return nil, publishError, nil
	}

	resp := &api.Response{
		Status:  status,
		EventID: event.ID(),
		Reason:  getPublishStatusReason(&status),
	}
	return resp, nil, nil

}

// ServeHTTP implements http.Handler
// TODO(nachtmaar) add tracing and
func (handler *CloudEventHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Add the transport context to ctx.
	ctx := req.Context()
	ctx = cehttp.WithTransportContext(ctx, cehttp.NewTransportContext(req))
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

	event, err := handler.Transport.MessageToEvent(ctx, &cehttp.Message{
		Header: req.Header,
		Body:   body,
	})

	if err != nil {
		fmt.Printf("%v", err)
		//TODO(k15r): handle this here
		return
	}

	specErrors := []api.ErrorDetail(nil)
	err = event.Validate()

	if err != nil {
		specErrors = errorToDetails(err)
	}

	kymaErrors := validateKymaSpecific(event)

	allErrors := append(specErrors, kymaErrors...)
	if len(allErrors) != 0 {
		error := api.Error{
			Status:  http.StatusBadRequest,
			Message: api.ErrorMessageBadRequest,
			Type:    api.ErrorTypeBadRequest,
			Details: allErrors,
		}
		err := respondWithError(w, error)
		if err != nil {
			//TODO(k15r): handle this
		}
	}

	apiResponse, apiError, err := handler.HandleEvent(ctx, *event)
	if err != nil {
		//TODO(k15r): do shit
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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

		return
	}

	// Yeah... we got here
	if apiResponse != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(apiResponse)
	}
	w.WriteHeader(http.StatusInternalServerError)

}

func respondWithError(w http.ResponseWriter, error api.Error) error {
	w.WriteHeader(error.Status)
	if err := json.NewEncoder(w).Encode(error); err != nil {
		return err
	}
	return nil
}

func respondWithSuccess(w http.ResponseWriter, event *cloudevents.Event, status string) {

}

func errorToDetails(err error) []api.ErrorDetail {
	errors := []api.ErrorDetail(nil)

	for _, error := range strings.Split(strings.TrimSuffix(err.Error(), "\n"), "\n") {
		errors = append(errors, api.ErrorDetail{
			Message: error,
		})
	}

	return errors
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
