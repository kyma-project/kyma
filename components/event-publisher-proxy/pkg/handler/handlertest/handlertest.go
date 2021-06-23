// Package handlertest provides utilities for Handler testing.
package handlertest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// isValidEventID checks whether EventID is valid or not
func isValidEventID(id string) bool {
	return regexp.MustCompile(legacy.AllowedEventIDChars).MatchString(id)
}

// validateErrorResponse validates Error Response
func ValidateErrorResponse(t *testing.T, resp http.Response, tcWantResponse *legacyapi.PublishEventResponses) {
	legacyResponse := legacyapi.PublishEventResponses{}
	legacyError := legacyapi.Error{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if err = json.Unmarshal(bodyBytes, &legacyError); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	legacyResponse.Error = &legacyError
	if !reflect.DeepEqual(tcWantResponse.Error, legacyResponse.Error) {
		t.Fatalf("Invalid error, want: %v, got: %v", tcWantResponse.Error, legacyResponse.Error)
	}
}

// validateOkResponse validates Ok Response
func ValidateOkResponse(t *testing.T, resp http.Response, tcWantResponse *legacyapi.PublishEventResponses) {
	legacyOkResponse := legacyapi.PublishResponse{}
	legacyResponse := legacyapi.PublishEventResponses{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if err = json.Unmarshal(bodyBytes, &legacyOkResponse); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	legacyResponse.Ok = &legacyOkResponse
	if err = resp.Body.Close(); err != nil {
		t.Fatalf("failed to close body: %v", err)
	}

	if tcWantResponse.Ok.EventID != "" && tcWantResponse.Ok.EventID != legacyResponse.Ok.EventID {
		t.Errorf("invalid event-id, want: %v, got: %v", tcWantResponse.Ok.EventID, legacyResponse.Ok.EventID)
	}

	if tcWantResponse.Ok.EventID == "" && !isValidEventID(legacyResponse.Ok.EventID) {
		t.Errorf("should match regex: [%s] Not a valid event-id: %v ", legacy.AllowedEventIDChars, legacyResponse.Ok.EventID)
	}
	if tcWantResponse.Ok.Reason != legacyResponse.Ok.Reason {
		t.Errorf("invalid reason, want: %v, got: %v", tcWantResponse.Ok.Reason, legacyResponse.Ok.Reason)
	}
	if tcWantResponse.Ok.Status != legacyResponse.Ok.Status {
		t.Errorf("invalid status, want: %v, got: %v", tcWantResponse.Ok.Status, legacyResponse.Ok.Status)
	}
}

// getMissingFieldValidationError generates an Error message for a missing field
func getMissingFieldValidationError(field string) *legacyapi.Error {
	return &legacyapi.Error{
		Status:  400,
		Type:    "validation_violation",
		Message: "Missing field",
		Details: []legacyapi.ErrorDetail{
			{
				Field:    field,
				Type:     "missing_field",
				Message:  "Missing field",
				MoreInfo: "",
			},
		},
	}
}

// getInvalidValidationError generates an Error message for an invalid field
func getInvalidValidationError(field string) *legacyapi.Error {
	return &legacyapi.Error{
		Status:  400,
		Type:    "validation_violation",
		Message: "Invalid field",
		Details: []legacyapi.ErrorDetail{
			{
				Field:    field,
				Type:     "invalid_field",
				Message:  "Invalid field",
				MoreInfo: "",
			},
		},
	}
}

func NewApplicationListerOrDie(ctx context.Context, appName string) *application.Lister {
	app := applicationtest.NewApplication(appName, nil)
	appLister := fake.NewListerOrDie(ctx, app)
	return appLister
}

// common test-cases for the HTTP handler and NATS handler
var (
	TestCasesForCloudEvents = []struct {
		Name           string
		ProvideMessage func() (string, http.Header)
		WantStatusCode int
	}{
		// structured cloudevents
		{
			Name: "Structured CloudEvent without id",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayloadWithoutID, testingutils.GetStructuredMessageHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Structured CloudEvent without type",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayloadWithoutType, testingutils.GetStructuredMessageHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Structured CloudEvent without specversion",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayloadWithoutSpecVersion, testingutils.GetStructuredMessageHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Structured CloudEvent without source",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayloadWithoutSource, testingutils.GetStructuredMessageHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Structured CloudEvent is valid",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayload, testingutils.GetStructuredMessageHeaders()
			},
			WantStatusCode: http.StatusNoContent,
		},
		// binary cloudevents
		{
			Name: "Binary CloudEvent without CE-ID header",
			ProvideMessage: func() (string, http.Header) {
				headers := testingutils.GetBinaryMessageHeaders()
				headers.Del(testingutils.CeIDHeader)
				return fmt.Sprintf(`"%s"`, testingutils.CloudEventData), headers
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Binary CloudEvent without CE-Type header",
			ProvideMessage: func() (string, http.Header) {
				headers := testingutils.GetBinaryMessageHeaders()
				headers.Del(testingutils.CeTypeHeader)
				return fmt.Sprintf(`"%s"`, testingutils.CloudEventData), headers
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Binary CloudEvent without CE-SpecVersion header",
			ProvideMessage: func() (string, http.Header) {
				headers := testingutils.GetBinaryMessageHeaders()
				headers.Del(testingutils.CeSpecVersionHeader)
				return fmt.Sprintf(`"%s"`, testingutils.CloudEventData), headers
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Binary CloudEvent without CE-Source header",
			ProvideMessage: func() (string, http.Header) {
				headers := testingutils.GetBinaryMessageHeaders()
				headers.Del(testingutils.CeSourceHeader)
				return fmt.Sprintf(`"%s"`, testingutils.CloudEventData), headers
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Binary CloudEvent is valid with required headers",
			ProvideMessage: func() (string, http.Header) {
				return fmt.Sprintf(`"%s"`, testingutils.CloudEventData), testingutils.GetBinaryMessageHeaders()
			},
			WantStatusCode: http.StatusNoContent,
		},
	}

	TestCasesForLegacyEvents = []struct {
		Name           string
		ProvideMessage func() (string, http.Header)
		WantStatusCode int
		WantResponse   legacyapi.PublishEventResponses
	}{
		{
			Name: "Send a legacy event successfully with event-id",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.ValidLegacyEventPayloadWithEventId, testingutils.GetApplicationJSONHeaders()
			},
			WantStatusCode: http.StatusOK,
			WantResponse: legacyapi.PublishEventResponses{
				Ok: &legacyapi.PublishResponse{
					EventID: testingutils.EventID,
					Status:  "",
					Reason:  "",
				},
			},
		},
		{
			Name: "Send a legacy event successfully without event-id",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.ValidLegacyEventPayloadWithoutEventId, testingutils.GetApplicationJSONHeaders()
			},
			WantStatusCode: http.StatusOK,
			WantResponse: legacyapi.PublishEventResponses{
				Ok: &legacyapi.PublishResponse{
					EventID: "",
					Status:  "",
					Reason:  "",
				},
			},
		},
		{
			Name: "Send a legacy event with invalid event id",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithInvalidEventId, testingutils.GetApplicationJSONHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getInvalidValidationError("event-id"),
			},
		},
		{
			Name: "Send a legacy event without event time",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithoutEventTime, testingutils.GetApplicationJSONHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationError("event-time"),
			},
		},
		{
			Name: "Send a legacy event without event type",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithoutEventType, testingutils.GetApplicationJSONHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationError("event-type"),
			},
		},
		{
			Name: "Send a legacy event with invalid event time",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithInvalidEventTime, testingutils.GetApplicationJSONHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getInvalidValidationError("event-time"),
			},
		},
		{
			Name: "Send a legacy event without event version",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithWithoutEventVersion, testingutils.GetApplicationJSONHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationError("event-type-version"),
			},
		},
		{
			Name: "Send a legacy event without data field",
			ProvideMessage: func() (string, http.Header) {
				return testingutils.ValidLegacyEventPayloadWithoutData, testingutils.GetApplicationJSONHeaders()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationError("data"),
			},
		},
	}

	TestCasesForSubscribedEndpoint = []struct {
		Name               string
		AppName            string
		InputSubscriptions []eventingv1alpha1.Subscription
		WantStatusCode     int
		WantResponse       subscribed.Events
	}{
		{
			Name:           "Send a request with a valid application Name",
			AppName:        testingutils.ApplicationName,
			WantStatusCode: http.StatusOK,
			WantResponse: subscribed.Events{
				EventsInfo: []subscribed.Event{{
					Name:    testingutils.LegacyEventType,
					Version: testingutils.LegacyEventTypeVersion,
				}},
			},
		},
		{
			Name:           "Send a request with an invalid application Name",
			AppName:        "invalid-app",
			WantStatusCode: http.StatusOK,
			WantResponse: subscribed.Events{
				EventsInfo: []subscribed.Event{},
			},
		},
	}
)
