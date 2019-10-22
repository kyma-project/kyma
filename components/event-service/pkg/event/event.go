package event

import (
	"context"
	"encoding/json"
	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	// TODO(k15r): get rid off publish import
	"github.com/kyma-project/kyma/components/event-bus/api/publish"
	"strings"

	"github.com/kyma-project/kyma/components/event-service/pkg/event/api"
	"net/http"
)

// DecodeMessage tries to convert a http.Message to a cloudevent and validates it
func DecodeMessage(t *cehttp.Transport, ctx context.Context, message cehttp.Message) (*cloudevents.Event, *api.Error, error) {
	event, err := t.MessageToEvent(ctx, &message)
	if err != nil {
		return nil, nil, err
	}
	specErrors := []api.ErrorDetail(nil)
	err = event.Validate()
	if err != nil {
		specErrors = errorToDetails(err)
	}

	kymaErrors := Validate(event)
	allErrors := append(specErrors, kymaErrors...)
	if len(allErrors) != 0 {
		return event,  &api.Error{
			Status:  http.StatusBadRequest,
			Message: publish.ErrorMessageBadRequest,
			Type:    publish.ErrorTypeBadRequest,
			Details: allErrors,
		}, nil
	}
	return event, nil, nil
}

// RespondWithError sends an api.Error and its mentioned status as response
func RespondWithError(w http.ResponseWriter, error api.Error) error {
	w.WriteHeader(error.Status)
	if err := json.NewEncoder(w).Encode(error); err != nil {
		return err
	}
	return nil
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


// Further Kyma specific validations in addition to CloudEvents specification
func Validate(event *cloudevents.Event) []api.ErrorDetail {
	var errors []api.ErrorDetail
	eventBytes, err := event.DataBytes()
	if err != nil {
		errors = append(errors, api.ErrorDetail{
			Field:   "data",
			Type:    publish.ErrorTypeBadPayload,
			Message: err.Error(),
		})
	}
	// empty payload is considered as error by earlier /v2 endpoint which was not using cloudevents sdk-go yet
	if len(eventBytes) == 0 {
		errors = append(errors, api.ErrorDetail{
			Field:   "data",
			Type:    publish.ErrorTypeBadPayload,
			Message: "payload is missing",
		})
	}
	_, err = event.Context.GetExtension(publish.FieldEventTypeVersion)
	if err != nil {
		errors = append(errors, api.ErrorDetail{
			Field:   publish.FieldEventTypeVersion,
			Type:    publish.ErrorTypeMissingField,
			Message: publish.ErrorMessageMissingField,
		})
	}

	return errors
}
