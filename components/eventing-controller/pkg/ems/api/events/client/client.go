package client

import (
	"fmt"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/auth"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/httpclient"
)

// compile time check
var _ Interface = Client{}

type Interface interface {
	Publish(cloudEvent cloudevents.Event, qos types.Qos) (*types.PublishResponse, error)
	Create(subscription *types.Subscription) (*types.CreateResponse, error)
	List() (*types.Subscriptions, *types.Response, error)
	Get(name string) (*types.Subscription, *types.Response, error)
	Delete(name string) (*types.DeleteResponse, error)
	TriggerHandshake(name string) (*types.TriggerHandshake, error)
	UpdateState(name string, state types.State) (*types.UpdateStateResponse, error)
}

type Client struct {
	config     *config.Config
	httpClient *httpclient.Client
}

func NewClient(config *config.Config, authenticator *auth.Authenticator) *Client {
	return &Client{
		config:     config,
		httpClient: authenticator.GetClient(),
	}
}

func (c Client) GetHttpClient() *httpclient.Client {
	return c.httpClient
}

func (c Client) Publish(event cloudevents.Event, qos types.Qos) (*types.PublishResponse, error) {
	req, err := c.httpClient.NewRequest(http.MethodPost, c.config.PublishURL, event)
	if err != nil {
		return nil, err
	}

	// set required headers
	req.Header.Set("qos", string(qos))

	var response types.PublishResponse
	if resp, responseBody, err := c.httpClient.Do(req, &response); err != nil {
		return nil, err
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
		if responseBody != nil {
			response.Message = response.Message + ";" + string(*responseBody)
		}
	}

	return &response, nil
}

func (c Client) Create(subscription *types.Subscription) (*types.CreateResponse, error) {
	req, err := c.httpClient.NewRequest(http.MethodPost, c.config.CreateURL, subscription)
	if err != nil {
		return nil, err
	}

	var response *types.CreateResponse
	if resp, responseBody, err := c.httpClient.Do(req, &response); err != nil {
		return nil, err
	} else if resp == nil {
		return nil, fmt.Errorf("could not unmarshal response: %v", resp)
	} else {
		if response == nil {
			response = &types.CreateResponse{}
		}
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
		if responseBody != nil {
			response.Message = response.Message + ";" + string(*responseBody)
		}
	}

	return response, nil
}

func (c Client) List() (*types.Subscriptions, *types.Response, error) {
	req, err := c.httpClient.NewRequest(http.MethodGet, c.config.ListURL, nil)
	if err != nil {
		return nil, nil, err
	}

	var subscriptions *types.Subscriptions
	var response types.Response
	if resp, responseBody, err := c.httpClient.Do(req, &subscriptions); err != nil {
		return nil, nil, err
	} else if resp == nil {
		return nil, nil, fmt.Errorf("could not unmarshal response: %v", resp)
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
		if subscriptions == nil && responseBody != nil {
			response.Message = response.Message + ";" + string(*responseBody)
		}
	}

	return subscriptions, &response, nil
}

func (c Client) Get(name string) (*types.Subscription, *types.Response, error) {
	req, err := c.httpClient.NewRequest(http.MethodGet, fmt.Sprintf(c.config.GetURLFormat, name), nil)
	if err != nil {
		return nil, nil, err
	}

	var subscription *types.Subscription
	var response types.Response
	if resp, responseBody, err := c.httpClient.Do(req, &subscription); err != nil {
		return nil, nil, err
	} else if resp == nil {
		return nil, nil, fmt.Errorf("could not unmarshal response: %v", resp)
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
		if subscription == nil && responseBody != nil {
			response.Message = response.Message + ";" + string(*responseBody)
		}
	}

	return subscription, &response, nil
}

func (c Client) Delete(name string) (*types.DeleteResponse, error) {
	req, err := c.httpClient.NewRequest(http.MethodDelete, fmt.Sprintf(c.config.DeleteURLFormat, name), nil)
	if err != nil {
		return nil, err
	}

	var response types.DeleteResponse
	if resp, responseBody, err := c.httpClient.Do(req, &response); err != nil {
		return nil, err
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
		if responseBody != nil {
			response.Message = response.Message + ";" + string(*responseBody)
		}
	}

	return &response, nil
}

func (c Client) TriggerHandshake(name string) (*types.TriggerHandshake, error) {
	req, err := c.httpClient.NewRequest(http.MethodPost, fmt.Sprintf(c.config.HandshakeURLFormat, name), nil)
	if err != nil {
		return nil, err
	}

	var response types.TriggerHandshake
	if resp, responseBody, err := c.httpClient.Do(req, &response); err != nil {
		return nil, err
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
		if responseBody != nil {
			response.Message = response.Message + ";" + string(*responseBody)
		}
	}

	return &response, nil
}

func (c Client) UpdateState(name string, state types.State) (*types.UpdateStateResponse, error) {
	req, err := c.httpClient.NewRequest(http.MethodPut, fmt.Sprintf(c.config.UpdateStateURLFormat, name), state)
	if err != nil {
		return nil, err
	}

	var response types.UpdateStateResponse
	if resp, responseBody, err := c.httpClient.Do(req, &response); err != nil {
		return nil, err
	} else {
		response.StatusCode = resp.StatusCode
		response.Message = resp.Status
		if responseBody != nil {
			response.Message = response.Message + ";" + string(*responseBody)
		}
	}

	return &response, nil
}
