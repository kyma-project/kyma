package testing

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// binary cloudevent headers
const (
	CeIDHeader          = "ce-id"
	CeTypeHeader        = "ce-type"
	CeSourceHeader      = "ce-source"
	CeSpecVersionHeader = "ce-specversion"
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

	for k, v := range headers {
		req.Header[k] = v
	}

	client := http.Client{}
	defer client.CloseIdleConnections()

	return client.Do(req)
}

func GetBinaryMessageHeaders() http.Header {
	headers := make(http.Header)
	headers.Add(CeIDHeader, EventID)
	headers.Add(CeTypeHeader, CloudEventTypeNotClean)
	headers.Add(CeSourceHeader, CloudEventSource)
	headers.Add(CeSpecVersionHeader, CloudEventSpecVersion)
	return headers
}

type SubscriptionOpt func(*eventingv1alpha1.Subscription)

func NewSubscription(opts ...SubscriptionOpt) *eventingv1alpha1.Subscription {
	subscription := &eventingv1alpha1.Subscription{
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
			Filter: &eventingv1alpha1.BEBFilters{
				Filters: []*eventingv1alpha1.BEBFilter{
					{
						EventSource: &eventingv1alpha1.Filter{
							Type:     "exact",
							Property: "source",
							Value:    MessagingNamespace,
						},
						EventType: &eventingv1alpha1.Filter{
							Type:     "exact",
							Property: "type",
							Value:    CloudEventType,
						},
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(subscription)
	}

	return subscription
}

func SubscriptionWithFilter(eventSource, eventType string) SubscriptionOpt {
	return func(subscription *eventingv1alpha1.Subscription) {
		subscription.Spec.Filter = &eventingv1alpha1.BEBFilters{
			Filters: []*eventingv1alpha1.BEBFilter{
				{
					EventSource: &eventingv1alpha1.Filter{
						Type:     "exact",
						Property: "source",
						Value:    eventSource,
					},
					EventType: &eventingv1alpha1.Filter{
						Type:     "exact",
						Property: "type",
						Value:    eventType,
					},
				},
			},
		}
	}
}

// WaitForEndpointStatusCodeOrFail waits for endpoint status code or timeout.
func WaitForEndpointStatusCodeOrFail(endpoint string, statusCode int) {
	timeout := time.After(time.Second * 30)
	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-timeout:
			{
				log.Fatalf("Endpoint:%s did not respond with the expected status-code:%d", endpoint, statusCode)
			}
		case <-ticker.C:
			{
				if resp, err := http.Get(endpoint); err != nil { //nolint:gosec
					continue
				} else if resp.StatusCode == statusCode {
					return
				}
			}
		}
	}
}

// GeneratePortOrDie generates a random 5 digit port or fail
func GeneratePortOrDie() int {
	tick := time.NewTicker(time.Second / 2)
	defer tick.Stop()

	timeout := time.NewTimer(time.Minute)
	defer timeout.Stop()

	for {
		select {
		case <-tick.C:
			{
				port, err := generatePort()
				if err != nil {
					break
				}

				if !isPortAvailable(port) {
					break
				}

				return port
			}
		case <-timeout.C:
			{
				log.Fatal("Failed to generate port")
			}
		}
	}
}

func generatePort() (int, error) {
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

func Is2XX(statusCode int) bool {
	return statusCode/100 == 2
}

func IsNot4XX(statusCode int) bool {
	return statusCode/100 != 4
}

// isPortAvailable returns true if the port is available for use, otherwise returns false
func isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}

	if err := listener.Close(); err != nil {
		return false
	}

	return true
}
