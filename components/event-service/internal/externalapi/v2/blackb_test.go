package v2

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
)

func TestErrorNoContent(t *testing.T) {
	want := &api.Error{
		Status:  http.StatusBadRequest,
		Type:    shared.ErrorTypeValidationViolation,
		Message: shared.ErrorMessageMissingField,
	}

	result, err := sendAndReceiveError(t, nil)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkResult(t, result, want)
	}
}

// TODO(k15r):	reconsider these checks
func TestErrorNoEvent(t *testing.T) {
	t.SkipNow()
	s := ""
	want := &api.Error{
		Status:  http.StatusBadRequest,
		Type:    shared.ErrorTypeValidationViolation,
		Message: shared.ErrorMessageMissingField,
	}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkResult(t, result, want)
	}
}

func TestErrorNoType(t *testing.T) {
	t.SkipNow()
	s := &stupidEventMock{
		Eventtypeversion: "v1",
		Specversion:      "0.3",
		Id:               "31109198-4d69-4ae0-972d-76117f3748c8",
		Time:             "2012-11-01T22:08:41+00:00",
	}

	ss := eventMockToString(t, s)

	//wantErrorDetail := api.ErrorDetail{Field: shared.FieldEventTypeV2, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	want := &api.Error{
		Status: http.StatusBadRequest,
		//Type:    shared.ErrorTypeValidationViolation,
		//Message: shared.ErrorMessageMissingField,
		//Details: []api.ErrorDetail{wantErrorDetail},
	}
	result, err := sendAndReceiveError(t, &ss)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkResult(t, result, want)
	}
}

func TestErrorEmptyType(t *testing.T) {
	t.SkipNow()
	s := &stupidEventMock{
		Typ:              "",
		Eventtypeversion: "v1",
		Specversion:      "0.3",
		Id:               "31109198-4d69-4ae0-972d-76117f3748c8",
		Time:             "2012-11-01T22:08:41+00:00",
	}

	ss := eventMockToString(t, s)
	//s := "{\"type\":\"\",\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"2012-11-01T22:08:41+00:00\"}"
	//wantErrorDetail := api.ErrorDetail{Field: shared.FieldEventTypeV2, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	want := &api.Error{
		Status: http.StatusBadRequest,
		//Type:    shared.ErrorTypeValidationViolation,
		//Message: shared.ErrorMessageMissingField,
		//Details: []api.ErrorDetail{wantErrorDetail},
	}
	result, err := sendAndReceiveError(t, &ss)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkResult(t, result, want)
	}
	//result, err := sendAndReceiveError(t, &s)
	//if err != nil {
	//	t.Errorf("%s", err)
	//} else {
	//	checkEmptyParameter(t, result, &wantErrorDetail)
	//}
}

func TestErrorEmptyEventTime(t *testing.T) {
	t.SkipNow()
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
	t.SkipNow()
	s := "{\"type\":\"order.created\",\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"2012-11-01T22\"}"
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkWrongEventTime(t, result)
	}
}

func TestErrorEventIdDoesNotMatchUUIDPattern(t *testing.T) {
	s := &stupidEventMock{
		Typ:              "mysupertype",
		Eventtypeversion: "v1",
		Specversion:      "0.3",
		Id:               "31109198",
		Time:             "2012-11-01T22:08:41+00:00",
		Data:             "bla",
	}

	ss := eventMockToString(t, s)
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldEventIDV2, Type: shared.ErrorTypeInvalidField, Message: shared.ErrorMessageInvalidField, MoreInfo: ""}
	want := &api.Error{
		Status: http.StatusBadRequest,
		//Type:    shared.ErrorTypeValidationViolation,
		//Message: shared.ErrorMessageMissingField,
		Details: []api.ErrorDetail{wantErrorDetail},
	}
	result, err := sendAndReceiveError(t, &ss)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkResult(t, result, want)
	}
}

func sendAndReceiveError(t *testing.T, s *string) (result *api.Error, err error) {
	var req *http.Request
	if s != nil {
		reader := strings.NewReader(*s)
		req, err = http.NewRequest("POST", shared.EventsV2Path, reader)
	} else {
		req, err = http.NewRequest("POST", shared.EventsV2Path, http.NoBody)
	}
	if err != nil {
		t.Fatal(err)
	}
	// send CloudEvent in structured encoding
	req.Header.Add("Content-Type", "application/cloudevents+json")

	// init bus config
	sourceID, targetURLV1, targetURLV2 := "some dummy source", "http://kyma-domain/v1/events", "http://kyma-domain/v2/events"
	if err := bus.Init(sourceID, targetURLV1, targetURLV2); err != nil {
		t.Fatalf("unable to init bus")
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

func checkResult(t *testing.T, result *api.Error, want *api.Error) {
	t.Helper()
	if want.Status > 0 && result.Status != want.Status {
		t.Errorf("Wrong result.Status: got %+v want %+v", result.Status, want.Status)
	}
	if len(want.Type) > 0 && result.Type != want.Type {
		t.Errorf("Wrong result.Type: got %+v want %+v", result.Type, want.Type)
	}
	if len(want.Message) > 0 && result.Message != want.Message {
		t.Errorf("Wrong result.Message: got %+v want %+v", result.Message, want.Message)
	}
	if want.Details != nil {
		if result.Details == nil {
			t.Errorf("Wrong ErrorDetails: got %+v want %+v", result.Details, want.Details)
		}
		if result.Details[0] != want.Details[0] {
			t.Errorf("Wrong ErrorDetails: got %+v want %+v", result.Details[0], want.Details[0])
		}
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
}
