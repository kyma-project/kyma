package assertions

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/clientset"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/util"
)

const (
	eventsURLFormat            = "%s/%s/v1/events"
	applicationJSONContentType = "application/json"
)

type PublishRequest struct {
	EventType        string   `json:"event-type,omitempty"`
	EventTypeVersion string   `json:"event-type-version,omitempty"`
	EventID          string   `json:"event-id,omitempty"`
	EventTime        string   `json:"event-time,omitempty"`
	Data             AnyValue `json:"data,omitempty"`
}

type PublishResponse struct {
	EventID string `json:"event-id,omitempty"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

type AnyValue interface{}

type EventAPIAccessChecker struct {
	eventsBaseURL      string
	directorClient     *compass.Client
	connectorClientSet *clientset.ConnectorClientSet
	skipTLSVerify      bool
}

func NewEventAPIAccessChecker(eventsURL string, directorClient *compass.Client, skipTLSVerify bool) *EventAPIAccessChecker {
	return &EventAPIAccessChecker{
		eventsBaseURL:      eventsURL, // TODO: events URL should be taken from Application after it is implemented in Director
		directorClient:     directorClient,
		connectorClientSet: clientset.NewConnectorClientSet(clientset.WithSkipTLSVerify(skipTLSVerify)),
		skipTLSVerify:      skipTLSVerify,
	}
}

func (c *EventAPIAccessChecker) AssertEventAPIAccess(t *testing.T, application compass.Application, certificate tls.Certificate) {
	response, err := c.SendEvent(t, application, certificate)
	require.NoError(t, err, "failed to send event for Application: %s", application.GetContext())
	defer response.Body.Close()

	util.RequireStatus(t, http.StatusOK, response)
}

func (c *EventAPIAccessChecker) SendEvent(t *testing.T, application compass.Application, certificate tls.Certificate) (*http.Response, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{certificate},
				InsecureSkipVerify: c.skipTLSVerify,
			},
		},
	}

	eventsURL := fmt.Sprintf(eventsURLFormat, c.eventsBaseURL, application.Name)

	eventId := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	publishRequest := PublishRequest{
		EventType:        "order.created",
		EventTypeVersion: "v1",
		EventID:          eventId,
		EventTime:        "2012-11-01T22:08:41+00:00",
		Data:             "payload",
	}
	publishRequestEncoded, err := json.Marshal(publishRequest)
	require.NoError(t, err)

	return client.Post(eventsURL, applicationJSONContentType, bytes.NewBuffer(publishRequestEncoded))
}
