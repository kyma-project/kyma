package externalapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/mesh"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
	meshtesting "github.com/kyma-project/kyma/components/event-service/internal/testing"
)

func TestNewEventsHandler(t *testing.T) {

	const (
		source      = "mock"
		requestSize = 65536
		v1Endpoint  = "/" + source + "/v1/events"
	)
	meshURL, closeFn := meshtesting.MockEventMesh(t)
	defer closeFn()

	conf, err := mesh.InitConfig(source, meshURL)
	if err != nil {
		t.Fatalf("Error init config: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle(v1Endpoint, NewEventsHandler(conf, requestSize))

	tests := []struct {
		name           string
		givenPayload   string
		wantCode       int
		wantEventID    string
		wantErrorType  string
		wantErrorField string
	}{
		{
			name:          "invalid event bad payload",
			wantCode:      http.StatusBadRequest,
			wantErrorType: shared.ErrorTypeBadPayload,
		},
		{
			name:           "invalid event missing field event type",
			givenPayload:   `{"event-type-version":"v1", "event-time":"2018-11-02T22:08:41+00:00", "data":{ "order-number":123}}`,
			wantCode:       http.StatusBadRequest,
			wantErrorType:  shared.ErrorTypeValidationViolation,
			wantErrorField: shared.FieldEventType,
		},
		{
			name:           "invalid event missing field event type version",
			givenPayload:   `{"event-type":"order.created", "event-time":"2018-11-02T22:08:41+00:00", "data":{ "order-number":123}}`,
			wantCode:       http.StatusBadRequest,
			wantErrorType:  shared.ErrorTypeValidationViolation,
			wantErrorField: shared.FieldEventTypeVersion,
		},
		{
			name:           "invalid event wrong event type version",
			givenPayload:   `{"event-type":"order.created", "event-type-version":"!", "event-time":"2018-11-02T22:08:41+00:00", "data":{ "order-number":123}}`,
			wantCode:       http.StatusBadRequest,
			wantErrorType:  shared.ErrorTypeValidationViolation,
			wantErrorField: shared.FieldEventTypeVersion,
		},
		{
			name:           "invalid event missing field event time",
			givenPayload:   `{"event-type":"order.created", "event-type-version":"v1", "data":{ "order-number":123}}`,
			wantCode:       http.StatusBadRequest,
			wantErrorType:  shared.ErrorTypeValidationViolation,
			wantErrorField: shared.FieldEventTime,
		},
		{
			name:           "invalid event wrong event time",
			givenPayload:   `{"event-type":"order.created", "event-type-version":"v1", "event-time":"invalid", "data":{ "order-number":123}}`,
			wantCode:       http.StatusBadRequest,
			wantErrorType:  shared.ErrorTypeValidationViolation,
			wantErrorField: shared.FieldEventTime,
		},
		{
			name:           "invalid event wrong event id",
			givenPayload:   `{"event-id":"!", "event-type":"order.created", "event-type-version":"v1", "event-time":"2018-11-02T22:08:41+00:00", "data":{ "order-number":123}}`,
			wantCode:       http.StatusBadRequest,
			wantErrorType:  shared.ErrorTypeValidationViolation,
			wantErrorField: shared.FieldEventID,
		},
		{
			name:           "invalid event missing field data",
			givenPayload:   `{"event-type":"order.created", "event-type-version":"v1", "event-time":"2018-11-02T22:08:41+00:00"}`,
			wantCode:       http.StatusBadRequest,
			wantErrorType:  shared.ErrorTypeValidationViolation,
			wantErrorField: shared.FieldData,
		},
		{
			name:           "invalid event empty field data",
			givenPayload:   `{"event-type":"order.created", "event-type-version":"v1", "event-time":"2018-11-02T22:08:41+00:00", "data":""}`,
			wantCode:       http.StatusBadRequest,
			wantErrorType:  shared.ErrorTypeValidationViolation,
			wantErrorField: shared.FieldData,
		},
		{
			name:         "valid event without event-id",
			givenPayload: `{"event-type":"order.created", "event-type-version":"v1", "event-time":"2018-11-02T22:08:41+00:00", "data":{ "order-number":123}}`,
			wantCode:     http.StatusOK,
		},
		{
			name:         "valid event with event-id",
			givenPayload: `{"event-id":"8954ad1c-78ed-4c58-a639-68bd44031de0", "event-type":"order.created", "event-type-version":"v1", "event-time":"2018-11-02T22:08:41+00:00", "data":{ "order-number":123}}`,
			wantCode:     http.StatusOK,
			wantEventID:  "8954ad1c-78ed-4c58-a639-68bd44031de0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", v1Endpoint, strings.NewReader(test.givenPayload))
			if err != nil {
				t.Fatalf("Post request error: %s", err)
			}

			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, req)

			if test.wantCode != recorder.Code {
				t.Fatalf("Test '%s' failed with status code mismatch, want '%d' but got '%d'", test.name, test.wantCode, recorder.Code)
			}
			if recorder.Body == nil {
				t.Fatalf("Test '%s' failed with nil body", test.name)
			}
			if recorder.Body.Len() == 0 {
				t.Fatalf("Test '%s' failed with empty body", test.name)
			}
			if recorder.Code == http.StatusOK {
				response := &api.PublishResponse{}
				if err := json.NewDecoder(recorder.Body).Decode(response); err != nil {
					t.Fatalf("Failed to decode response %v", err)
				}
				if len(test.wantEventID) > 0 && test.wantEventID != response.EventID {
					t.Fatalf("Test '%s' failed with response event ID mismatch, want '%s' but got '%s'", test.name, test.wantEventID, response.EventID)
				}
			} else {
				response := &api.Error{}
				if err = json.NewDecoder(recorder.Body).Decode(response); err != nil {
					t.Fatalf("Failed to decode response %v", err)
				}
				if test.wantErrorType != response.Type {
					t.Fatalf("Test '%s' failed with response error type mismatch, want '%s' but got '%s'", test.name, test.wantErrorType, response.Type)
				}
				if len(response.Details) > 0 && test.wantErrorField != response.Details[0].Field {
					t.Fatalf("Test '%s' failed with response error field mismatch, want '%s' but got '%s'", test.name, test.wantErrorField, response.Details[0].Field)
				}
			}
		})
	}
}

