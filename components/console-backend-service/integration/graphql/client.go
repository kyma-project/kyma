package graphql

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	timeout = 10 * time.Second
)

type Client struct {
	gqlClient *graphql.Client
	endpoint  string
	user      string
}

func New(endpoint, user string) (*Client, error) {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	gqlClient := graphql.NewClient(endpoint, graphql.WithHTTPClient(httpClient))

	client := &Client{
		gqlClient: gqlClient,
		endpoint:  endpoint,
		user:      user,
	}

	return client, nil
}

func (c *Client) DoQuery(q string, res interface{}) error {
	req := NewRequest(q)
	return c.Do(req, res)
}

func (c *Client) Do(req *Request, res interface{}) error {
	req.AddHeader("user", c.user)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := c.gqlClient.Run(ctx, req.req, res)
	return err
}

func (c *Client) Subscribe(req *Request) *Subscription {
	if req == nil {
		return errorSubscription(errors.New("invalid request"))
	}

	url := strings.Replace(c.endpoint, "http://", "ws://", -1)
	url = strings.Replace(url, "https://", "wss://", -1)

	connection, err := newWebsocket(url, req.req.Header)
	if err != nil {
		return errorSubscription(err)
	}

	err = connection.Handshake()
	if err != nil {
		return errorSubscription(err)
	}

	js, err := req.JSON()
	if err != nil {
		return errorSubscription(errors.Wrapf(err, "while converting request to json"))
	}

	err = connection.Start(js)
	if err != nil {
		return errorSubscription(err)
	}

	return newSubscription(connection)
}
