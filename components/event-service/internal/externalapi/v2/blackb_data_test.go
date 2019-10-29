package v2

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
)

type stupidEventMock struct {
	Typ              string `json:"type,omitempty"`
	Eventtypeversion string `json:"eventtypeversion,omitempty"`
	Specversion      string `json:"specversion,omitempty"`
	ID               string `json:"id,omitempty"`
	Time             string `json:"time,omitempty"`
	Data             string `json:"data,omitempty"`
}

// TODO(nachtmaar): we cannot check the api.Error.Details except for `Field` since we are using the CE sdk validation
func TestErrorEmptyData(t *testing.T) {
	s := &stupidEventMock{
		Typ:              "order.created",
		Eventtypeversion: "v1",
		Specversion:      "0.3",
		ID:               "31109198-4d69-4ae0-972d-76117f3748c8",
		Time:             "2012-11-01T22:08:41+00:00",
	}

	ss := eventMockToString(t, s)

	wantErrorDetail := api.ErrorDetail{Field: shared.FieldData, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &ss)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyParameter(t, result, &wantErrorDetail)
	}
}

func eventMockToString(t *testing.T, s *stupidEventMock) string {
	t.Helper()
	bytes, err := json.Marshal(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	ss := string(bytes)
	return ss
}

// TODO(k15r): figure out if this test is valid. it might be perfectly fine to send an event with a `null` payload. According to the specs we would have to check the datacontenttype to verify the contents of this field
func TestErrorWrongDataNullValue(t *testing.T) {
	t.SkipNow()
	s := "{\"type\":\"order.created\",\"specversion\":\"0.3\",\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"2012-11-01T22:08:41+00:00\", \"data\":null }"
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldData, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyParameter(t, result, &wantErrorDetail)
	}
}

// TODO(k15r): figure out if this test is valid. it might be perfectly fine to send an event with a `""` payload. According to the specs we would have to check the datacontenttype to verify the contents of this field
func TestErrorWrongDataEmptyStringValue(t *testing.T) {
	t.SkipNow()
	s := "{\"type\":\"order.created\",\"specversion\":\"0.3\",\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"2012-11-01T22:08:41+00:00\"" +
		",\"data\":\"\" }"
	wantErrorDetail := api.ErrorDetail{Field: shared.FieldData, Type: shared.ErrorTypeMissingField, Message: shared.ErrorMessageMissingField, MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkEmptyParameter(t, result, &wantErrorDetail)
	}
}

func TestErrorMissingSpecVersion(t *testing.T) {
	s := &stupidEventMock{
		Typ:              "order.created",
		Eventtypeversion: "v1",
		ID:               "31109198-4d69-4ae0-972d-76117f3748c8",
		Time:             "2012-11-01T22:08:41+00:00",
		Data:             "foo",
	}

	ss := eventMockToString(t, s)
	want := &api.Error{
		Status:  http.StatusBadRequest,
		Type:    shared.ErrorTypeValidationViolation,
		Message: shared.ErrorMessageMissingField,
		Details: []api.ErrorDetail{
			{
				Field:    shared.FieldSpecVersionV2,
				Type:     shared.ErrorTypeMissingField,
				Message:  shared.ErrorMessageMissingField,
				MoreInfo: "",
			},
		},
	}

	result, err := sendAndReceiveError(t, &ss)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkResult(t, result, want)
	}
}

func TestErrorWrongDataJsonValue(t *testing.T) {
	// TODO(k15r): currently there is no way to differentiate between a broken json message and a missing spec version (see implementation of sdk-go for details)
	t.SkipNow()
	s := "{\"type\":\"order.created\",\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"2012-11-01T22:08:41+00:00\"" +
		",\"data\":\"{\"number\":\"123\"}\" }"
	wantError := api.Error{Status: http.StatusBadRequest, Type: shared.ErrorTypeBadPayload, Message: shared.ErrorMessageBadPayload,
		MoreInfo: ""}
	result, err := sendAndReceiveError(t, &s)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		checkBadRequest(t, result, &wantError)
	}
}