func TestNewEventsHandlerWithSmallRequestSize(t *testing.T) {

	const (
		requestSize = 0
		source      = "mock"
		v1Endpoint  = "/" + source + "/v1/events"
	)
	meshURL, closeFn := meshtesting.MockEventMesh(t)
	defer closeFn()

	conf, err := mesh.InitConfig(source, meshURL)
	if err != nil {
		t.Fatalf("Error init config: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle(v1Endpoint, NewEventsHandler(conf, requestSize))

	tests := []struct {
		name          string
		givenPayload  string
		wantCode      int
		wantErrorType string
	}{
		{
			name:          "valid event with large payload",
			givenPayload:  `{"event-type":"order.created", "event-type-version":"v1", "event-time":"2018-11-02T22:08:41+00:00", "data":{ "order-number":123}}`,
			wantCode:      http.StatusRequestEntityTooLarge,
			wantErrorType: shared.ErrorTypeRequestBodyTooLarge,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", v1Endpoint, strings.NewReader(test.givenPayload))
			if err != nil {
				t.Fatalf("Post request error: %s", err)
			}

			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, req)
			response := &api.Error{}

			if test.wantCode != recorder.Code {
				t.Fatalf("Test '%s' failed with status code mismatch, want '%d' but got '%d'", test.name, test.wantCode, recorder.Code)
			}
			if recorder.Body == nil {
				t.Fatalf("Test '%s' failed with nil body", test.name)
			}
			if recorder.Body.Len() == 0 {
				t.Fatalf("Test '%s' failed with empty body", test.name)
			}
			if err = json.NewDecoder(recorder.Body).Decode(response); err != nil {
				t.Fatalf("Failed to decode response %v", err)
			}
			if test.wantErrorType != response.Type {
				t.Fatalf("Test '%s' failed with response error type mismatch, want '%s' but got '%s'", test.name, test.wantErrorType, response.Type)
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	t.Parallel()

	// test meta
	const (
		v2Endpoint  = "/mock/v2/events"
		mockMeshURL = "http://localhost:8080/events"
	)

	// prepare a post request to the legacy v2 endpoint
	req, err := http.NewRequest("POST", v2Endpoint, nil)
	if err != nil {
		t.Fatalf("post request error %s:", err)
	}

	// prepare a response recorder
	recorder := httptest.NewRecorder()

	// prepare an HTTP handler
	handler := http.NewServeMux()
	handler.Handle(v2Endpoint, NewPermanentRedirectionHandler(mockMeshURL))
	handler.ServeHTTP(recorder, req)

	// assert correct status code
	if statusCode := recorder.Code; statusCode != http.StatusMovedPermanently {
		t.Fatalf("invalid status code, want: %d but got: %d", http.StatusMovedPermanently, statusCode)
	}

	// assert empty body
	if responseBody := recorder.Body.String(); len(responseBody) > 0 {
		t.Fatalf("response body should be empty, but got: '%s'", responseBody)
	}

	// assert correct redirect location
	if redirectLocation := recorder.Header().Get("Location"); len(redirectLocation) == 0 {
		t.Fatalf("redirect location header is not found")
	} else if redirectLocation != mockMeshURL {
		t.Fatalf("invalid redirect location header, want: '%s' but got: '%s'", mockMeshURL, redirectLocation)
	}
}
