package testkit

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func SendEvent(url string, event *ExampleEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	httpClient := newHttpClient(true)

	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
