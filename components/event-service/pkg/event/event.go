package event

import (
	"context"
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
	"reflect"
	"regexp"
	"strings"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"net/http"
)

var (
	isValidEventID = regexp.MustCompile(shared.AllowedEventIDChars).MatchString
)

// FromMessage tries to convert a http.Message to a cloudevent and validates it
func FromMessage(t *cehttp.Transport, ctx context.Context, message cehttp.Message) (*cloudevents.Event, *api.Error, error) {
	event, err := t.MessageToEvent(ctx, &message)
	if err != nil {
		if err.Error() == "transport http failed to convert message: cloudevents version unknown" {
			apiErr := &api.Error{
				Status:   http.StatusBadRequest,
				Type:     shared.ErrorTypeValidationViolation,
				Message:  shared.ErrorMessageMissingField,
				MoreInfo: "",
				Details: []api.ErrorDetail{
					api.ErrorDetail{
						Field:    shared.FieldSpecVersionV2,
						Type:     shared.ErrorTypeMissingField,
						Message:  shared.ErrorMessageMissingField,
						MoreInfo: "",
					},
				},
			}
			return nil, apiErr, nil
		}
		return nil, nil, err
	}
	specErrors := []api.ErrorDetail(nil)

	// add source to the incoming request
	// NOTE: this needs to be done before validating the event via CloudEvents sdk
	event, err = AddSource(*event)
	if err != nil {
		return event, nil, err
	}

	// validate event the CloudEvents way
	err = event.Validate()
	if err != nil {
		specErrors = errorToDetails(err)
	}

	// validate event the Kyma way
	kymaErrors := Validate(event)

	allErrors := append(kymaErrors, specErrors...)
	if len(allErrors) != 0 {
		return event, &api.Error{
			Status:  http.StatusBadRequest,
			Message: shared.ErrorMessageMissingField,
			Type:    shared.ErrorTypeValidationViolation,
			Details: allErrors,
		}, nil
	}

	tmpetv := event.Extensions()[shared.FieldEventTypeVersionV2]

	var etv string

	switch v := tmpetv.(type) {
	case string:
		etv = v
	case *string:
		etv = *v
	case json.RawMessage:
		if err := json.Unmarshal(v, &etv); err != nil {
			return nil, &api.Error{
				Status:  http.StatusBadRequest,
				Message: shared.ErrorMessageMissingField,
				Type:    shared.ErrorTypeValidationViolation,
				Details: allErrors,
			}, err
		}
	}

	event.SetExtension(shared.FieldEventTypeVersionV2, etv)

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
			Field:   shared.FieldData,
			Type:    shared.ErrorTypeMissingField,
			Message: err.Error(),
		})
	}
	// empty payload is considered as error by earlier /v2 endpoint which was not using cloudevents sdk-go yet
	if len(eventBytes) == 0 {
		errors = append(errors, api.ErrorDetail{
			Field:   shared.FieldData,
			Type:    shared.ErrorTypeMissingField,
			Message: shared.ErrorMessageMissingField,
		})
	}
	_, err = event.Context.GetExtension(shared.FieldEventTypeVersionV2)
	if err != nil {
		errors = append(errors, api.ErrorDetail{
			Field:   shared.FieldEventTypeVersionV2,
			Type:    shared.ErrorTypeMissingField,
			Message: shared.ErrorMessageMissingField,
		})
	}

	if !isValidEventID(event.ID()) {
		errors = append(errors, api.ErrorDetail{
			Field:   shared.FieldEventIDV2,
			Type:    shared.ErrorTypeInvalidField,
			Message: shared.ErrorMessageInvalidField,
		})

	}

	return errors
}

func ToMessage(ctx context.Context, event cloudevents.Event, encoding cehttp.Encoding) (*cehttp.Message, error) {

	codec := &cehttp.Codec{
		Encoding:                   encoding,
		DefaultEncodingSelectionFn: nil,
	}

	message, err := codec.Encode(ctx, event)
	if err != nil {
		return nil, err
	}

	msg, ok := message.(*cehttp.Message)
	if !ok {
		return nil, fmt.Errorf("cannot convert to http.Message: %v type: %v", message, reflect.TypeOf(message))
	}

	return msg, nil
}

// AddSource adds the "source" related data to the incoming request
func AddSource(event cloudevents.Event) (*cloudevents.Event, error) {
	if err := bus.CheckConf(); err != nil {
		return nil, err
	}
	event.SetSource(bus.Conf.SourceID)
	return &event, nil
}
