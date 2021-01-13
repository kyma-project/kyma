package testing

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"time"

	http2 "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// binary cloudevent headers
	CeIDHeader          = "CE-ID"
	CeTypeHeader        = "CE-Type"
	CeSourceHeader      = "CE-Source"
	CeSpecVersionHeader = "CE-SpecVersion"

	// cloudevent attributes
	CeID          = "00000"
	CeType        = "someType"
	CeSource      = "someSource"
	CeSpecVersion = "1.0"
)

func QuerySubscribedEndpoint(endpoint string) (*http.Response, error) {
	emptyBody := bytes.NewBuffer([]byte(""))
	req, err := http.NewRequest(http.MethodGet, endpoint, emptyBody)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	defer client.CloseIdleConnections()

	return client.Do(req)
}

func SendEvent(endpoint, body string, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}

	if headers != nil {
		for k, v := range headers {
			req.Header[k] = v
		}
	}

	client := http.Client{}
	defer client.CloseIdleConnections()

	return client.Do(req)
}

func GetStructuredMessageHeaders() http.Header {
	return http.Header{"Content-Type": []string{"application/cloudevents+json"}}
}

func GetBinaryMessageHeaders() http.Header {
	headers := make(http.Header)
	headers.Add(CeIDHeader, CeID)
	headers.Add(CeTypeHeader, CeType)
	headers.Add(CeSourceHeader, CeSource)
	headers.Add(CeSpecVersionHeader, CeSpecVersion)
	return headers
}

func GetApplicationJSONHeaders() http.Header {
	headers := make(http.Header)
	headers.Add(http2.ContentType, "application/json")
	return headers
}

func NewSubscription() *eventingv1alpha1.Subscription {
	filter := &eventingv1alpha1.BebFilter{
		EventSource: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "source",
			Value:    "beb.namespace",
		},
		EventType: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "type",
			Value:    "event.type.prefix.valid-app.order.created.v1",
		},
	}
	return &eventingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: eventingv1alpha1.SubscriptionSpec{
			ID:               "",
			Protocol:         "",
			ProtocolSettings: nil,
			Sink:             "",
			Filter: &eventingv1alpha1.BebFilters{
				Filters: []*eventingv1alpha1.BebFilter{
					filter,
				},
			},
		},
	}
}

// GetMissingFieldValidationErrorFor generates an Error message for a missing field
func GetMissingFieldValidationErrorFor(field string) *legacyapi.Error {
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

// IsValidEventID checks whether EventID is valid or not
func IsValidEventID(id string) bool {
	return regexp.MustCompile(legacy.AllowedEventIDChars).MatchString(id)
}

// GetInvalidValidationErrorFor generates an Error message for an invalid field
func GetInvalidValidationErrorFor(field string) *legacyapi.Error {
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

// ValidateErrorResponse validates Error Response
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

// ValidateOkResponse validates Ok Response
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

	if tcWantResponse.Ok.EventID == "" && !IsValidEventID(legacyResponse.Ok.EventID) {
		t.Errorf("should match regex: [%s] Not a valid event-id: %v ", legacy.AllowedEventIDChars, legacyResponse.Ok.EventID)
	}
	if tcWantResponse.Ok.Reason != legacyResponse.Ok.Reason {
		t.Errorf("invalid reason, want: %v, got: %v", tcWantResponse.Ok.Reason, legacyResponse.Ok.Reason)
	}
	if tcWantResponse.Ok.Status != legacyResponse.Ok.Status {
		t.Errorf("invalid status, want: %v, got: %v", tcWantResponse.Ok.Status, legacyResponse.Ok.Status)
	}
}

// WaitForHandlerToStart is waits for the test handler to start before testing could start
func WaitForHandlerToStart(t *testing.T, healthEndpoint string) {
	timeout := time.After(time.Second * 30)
	tick := time.Tick(time.Second * 1)

	for {
		select {
		case <-timeout:
			{
				t.Fatal("Failed to start handler")
			}
		case <-tick:
			{
				if resp, err := http.Get(healthEndpoint); err != nil {
					continue
				} else if resp.StatusCode == http.StatusOK {
					return
				}
			}
		}
	}
}

// GeneratePort generates a random 5 digit port
func GeneratePort() (int, error) {
	max := 4
	// Add 4 as prefix to make it 5 digits but less than 65535
	add4AsPrefix := "4"
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		return 0, err
	}
	if err != nil {
		return 0, err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}

	num, err := strconv.Atoi(fmt.Sprintf("%s%s", add4AsPrefix, string(b)))
	if err != nil {
		return 0, err
	}

	return num, nil
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9'}
