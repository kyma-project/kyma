package client

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"net/http"

	config2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/config"
	types2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/types"
	auth2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/auth"
	httpclient2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/httpclient"
)

// compile time check
var _ Interface = Client{}

type Interface interface {
	Publish(token *auth2.AccessToken, cloudEvent cloudevents.Event, qos types2.Qos) (*types2.PublishResponse, error)
	Create(token *auth2.AccessToken, subscription *types2.Subscription) (*types2.CreateResponse, error)
	List(token *auth2.AccessToken) (types2.Subscriptions, *types2.Response, error)
	Get(token *auth2.AccessToken, name string) (*types2.Subscription, *types2.Response, error)
	Delete(token *auth2.AccessToken, name string) (*types2.DeleteResponse, error)
	TriggerHandshake(token *auth2.AccessToken, name string) (*types2.TriggerHandshake, error)
	UpdateState(token *auth2.AccessToken, name string, state types2.State) (*types2.UpdateStateResponse, error)
}

type Client struct {
	config     *config2.Config
	httpClient *httpclient2.Client
}

func NewClient(config *config2.Config) *Client {
	return &Client{
		config:     config,
		httpClient: httpclient2.NewHttpClient(),
	}
}

func (c Client) Publish(token *auth2.AccessToken, event cloudevents.Event, qos types2.Qos) (*types2.PublishResponse, error) {
	req, err := c.httpClient.NewRequest(token, http.MethodPost, c.config.PublishURL, event)
	if err != nil {
		return nil, err
	}

	// set required headers
	req.Header.Set("qos", string(qos))
	req.Header.Set("Content-Type", "application/cloudevents+json")

	var response types2.PublishResponse
	if resp, err := c.httpClient.Do(req, &response); err != nil {
		return nil, err
	} else {
		response.StatusCode = resp.StatusCode
	}

	return &response, nil
}

func (c Client) Create(token *auth2.AccessToken, subscription *types2.Subscription) (*types2.CreateResponse, error) {
	req, err := c.httpClient.NewRequest(token, http.MethodPost, c.config.CreateURL, subscription)
	if err != nil {
		return nil, err
	}

	var response types2.CreateResponse
	if resp, err := c.httpClient.Do(req, response); err != nil {
		return nil, err
	} else if resp == nil {
		return nil, fmt.Errorf("could not unmarshal response: %v", resp)
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
	}

	return &response, nil
}

func (c Client) List(token *auth2.AccessToken) (types2.Subscriptions, *types2.Response, error) {
	req, err := c.httpClient.NewRequest(token, http.MethodGet, c.config.ListURL, nil)
	if err != nil {
		return nil, nil, err
	}

	var subscriptions types2.Subscriptions
	var response types2.Response
	if resp, err := c.httpClient.Do(req, &subscriptions); err != nil {
		return nil, nil, err
	} else if resp == nil {
		return nil, nil, fmt.Errorf("could not unmarshal response: %v", resp)
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
	}

	return subscriptions, &response, nil
}

func (c Client) Get(token *auth2.AccessToken, name string) (*types2.Subscription, *types2.Response, error) {
	req, err := c.httpClient.NewRequest(token, http.MethodGet, fmt.Sprintf(c.config.GetURLFormat, name), nil)
	if err != nil {
		return nil, nil, err
	}

	var subscription *types2.Subscription
	var response types2.Response
	if resp, err := c.httpClient.Do(req, &subscription); err != nil {
		return nil, nil, err
	} else if resp == nil {
		return nil, nil, fmt.Errorf("could not unmarshal response: %v", resp)
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
	}

	return subscription, &response, nil
}

func (c Client) Delete(token *auth2.AccessToken, name string) (*types2.DeleteResponse, error) {
	req, err := c.httpClient.NewRequest(token, http.MethodDelete, fmt.Sprintf(c.config.DeleteURLFormat, name), nil)
	if err != nil {
		return nil, err
	}

	var response types2.DeleteResponse
	if resp, err := c.httpClient.Do(req, &response); err != nil {
		return nil, err
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
	}

	return &response, nil
}

func (c Client) TriggerHandshake(token *auth2.AccessToken, name string) (*types2.TriggerHandshake, error) {
	req, err := c.httpClient.NewRequest(token, http.MethodPost, fmt.Sprintf(c.config.HandshakeURLFormat, name), nil)
	if err != nil {
		return nil, err
	}

	var response types2.TriggerHandshake
	if resp, err := c.httpClient.Do(req, &response); err != nil {
		return nil, err
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
	}

	return &response, nil
}

func (c Client) UpdateState(token *auth2.AccessToken, name string, state types2.State) (*types2.UpdateStateResponse, error) {
	req, err := c.httpClient.NewRequest(token, http.MethodPut, fmt.Sprintf(c.config.UpdateStateURLFormat, name), state)
	if err != nil {
		return nil, err
	}

	var response types2.UpdateStateResponse
	if resp, err := c.httpClient.Do(req, &response); err != nil {
		return nil, err
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
	}

	return &response, nil
}
