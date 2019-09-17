package testkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/common/resilient"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type EventSender struct {
	httpClient resilient.HttpClient
	domain     string
}

func NewEventSender(httpClient resilient.HttpClient, domain string) *EventSender {
	return &EventSender{
		httpClient: httpClient,
		domain:     domain,
	}
}

func (s *EventSender) SendEvent(appName string, event *ExampleEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://gateway.%s/%s/v1/events", s.domain, appName)
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	respBody, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("Response from publisher: %s\n", string(respBody))

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("send event failed: %v\nrequest: %v\nresponse: %v", response.StatusCode, request, response)
	}

	return nil
}
