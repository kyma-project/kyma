// Package handlertest provides utilities for Handler testing.
package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

// ValidateLegacyErrorResponse validates error responses for the legacy events endpoint.
func ValidateLegacyErrorResponse(t *testing.T, resp http.Response, wantResponse *legacyapi.PublishEventResponses) {
	legacyResponse := legacyapi.PublishEventResponses{}
	legacyErrorResponse := legacyapi.Error{}
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &legacyErrorResponse)
	require.NoError(t, err)
	legacyResponse.Error = &legacyErrorResponse
	require.Equal(t, wantResponse.Error, legacyResponse.Error)
	err = resp.Body.Close()
	require.NoError(t, err)
}

// ValidateLegacyOkResponse validates ok responses for the legacy events endpoint.
func ValidateLegacyOkResponse(t *testing.T, resp http.Response, wantResponse *legacyapi.PublishEventResponses) {
	legacyResponse := legacyapi.PublishEventResponses{}
	legacyOkResponse := legacyapi.PublishResponse{}
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &legacyOkResponse)
	require.NoError(t, err)
	legacyResponse.Ok = &legacyOkResponse
	require.Equal(t, wantResponse.Error, legacyResponse.Error)
	err = resp.Body.Close()
	require.NoError(t, err)
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
	appLister := fake.NewApplicationListerOrDie(ctx, app)
	return appLister
}

var (
	TestCasesForCloudEvents = []struct {
		Name           string
		ProvideMessage func(string) (string, http.Header)
		WantStatusCode int
	}{
		// structured cloudevents
		{
			Name: "Structured CloudEvent without id",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventType(eventType),
					testingutils.WithCloudEventID(""),
				)
				return builder.BuildStructured()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Structured CloudEvent without type",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventType(""),
				)
				return builder.BuildStructured()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Structured CloudEvent without specversion",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventType(eventType),
					testingutils.WithCloudEventSpecVersion(""),
				)
				return builder.BuildStructured()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Structured CloudEvent without source",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventType(eventType),
					testingutils.WithCloudEventSource(""),
				)
				return builder.BuildStructured()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Structured CloudEvent is valid",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventType(eventType),
				)
				return builder.BuildStructured()
			},
			WantStatusCode: http.StatusNoContent,
		},
		// binary cloudevents
		{
			Name: "Binary CloudEvent without CE-ID header",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventID(""),
				)
				return builder.BuildBinary()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Binary CloudEvent without CE-Type header",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventType(""),
				)
				return builder.BuildBinary()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Binary CloudEvent without CE-SpecVersion header",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventType(eventType),
					testingutils.WithCloudEventSpecVersion(""),
				)
				return builder.BuildBinary()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Binary CloudEvent without CE-Source header",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventType(eventType),
					testingutils.WithCloudEventSource(""),
				)
				return builder.BuildBinary()
			},
			WantStatusCode: http.StatusBadRequest,
		},
		{
			Name: "Binary CloudEvent is valid with required headers",
			ProvideMessage: func(eventType string) (string, http.Header) {
				builder := testingutils.NewCloudEventBuilder(
					testingutils.WithCloudEventType(eventType),
				)
				return builder.BuildBinary()
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
			Name: "Send a legacy event successfully with event id",
			ProvideMessage: func() (string, http.Header) {
				builder := testingutils.NewLegacyEventBuilder()
				return builder.Build()
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
			Name: "Send a legacy event successfully without event id",
			ProvideMessage: func() (string, http.Header) {
				builder := testingutils.NewLegacyEventBuilder(
					testingutils.WithLegacyEventID(""),
				)
				return builder.Build()
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
				builder := testingutils.NewLegacyEventBuilder(
					testingutils.WithLegacyEventID("invalid-id"),
				)
				return builder.Build()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getInvalidValidationError("event-id"),
			},
		},
		{
			Name: "Send a legacy event without event time",
			ProvideMessage: func() (string, http.Header) {
				builder := testingutils.NewLegacyEventBuilder(
					testingutils.WithLegacyEventTime(""),
				)
				return builder.Build()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationError("event-time"),
			},
		},
		{
			Name: "Send a legacy event without event type",
			ProvideMessage: func() (string, http.Header) {
				builder := testingutils.NewLegacyEventBuilder(
					testingutils.WithLegacyEventType(""),
				)
				return builder.Build()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationError("event-type"),
			},
		},
		{
			Name: "Send a legacy event with invalid event time",
			ProvideMessage: func() (string, http.Header) {
				builder := testingutils.NewLegacyEventBuilder(
					testingutils.WithLegacyEventTime("invalid-time"),
				)
				return builder.Build()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getInvalidValidationError("event-time"),
			},
		},
		{
			Name: "Send a legacy event without event type version",
			ProvideMessage: func() (string, http.Header) {
				builder := testingutils.NewLegacyEventBuilder(
					testingutils.WithLegacyEventTypeVersion(""),
				)
				return builder.Build()
			},
			WantStatusCode: http.StatusBadRequest,
			WantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationError("event-type-version"),
			},
		},
		{
			Name: "Send a legacy event without data",
			ProvideMessage: func() (string, http.Header) {
				builder := testingutils.NewLegacyEventBuilder(
					testingutils.WithLegacyEventData(""),
				)
				return builder.Build()
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
					Name:    testingutils.EventName,
					Version: testingutils.EventVersion,
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
