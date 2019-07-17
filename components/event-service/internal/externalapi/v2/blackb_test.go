package v2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
)

func TestErrorNoContent(t *testing.T) {
	req, err := http.NewRequest("POST", shared.EventsV2Path, nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler := NewEventsHandler(maxRequestSize)
	handler.ServeHTTP(recorder, req)
	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	body, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	result := &api.Error{}
	err = json.Unmarshal(body, result)
	if err != nil {
		t.Fatal(err)
	}
	wantError := api.Error{Status: http.StatusBadRequest, Type: shared.ErrorTypeBadPayload, Message: shared.ErrorMessageBadPayload,
		MoreInfo: "Empty request body", Details: []api.ErrorDetail{}}
	checkEmptyRequest(t, result, &wantError)
}

func TestErrorNoParameters(t *testing.T) {
	s := ""
	wantError := api.Error{Status: http.StatusBadRequest, Type: shared.ErrorTypeBadPayload, Message: shared.ErrorMessageBadPayload,
		MoreInfo: "", Details: []api.ErrorDetail{}}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyRequest(t, result, &wantError)
	}
}

func TestErrorNoType(t *testing.T) {
	s := "{\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"2012-11-01T22:08:41+00:00\"}"
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldEventTypeV2, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyParameter(t, result, &wantErrorDetail)
	}
}

func TestErrorEmptyType(t *testing.T) {
	s := "{\"type\":\"\",\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"2012-11-01T22:08:41+00:00\"}"
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldEventTypeV2, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyParameter(t, result, &wantErrorDetail)
	}
}

func TestErrorEmptyEventTime(t *testing.T) {
	s := "{\"type\":\"order.created\",\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"\"}"
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldEventTimeV2, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyParameter(t, result, &wantErrorDetail)
	}
}

func TestErrorWrongEventTime(t *testing.T) {
	s := "{\"type\":\"order.created\",\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"2012-11-01T22\"}"
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkWrongEventTime(t, result)
	}
}

func TestErrorWrongEventId(t *testing.T) {
	s := "{\"type\":\"order.created\",\"eventtypeversion\":\"v1\",\"id\":\"31109198\",\"time\":\"2012-11-01T22:08:41+00:00\", \"specversion\":\"0.3\"}"
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldEventIDV2, Type: shared.ErrorTypeInvalidField, Message: shared.ErrorMessageInvalidField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkWrongParameter(t, result, &wantErrorDetail)
	}
}

func sendAndReceiveError(t *testing.T, s *string) (result *api.Error, err error) {
	req, err := http.NewRequest("POST", shared.EventsV2Path, strings.NewReader(*s))
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler := NewEventsHandler(maxRequestSize)
	handler.ServeHTTP(recorder, req)
	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusBadRequest)
		return nil, fmt.Errorf("Unexpected http response code")
	}
	body, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	result = &api.Error{}
	err = json.Unmarshal(body, result)
	if err != nil {
		t.Fatal(err)
	}
	return result, err
}

func checkEmptyParameter(t *testing.T, result *api.Error, wantErrorDetail *api.ErrorDetail) {
	if result.Status != http.StatusBadRequest {
		t.Errorf("Wrong result.Status: got %v want %v", result.Status, http.StatusBadRequest)
	}
	if result.Type != shared.ErrorTypeValidationViolation {
		t.Errorf("Wrong result.Type: got %v want %v", result.Type, shared.ErrorTypeValidationViolation)
	}
	if result.Message != shared.ErrorMessageMissingField {
		t.Errorf("Wrong result.Message: got %v want %v", result.Message, shared.ErrorMessageMissingField)
	}
	if result.Details[0] != *wantErrorDetail {
		t.Errorf("Wrong ErrorDetails: got %v want %v", result.Details[0], *wantErrorDetail)
	}
}

func checkWrongParameter(t *testing.T, result *api.Error, wantErrorDetail *api.ErrorDetail) {
	if result.Status != http.StatusBadRequest {
		t.Errorf("Wrong result.Status: got %v want %v", result.Status, http.StatusBadRequest)
	}
	if result.Type != shared.ErrorTypeValidationViolation {
		t.Errorf("Wrong result.Type: got %v want %v", result.Type, shared.ErrorTypeValidationViolation)
	}
	if result.Message != shared.ErrorMessageInvalidField {
		t.Errorf("Wrong result.Message: got %v want %v", result.Message, shared.ErrorMessageInvalidField)
	}
	if result.Details[0] != *wantErrorDetail {
		t.Errorf("Wrong ErrorDetails: got %v want %v", result.Details[0], *wantErrorDetail)
	}
}

func checkWrongEventTime(t *testing.T, result *api.Error) {
	apiErrorDetail := api.ErrorDetail{Field: shared.FieldEventTimeV2, Type: shared.ErrorTypeInvalidField, Message: shared.ErrorMessageInvalidField, MoreInfo: ""}
	if result.Status != http.StatusBadRequest {
		t.Errorf("Wrong result.Status: got %v want %v", result.Status, http.StatusBadRequest)
	}
	if result.Type != shared.ErrorTypeValidationViolation {
		t.Errorf("Wrong result.Type: got %v want %v", result.Type, shared.ErrorTypeValidationViolation)
	}
	if result.Message != shared.ErrorMessageInvalidField {
		t.Errorf("Wrong result.Message: got %v want %v", result.Message, shared.ErrorMessageInvalidField)
	}
	if result.Details[0].Field != apiErrorDetail.Field {
		t.Errorf("Wrong ErrorDetails: got %v want %v", result.Details[0].Field, apiErrorDetail.Field)
	}
	if result.Details[0].Type != apiErrorDetail.Type {
		t.Errorf("Wrong ErrorDetails: got %v want %v", result.Details[0].Type, apiErrorDetail.Type)
	}
}

func checkEmptyRequest(t *testing.T, result *api.Error, wantError *api.Error) {
	if result.Status != http.StatusBadRequest {
		t.Errorf("Wrong result.Status: got %v want %v", result.Status, http.StatusBadRequest)
	}
	if result.Type != shared.ErrorTypeBadPayload {
		t.Errorf("Wrong result.Type: got %v want %v", result.Type, shared.ErrorTypeBadPayload)
	}
	if result.Message != shared.ErrorMessageBadPayload {
		t.Errorf("Wrong result.Message: got %v want %v", result.Message, shared.ErrorMessageBadPayload)
	}
	if len(result.Details) > 0 {
		t.Errorf("Wrong ErrorDetails: got %v want %v", result.Details, nil)
	}
}

func checkBadRequest(t *testing.T, result *api.Error, wantError *api.Error) {
	if result.Status != http.StatusBadRequest {
		t.Errorf("Wrong result.Status: got %v want %v", result.Status, http.StatusBadRequest)
	}
	if result.Type != shared.ErrorTypeBadPayload {
		t.Errorf("Wrong result.Type: got %v want %v", result.Type, shared.ErrorTypeBadPayload)
	}
	if result.Message != shared.ErrorMessageBadPayload {
		t.Errorf("Wrong result.Message: got %v want %v", result.Message, shared.ErrorMessageBadPayload)
	}
	if len(result.Details) > 0 {
		t.Errorf("Wrong ErrorDetails: got %v want %v", result.Details, nil)
	}
}
