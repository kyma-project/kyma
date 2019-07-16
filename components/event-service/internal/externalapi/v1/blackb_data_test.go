package v1

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
)

func TestErrorEmptyData(t *testing.T) {
	s := "{\"event-type\":\"order.created\",\"event-type-version\":\"v1\",\"event-id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"event-time\":\"2012-11-01T22:08:41+00:00\"}"
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldData, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyParameter(t, result, &wantErrorDetail)
	}
}

func TestErrorWrongDataNullValue(t *testing.T) {
	s := "{\"event-type\":\"order.created\",\"event-type-version\":\"v1\",\"event-id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"event-time\":\"2012-11-01T22:08:41+00:00\"" +
		",\"data\":null }"
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldData, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyParameter(t, result, &wantErrorDetail)
	}
}

func TestErrorWrongDataEmptyStringValue(t *testing.T) {
	s := "{\"event-type\":\"order.created\",\"event-type-version\":\"v1\",\"event-id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"event-time\":\"2012-11-01T22:08:41+00:00\"" +
		",\"data\":\"\" }"
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldData, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyParameter(t, result, &wantErrorDetail)
	}
}

func TestErrorWrongDataJsonValue(t *testing.T) {
	s := "{\"event-type\":\"order.created\",\"event-type-version\":\"v1\",\"event-id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"event-time\":\"2012-11-01T22:08:41+00:00\"" +
		",\"data\":\"{\"number\":\"123\"}\" }"
	wantError := api.Error{Status: http.StatusBadRequest, Type: shared.ErrorTypeBadPayload, Message: shared.ErrorMessageBadPayload,
		MoreInfo: "", Details: []api.ErrorDetail{}}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkBadRequest(t, result, &wantError)
	}
}
